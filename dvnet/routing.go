package dvnet

import (
	"net"
	"runtime"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type graphRoute struct {
	destCIDR net.IPNet
	rawPath  []string
}

func routeContainer(subnetName string, route graphRoute, containerPID int) error {
	subnetAddresser := subnetAddressers[subnetName]

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

	gwIP := subnetAddresser.AssignedIPs[route.rawPath[0]]
	log.debug("adding route to %s through %s on container with PID %d\n", route.destCIDR.String(), gwIP.String(), containerPID)
	nlRoute := netlink.Route{Dst: &route.destCIDR, Gw: gwIP}

	if err := netlink.RouteAdd(&nlRoute); err != nil {
		netns.Set(origNS)
		return err
	}

	return netns.Set(origNS)
}

func addDefaultRoute(gwIP net.IP, containerPID int) error {
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

	log.debug("adding route to default through %s on container with PID %d\n", gwIP, containerPID)
	nlRoute := netlink.Route{
		Dst: func() *net.IPNet { _, ipNet, _ := net.ParseCIDR("0.0.0.0/0"); return ipNet }(),
		Gw:  gwIP}

	if err := netlink.RouteAdd(&nlRoute); err != nil {
		netns.Set(origNS)
		return err
	}

	return netns.Set(origNS)
}
