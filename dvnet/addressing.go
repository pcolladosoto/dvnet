package dvnet

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type subnetAddresser struct {
	cidrBlock   net.IPNet
	currentIP   net.IP
	AssignedIPs map[string]net.IP
}

func newSubnetAddresser(ns *NetworkState, subnetName string, subnetBlock net.IPNet) (subnetAddresser, error) {
	for _, addresser := range ns.Addressers {
		if addresser.cidrBlock.String() == subnetBlock.String() {
			return subnetAddresser{}, fmt.Errorf("subnet with CIDR %s has already been used up", &subnetBlock)
		}
	}
	if _, ok := ns.Addressers[subnetName]; ok {
		return subnetAddresser{}, fmt.Errorf("subnet %s has already been used up", &subnetBlock)
	}
	ns.Addressers[subnetName] = subnetAddresser{
		cidrBlock: subnetBlock, currentIP: make(net.IP, len(subnetBlock.IP)), AssignedIPs: map[string]net.IP{}}
	copy(ns.Addressers[subnetName].currentIP, subnetBlock.IP)
	return ns.Addressers[subnetName], nil
}

func (sA subnetAddresser) nextIP(hostName string) string {
	binary.BigEndian.PutUint32(sA.currentIP, binary.BigEndian.Uint32(sA.currentIP)+1)
	sA.AssignedIPs[hostName] = sA.currentIP
	return sA.currentIP.String()
}

func (sA subnetAddresser) nextCIDR(hostName string) string {
	return fmt.Sprintf("%s/%s", sA.nextIP(hostName), strings.Split(sA.cidrBlock.String(), "/")[1])
}

func addressContainer(cidr string, iface netlink.Link, containerPID int) error {
	netlinkCIDR, err := netlink.ParseAddr(cidr)
	if err != nil {
		return err
	}

	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save the current network namespace
	origNS, _ := netns.Get()
	defer origNS.Close()

	// Get the container's network namespace
	containerNS, err := netns.GetFromPid(containerPID)
	if err != nil {
		return err
	}
	defer containerNS.Close()

	netns.Set(containerNS)

	if err := netlink.AddrAdd(iface, netlinkCIDR); err != nil {
		return err
	}

	return netns.Set(origNS)
}

func dumpAddressAssignments(ns *NetworkState, path string) error {
	addressers, err := json.Marshal(ns.Addressers)
	if err != nil {
		return err
	}
	return os.WriteFile(path, addressers, 0644)
}
