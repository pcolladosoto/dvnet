package dvnet

import (
	"fmt"
	"strings"
)

func createSubnet(netState *NetworkState, subnetName string, def subnetDef) error {
	log.debug("creating subnet %s\n", subnetName)

	if _, ok := netState.Subnets[subnetName]; ok {
		return fmt.Errorf("subnet %s has already been defined", subnetName)
	}

	subnetAddresser, err := newSubnetAddresser(subnetName, def.CIDRBlock)
	if err != nil {
		return err
	}

	subnetBridge, err := createBridge(subnetName)
	if err != nil {
		return fmt.Errorf("couldn't create bridge %s: %w", subnetName, err)
	}

	netState.Subnets[subnetName] = SubnetResources{Bridge: subnetBridge, Containers: map[string]containerInfo{}}

	for _, host := range def.Hosts {
		containerID, containerPID, err := runContainer("pcollado/dhost", host)
		if err != nil {
			return fmt.Errorf("couldn't start container for host %s: %w", host, err)
		}
		if _, ok := netState.Subnets[subnetName].Containers[host]; ok {
			return fmt.Errorf("host %s has been defined more than once", host)
		}
		netState.Subnets[subnetName].Containers[host] = containerInfo{ID: containerID, PID: containerPID}
		veth, bridgeEnd, containerEnd, err := createVethPair(strings.ToLower(host))
		if err != nil {
			log.error("couldn't create veth %s-%s: %v\n", subnetBridge.Name, host, err)
			return err
		}

		log.debug("connecting %s to %s\n", veth.Name, subnetBridge.Name)
		if err := connectToBridge(bridgeEnd, subnetBridge); err != nil {
			log.error("couldn't connect %s to %s: %v\n", veth.Name, subnetBridge.Name, err)
			return err
		}

		log.debug("connecting %s to %s\n", veth.PeerName, host)
		if err := connectToContainer(containerEnd, containerPID); err != nil {
			log.error("couldn't connect %s to %s: %v\n", veth.PeerName, host, err)
			return err
		}

		assignedCIDR := subnetAddresser.nextCIDR()
		log.debug("assigning %s to %s on %s\n", assignedCIDR, veth.PeerName, host)
		if err := addressContainer(assignedCIDR, containerEnd, containerPID); err != nil {
			log.error("couldn't address %s to %s on %s: %v\n", assignedCIDR, veth.PeerName, host, err)
			return err
		}
	}
	return nil
}

func createRouter(netState *NetworkState, routerName string, def routerDef) error {
	containerID, containerPID, err := runContainer("pcollado/drouter", routerName)
	if err != nil {
		return fmt.Errorf("couldn't start container for router %s: %w", routerName, err)
	}
	if _, ok := netState.Routers[routerName]; ok {
		return fmt.Errorf("router %s has been defined more than once", routerName)
	}
	netState.Routers[routerName] = containerInfo{ID: containerID, PID: containerPID}

	for _, subnetName := range def.Subnets {
		subnetAddresser, okAddresses := subnetAddressers[subnetName]
		subnetResources, okResources := netState.Subnets[subnetName]
		if !okAddresses || !okResources {
			return fmt.Errorf("subnet %s should exist at this point", subnetName)
		}
		veth, bridgeEnd, containerEnd, err := createVethPair(strings.ToLower(fmt.Sprintf("%s-%s", routerName, subnetName)))
		if err != nil {
			log.error("couldn't create veth %s-%s: %v\n", subnetResources.Bridge.Name, routerName, err)
			return err
		}

		log.debug("connecting %s to %s\n", veth.Name, subnetResources.Bridge.Name)
		if err := connectToBridge(bridgeEnd, subnetResources.Bridge); err != nil {
			log.error("couldn't connect %s to %s: %v\n", veth.Name, subnetResources.Bridge.Name, err)
			return err
		}

		log.debug("connecting %s to %s\n", veth.PeerName, routerName)
		if err := connectToContainer(containerEnd, containerPID); err != nil {
			log.error("couldn't connect %s to %s: %v\n", veth.PeerName, routerName, err)
			return err
		}

		assignedCIDR := subnetAddresser.nextCIDR()
		log.debug("assigning %s to %s on %s\n", assignedCIDR, veth.PeerName, routerName)
		if err := addressContainer(assignedCIDR, containerEnd, containerPID); err != nil {
			log.error("couldn't address %s to %s on %s: %v\n", assignedCIDR, veth.PeerName, routerName, err)
			return err
		}
	}

	return nil
}
