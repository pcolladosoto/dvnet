package dvnet

import (
	"fmt"
	"net"
	"runtime"
	"strings"

	"github.com/docker/libnetwork/iptables"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func createBridge(name string) (*netlink.Bridge, error) {
	bridge := &netlink.Bridge{LinkAttrs: netlink.LinkAttrs{Name: bridgePrefix + strings.ToLower(name)}}
	if err := netlink.LinkAdd(bridge); err != nil {
		return nil, err
	}
	return bridge, netlink.LinkSetUp(bridge)
}

func connectToBridge(vethEnd netlink.Link, bridge *netlink.Bridge) error {
	if err := netlink.LinkSetMaster(vethEnd, bridge); err != nil {
		return err
	}
	return netlink.LinkSetUp(vethEnd)
}

func connectToContainer(vethEnd netlink.Link, containerPID int) error {
	if err := netlink.LinkSetNsPid(vethEnd, containerPID); err != nil {
		return err
	}
	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origNS, _ := netns.Get()
	defer origNS.Close()

	containerNS, err := netns.GetFromPid(containerPID)
	if err != nil {
		return err
	}
	defer containerNS.Close()

	netns.Set(containerNS)

	if err := netlink.LinkSetUp(vethEnd); err != nil {
		return err
	}

	return netns.Set(origNS)
}

func createVethPair(bridgePrefix, containerPrefix, suffix string) (*netlink.Veth, netlink.Link, netlink.Link, error) {
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: bridgePrefix + suffix},
		PeerName:  containerPrefix + suffix,
	}

	if err := netlink.LinkAdd(veth); err != nil {
		return nil, nil, nil, err
	}

	bridgeEnd, err := netlink.LinkByName(veth.Name)
	if err != nil {
		return nil, nil, nil, err
	}
	containerEnd, err := netlink.LinkByName(veth.PeerName)
	if err != nil {
		return nil, nil, nil, err
	}

	return veth, bridgeEnd, containerEnd, nil
}

func addressBridge(cidr string, bridge *netlink.Bridge) error {
	nlAddr, err := netlink.ParseAddr(cidr)
	if err != nil {
		log.warn("could't parse CIDR mask %s\n", cidr)
		return err
	}
	return netlink.AddrAdd(bridge, nlAddr)
}

func removeBridge(bridge *netlink.Bridge) error {
	return netlink.LinkDel(bridge)
}

// natOut allows the provided CIDR to be NATted
// out of the machine so that it can reach
// external networks.
func natOut(cidr string) error {
	masquerade := []string{
		"POSTROUTING", "-t", "nat",
		"-s", cidr,
		"-j", "MASQUERADE",
	}
	if _, err := iptables.Raw(
		append([]string{"-C"}, masquerade...)...,
	); err != nil {
		incl := append([]string{"-I"}, masquerade...)
		if output, err := iptables.Raw(incl...); err != nil {
			return err
		} else if len(output) > 0 {
			return &iptables.ChainError{
				Chain:  masquerade[0],
				Output: output,
			}
		}
	}
	return nil
}

func restoreNAT(cidr string) error {
	if cidr == "" {
		return nil
	}
	masquerade := []string{
		"POSTROUTING", "-t", "nat",
		"-s", cidr,
		"-j", "MASQUERADE",
	}
	if _, err := iptables.Raw(
		append([]string{"-C"}, masquerade...)...,
	); err == nil {
		incl := append([]string{"-D"}, masquerade...)
		if output, err := iptables.Raw(incl...); err != nil {
			return err
		} else if len(output) > 0 {
			return &iptables.ChainError{
				Chain:  masquerade[0],
				Output: output,
			}
		}
	}
	return nil
}

func enableForwarding(hopBridgeName string) error {
	forwardingRules := [][]string{
		{"FORWARD", "-i", hopBridgeName, "-j", "ACCEPT"},
		{"FORWARD", "-o", hopBridgeName, "-j", "ACCEPT"},
	}
	for _, rule := range forwardingRules {
		if _, err := iptables.Raw(
			append([]string{"-C"}, rule...)...,
		); err != nil {
			incl := append([]string{"-I"}, rule...)
			if output, err := iptables.Raw(incl...); err != nil {
				return err
			} else if len(output) > 0 {
				return &iptables.ChainError{
					Chain:  rule[0],
					Output: output,
				}
			}
		}
	}
	return nil
}

