package dvnet

import (
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

func createVethPair(suffix string) (*netlink.Veth, netlink.Link, netlink.Link, error) {
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: bridgeEthPrefix + suffix},
		PeerName:  containerEthPrefix + suffix,
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
				Chain:  "POSTROUTING",
				Output: output,
			}
		}
	}
	return nil
}
