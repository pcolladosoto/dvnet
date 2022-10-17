package dvnet

import (
	"errors"
	"fmt"
	"os"
	"strings"

	sysctl "github.com/lorenzosaino/go-sysctl"

	"github.com/docker/go-plugins-helpers/network"
)

var configurableSysctls map[string]string = map[string]string{
	"net.ipv4.ip_forward":                "1",
	"net.bridge.bridge-nf-call-iptables": "0",
}

func parseOptions(req *network.CreateNetworkRequest, netOpts *globalOpts) {
	if req.Options != nil {
		genericOpts, ok := req.Options[genericOptPrefix].(map[string]interface{})
		if !ok {
			log.warn("couldn't find generic options\n")
			return
		}

		if brdName, ok := genericOpts[bridgeNameOption].(string); ok {
			netOpts.bridgeName = brdName
		} else {
			netOpts.bridgeName = bridgePrefix + truncateID(req.NetworkID)
		}
		if confPath, ok := genericOpts[netDefOption].(string); ok {
			netOpts.netDefPath = confPath
		} else {
			netOpts.netDefPath = defaultNetDefPath
		}
	}

	gateway, mask, err := getGatewayIP(req)
	if err != nil {
		log.warn("couldn't get the gateway information: %v\n", err)
		netOpts.gateway = defaultGateway
		netOpts.mask = defaultMask
	}
	netOpts.gateway = gateway
	netOpts.mask = mask
}

func getGatewayIP(req *network.CreateNetworkRequest) (string, string, error) {
	var gatewayIP string

	if len(req.IPv4Data) > 0 {
		if req.IPv4Data[0] != nil {
			if req.IPv4Data[0].Gateway != "" {
				gatewayIP = req.IPv4Data[0].Gateway
			}
		}
	}

	if gatewayIP == "" {
		return "", "", fmt.Errorf("no gateway IP found")
	}

	parts := strings.Split(gatewayIP, "/")
	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("cannot split gateway IP address: %s", gatewayIP)
	}
	return parts[0], parts[1], nil
}

// truncateID truncates id's length to 5 characters
// to avoid overflowing the strict interface name
// length constraints of the Linux kernel.
func truncateID(id string) string {
	return id[:5]
}

func systemSetup() (map[string]string, error) {
	prevSysctls := map[string]string{}

	for sctl := range configurableSysctls {
		confdSysctl, err := sysctl.Get(sctl)
		if err != nil {
			log.warn("couldn't retrieve sysctl value for %s...\n", sctl)
			continue
		}
		prevSysctls[sctl] = confdSysctl
	}

	log.debug("configuring sysctl net.ipv4.ip_forward = 1\n")
	if err := sysctl.Set("net.ipv4.ip_forward", "1"); err != nil {
		return nil, fmt.Errorf("couldn't set up IPv4 forwarding on the host")
	}

	log.debug("configuring sysctl net.bridge.bridge-nf-call-iptables = 0\n")
	if err := sysctl.Set("net.bridge.bridge-nf-call-iptables", "0"); err != nil {
		log.warn("couldn't set up IPv4 forwarding on the host. Is the br_netfilter module loaded?")
	}

	if _, err := os.Stat("/var/run/netns"); os.IsNotExist(err) {
		log.debug("creating /var/run/netns\n")
		err := os.Mkdir("/var/run/netns", 0755)
		if err != nil {
			log.warn("couldn't create /var/run/netns...\n")
		}
	} else {
		log.debug("directory /var/run/netns already existed\n")
	}

	return prevSysctls, nil
}

func restoreSysctls(prevSysctls map[string]string) error {
	globalErr := errors.New("error restoring previous sysctls")
	retErr := false
	for sctl, val := range prevSysctls {
		if err := sysctl.Set(sctl, val); err != nil {
			retErr = true
			globalErr = fmt.Errorf("%w; couldn't restore %s to %s", globalErr, sctl, val)
		}
	}
	if retErr {
		return globalErr
	}
	return nil
}
