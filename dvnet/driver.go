package dvnet

import (
	"fmt"
	"strings"

	"github.com/docker/go-plugins-helpers/network"
	"github.com/vishvananda/netlink"

	"github.com/docker/docker/client"
)

const (
	scope                     string = "local"
	connectivityScope         string = "global"
	defaultRoute              string = "0.0.0.0/0"
	bridgePrefix              string = "dvn-"
	bridgeEthPrefix           string = "bth-"
	containerEthPrefix        string = "eth"
	defaultHopBridgePrefix    string = "hth-"
	defaultHopContainerPrefix string = "dth-"

	genericOptPrefix string = "com.docker.network.generic"

	mtuOption        string = "net.dvnet.mtu"
	modeOption       string = "net.dvnet.mode"
	bridgeNameOption string = "net.dvnet.name"
	netDefOption     string = "net.dvnet.def"

	modeNAT  string = "nat"
	modeFlat string = "flat"
)

var (
	defaultMTU         uint   = 1500
	defaultMode        string = modeNAT
	defaultNetDefPath  string = "/tmp/netDef.json"
	defaultGatewayName string = "dvhop"
	defaultBridgeName  string = ""
	defaultGateway     string = ""
	defaultMask        string = ""
)

type globalOpts struct {
	netDefPath string
	bridgeName string
	gateway    string
	mask       string
}

type Driver struct {
	networks map[string]*NetworkState
}

type SubnetResources struct {
	Bridge     *netlink.Bridge
	Containers map[string]containerInfo
}

type NetworkState struct {
	BridgeName      string
	BridgeInst      *netlink.Bridge
	HopCIDR         string
	MTU             uint
	Mode            string
	Gateway         string
	GatewayMask     string
	PreviousSysctls map[string]string
	Subnets         map[string]SubnetResources
	Addressers      map[string]subnetAddresser
	Routers         map[string]containerInfo
}

// GetCapabilities tells the Docker daemon the reach of the
// resource allocations this driver can perform. Check
// https://github.com/moby/libnetwork/blob/master/docs/remote.md#set-capability
// for more info on the topic.
func (d Driver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	log.debug("GetCapabilities request\n")
	return &network.CapabilitiesResponse{Scope: scope, ConnectivityScope: connectivityScope}, nil
}

// Check https://forums.docker.com/t/how-to-disable-ipam/81560 on how to disable IPAM
func (d Driver) CreateNetwork(req *network.CreateNetworkRequest) error {
	log.debug("CreateNetwork() request: %+v\n", req)

	prevSysctls, err := systemSetup()
	if err != nil {
		log.error("couldn't configure the host system: %v\n", err)
		return d.failWithCleanup(req.NetworkID, err)
	}

	var netOpts globalOpts

	parseOptions(req, &netOpts)
	log.debug("configured options: %+v\n", netOpts)

	ns := &NetworkState{
		BridgeName: netOpts.bridgeName,
		// BridgeInst:      bridgeInst,
		MTU:             defaultMTU,
		Mode:            defaultMode,
		Gateway:         netOpts.gateway,
		GatewayMask:     netOpts.mask,
		PreviousSysctls: prevSysctls,
		HopCIDR:         "",
		Subnets:         map[string]SubnetResources{},
		Addressers:      map[string]subnetAddresser{},
		Routers:         map[string]containerInfo{},
	}

	d.networks[req.NetworkID] = ns

	netDefinition, err := loadDef(netOpts.netDefPath)
	if err != nil {
		log.error("couldn't load the network definition: %v\n", err)
		return d.failWithCleanup(req.NetworkID, err)
	}

	log.debug("loaded network definition: %+v\n", netDefinition)

	netGraph, err := genGraph(netDefinition)
	if err != nil {
		return d.failWithCleanup(req.NetworkID, err)
	}

	netGrapPath := fmt.Sprintf("%s.netg", strings.Split(netOpts.netDefPath, ".")[0])
	log.debug("exported network graph to %s\n", netGrapPath)
	netGraph.ExportToFile(netGrapPath)

	for subnetName, subnetDef := range netDefinition.Subnets {
		if err := createSubnet(ns, subnetName, subnetDef); err != nil {
			return d.failWithCleanup(req.NetworkID, err)
		}
	}

	for routerName, def := range netDefinition.Routers {
		if err := createRouter(ns, routerName, def); err != nil {
			return d.failWithCleanup(req.NetworkID, err)
		}
	}

	if netDefinition.AutomaticRouting {
		for subnetName, subnetDef := range netDefinition.Subnets {
			routes, err := findSubnetRoutes(netGraph, netDefinition, subnetDef)
			if err != nil {
				return d.failWithCleanup(req.NetworkID, err)
			}
			for host := range subnetDef.Hosts {
				for _, route := range routes {
					if err := routeContainer(ns, subnetName, route, ns.Subnets[subnetName].Containers[host].PID); err != nil {
						return d.failWithCleanup(req.NetworkID, err)
					}
				}
			}
		}
	}

	if netDefinition.OutboundAccess.Enabled {
		if err := confOutboundAccess(ns, defaultGatewayName, netDefinition.OutboundAccess.HopCIDR); err != nil {
			return d.failWithCleanup(req.NetworkID, err)
		}
	}

	ipAddressesPath := fmt.Sprintf("%s.ipaddr", strings.Split(netOpts.netDefPath, ".")[0])
	if err := dumpAddressAssignments(ns, ipAddressesPath); err != nil {
		log.error("couldn't dump the assigned IPv4 addresses: %v\n", err)
	}
	log.debug("exported assigned addresses to to %s\n", ipAddressesPath)

	log.debug("built network state: %#v\n", *ns)

	return nil
}

