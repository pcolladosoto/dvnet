package dvnet

import (
	"encoding/binary"
	"fmt"
	"net"
	"runtime"
	"strings"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type subnetAddresser struct {
	cidrBlock net.IPNet
	currentIP net.IP
}

var subnetAddressers = map[string]subnetAddresser{}

func newSubnetAddresser(subnetName string, subnetBlock net.IPNet) (subnetAddresser, error) {
	for _, addresser := range subnetAddressers {
		if addresser.cidrBlock.String() == subnetBlock.String() {
			return subnetAddresser{}, fmt.Errorf("subnet with CIDR %s has already been used up", &subnetBlock)
		}
	}
	if _, ok := subnetAddressers[subnetName]; ok {
		return subnetAddresser{}, fmt.Errorf("subnet %s has already been used up", &subnetBlock)
	}
	subnetAddressers[subnetName] = subnetAddresser{cidrBlock: subnetBlock, currentIP: make(net.IP, len(subnetBlock.IP))}
	copy(subnetAddressers[subnetName].currentIP, subnetBlock.IP)
	return subnetAddressers[subnetName], nil
}

func (sA subnetAddresser) nextIP() string {
	binary.BigEndian.PutUint32(sA.currentIP, binary.BigEndian.Uint32(sA.currentIP)+1)
	return sA.currentIP.String()
}

func (sA subnetAddresser) nextCIDR() string {
	return fmt.Sprintf("%s/%s", sA.nextIP(), strings.Split(sA.cidrBlock.String(), "/")[1])
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