func restoreForwarding(hopBridgeName string) error {
	if hopBridgeName == "" {
		return nil
	}
	forwardingRules := [][]string{
		{"FORWARD", "-i", hopBridgeName, "-j", "ACCEPT"},
		{"FORWARD", "-o", hopBridgeName, "-j", "ACCEPT"},
	}
	for _, rule := range forwardingRules {
		if _, err := iptables.Raw(
			append([]string{"-C"}, rule...)...,
		); err == nil {
			incl := append([]string{"-D"}, rule...)
			if output, err := iptables.Raw(incl...); err != nil {
				return err
			} else if len(output) > 0 {
				return &iptables.ChainError{
					Chain:  rule[0],
					Output: output,
				}
			}
		}
	}
	return nil
}

func confOutboundAccess(netState *NetworkState, hopBridgeName string, hopBridgeCIDR net.IPNet) error {
	hopBrd, err := createBridge(hopBridgeName)
	if err != nil {
		return err
	}

	netState.Subnets["outboundSubnet"] = SubnetResources{Bridge: hopBrd, Containers: map[string]containerInfo{}}
	subnetAddresser, err := newSubnetAddresser(netState, "outboundSubnet", hopBridgeCIDR)
	if err != nil {
		return err
	}
	assignedHopBrdCIDR := subnetAddresser.nextCIDR(hopBridgeName)
	assignedHopBrdIP := strings.Split(assignedHopBrdCIDR, "/")[0]
	if err := addressBridge(assignedHopBrdCIDR, hopBrd); err != nil {
		return err
	}

	if err := natOut(hopBridgeCIDR.String()); err != nil {
		return err
	}
	netState.HopCIDR = hopBridgeCIDR.String()

	if err := enableForwarding(bridgePrefix + strings.ToLower(hopBridgeName)); err != nil {
		return err
	}

	for _, subnetResrc := range netState.Subnets {
		for containerName, containerInfo := range subnetResrc.Containers {
			veth, bridgeEnd, containerEnd, err := createVethPair(
				defaultHopBridgePrefix, defaultHopContainerPrefix,
				strings.ToLower(containerName))
			if err != nil {
				log.error("couldn't create veth %s-%s: %v\n", hopBrd.Name, containerName, err)
				return err
			}

			log.debug("connecting %s to %s\n", veth.Name, hopBrd.Name)
			if err := connectToBridge(bridgeEnd, hopBrd); err != nil {
				log.error("couldn't connect %s to %s: %v\n", veth.Name, hopBrd.Name, err)
				return err
			}

			log.debug("connecting %s to %s\n", veth.PeerName, containerName)
			if err := connectToContainer(containerEnd, containerInfo.PID); err != nil {
				log.error("couldn't connect %s to %s: %v\n", veth.PeerName, containerName, err)
				return err
			}

			assignedCIDR := subnetAddresser.nextCIDR(containerName)
			log.debug("assigning %s to %s on %s\n", assignedCIDR, veth.PeerName, containerName)
			if err := addressContainer(assignedCIDR, containerEnd, containerInfo.PID); err != nil {
				log.error("couldn't address %s to %s on %s: %v\n", assignedCIDR, veth.PeerName, containerName, err)
				return err
			}
			addDefaultRoute(net.ParseIP(assignedHopBrdIP), containerInfo.PID)
		}
	}

	for routerName, routerInfo := range netState.Routers {
		veth, bridgeEnd, containerEnd, err := createVethPair(
			defaultHopBridgePrefix, defaultHopContainerPrefix,
			strings.ToLower(fmt.Sprintf("%s-%s", routerName, "ob")))
		if err != nil {
			log.error("couldn't create veth %s-%s: %v\n", "ob", routerName, err)
			return err
		}

		log.debug("connecting %s to %s\n", veth.Name, "ob")
		if err := connectToBridge(bridgeEnd, hopBrd); err != nil {
			log.error("couldn't connect %s to %s: %v\n", veth.Name, "ob", err)
			return err
		}

		log.debug("connecting %s to %s\n", veth.PeerName, routerName)
		if err := connectToContainer(containerEnd, routerInfo.PID); err != nil {
			log.error("couldn't connect %s to %s: %v\n", veth.PeerName, routerName, err)
			return err
		}

		assignedCIDR := subnetAddresser.nextCIDR(routerName)
		log.debug("assigning %s to %s on %s\n", assignedCIDR, veth.PeerName, routerName)
		if err := addressContainer(assignedCIDR, containerEnd, routerInfo.PID); err != nil {
			log.error("couldn't address %s to %s on %s: %v\n", assignedCIDR, veth.PeerName, routerName, err)
			return err
		}

		addDefaultRoute(net.ParseIP(assignedHopBrdIP), routerInfo.PID)
	}

	return nil
}