func (d Driver) failWithCleanup(networkID string, err error) error {
	d.DeleteNetwork(&network.DeleteNetworkRequest{NetworkID: networkID})
	return err
}

func (d Driver) DeleteNetwork(req *network.DeleteNetworkRequest) error {
	log.debug("DeleteNetwork() request: %+v\n", req)
	ns, ok := d.networks[req.NetworkID]
	if !ok {
		log.warn("trying to remove a network we are unaware of: %s\n", req.NetworkID)
		return fmt.Errorf("the network driver is unaware of this network")
	}

	log.debug("trying to delete network whose state is %#v\n", *ns)

	if err := restoreSysctls(ns.PreviousSysctls); err != nil {
		log.error("%v\n", err)
		return err
	}

	if err := restoreNAT(ns.HopCIDR); err != nil {
		log.error("%v\n", err)
		return err
	}

	if err := restoreForwarding(bridgePrefix + strings.ToLower(defaultGatewayName)); err != nil {
		log.error("%v\n", err)
		return err
	}

	for _, subnet := range ns.Subnets {
		removeBridge(subnet.Bridge)

		for _, containerInfo := range subnet.Containers {
			log.debug("removing container with ID %s\n", containerInfo.ID)
			if err := removeContainer(containerInfo.ID); err != nil {
				log.error("couldn't remove container with ID %s: %v\n", containerInfo.ID, err)
			}
		}
	}

	for _, containerInfo := range ns.Routers {
		log.debug("removing container with ID %s\n", containerInfo.ID)
		if err := removeContainer(containerInfo.ID); err != nil {
			log.error("couldn't remove container with ID %s: %v\n", containerInfo.ID, err)
		}
	}

	d.networks[req.NetworkID] = nil

	return nil
}

func (d Driver) AllocateNetwork(req *network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	log.debug("AllocateNetwork() request: %+v\n", req)
	return nil, nil
}

func (d Driver) FreeNetwork(req *network.FreeNetworkRequest) error {
	log.debug("FreeNetwork() request: %+v\n", req)
	return nil
}

// Check https://github.com/moby/libnetwork/blob/master/docs/remote.md#create-endpoint
func (d Driver) CreateEndpoint(req *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	log.debug("CreateEndpoint() request received")
	log.debug("\tNetwork ID: %s Endpoint ID: %s\n", req.NetworkID[:5], req.EndpointID[:5])
	log.debug("\tInterface: Address: %s; MAC: %q\n", req.Interface.Address, req.Interface.MacAddress)

	return nil, nil
}

func (d Driver) DeleteEndpoint(req *network.DeleteEndpointRequest) error {
	log.debug("DeleteEndpoint() request: %+v\n", req)
	return nil
}

func (d Driver) EndpointInfo(req *network.InfoRequest) (*network.InfoResponse, error) {
	log.debug("EndpointInfo() request: %+v\n", req)
	return nil, nil
}

func (d Driver) Join(req *network.JoinRequest) (*network.JoinResponse, error) {
	log.debug("Join() request: %+v\n", req)
	return nil, nil
}

func (d Driver) Leave(req *network.LeaveRequest) error {
	log.debug("Leave() request: %+v\n", req)
	return nil
}

func (d Driver) DiscoverNew(req *network.DiscoveryNotification) error {
	log.debug("DiscoverNew() request: %+v\n", req)
	return nil
}

func (d Driver) DiscoverDelete(req *network.DiscoveryNotification) error {
	log.debug("DiscoverDelete() request: %+v\n", req)
	return nil
}

func (d Driver) ProgramExternalConnectivity(req *network.ProgramExternalConnectivityRequest) error {
	log.debug("ProgramExternalConnectivity() request: %+v\n", req)
	return nil
}

func (d Driver) RevokeExternalConnectivity(req *network.RevokeExternalConnectivityRequest) error {
	log.debug("ProgramExternalConnectivity() request: %+v\n", req)
	return nil
}

func GetHandler() *network.Handler {
	d := Driver{networks: make(map[string]*NetworkState)}
	h := network.NewHandler(d)

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.error("couldn't get a docker client: %v\n", err)
	}
	dockerCli = cli

	return h
}
