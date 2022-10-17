package dvnet

import (
	"fmt"

	"github.com/RyanCarrier/dijkstra"
)

func genGraph(net netDef) (*dijkstra.Graph, error) {
	netTopology := dijkstra.NewGraph()
	currentID := 0

	for _, subnetDef := range net.Subnets {
		for _, host := range subnetDef.Hosts {
			assignedID := netTopology.AddMappedVertex(host)
			log.debug("graphGen: currentID -> %d, assignedID -> %d", currentID, assignedID)
			if currentID != assignedID {
				return nil, fmt.Errorf("host %s has been defined more than once", host)
			}
			currentID++
		}
	}

	for routerName, routerDef := range net.Routers {
		assignedID := netTopology.AddMappedVertex(routerName)
		log.debug("graphGen: currentID -> %d, assignedID -> %d", currentID, assignedID)
		if currentID != assignedID {
			return nil, fmt.Errorf("router %s has been defined more than once", routerName)
		}
		currentID++
		for _, subnet := range routerDef.Subnets {
			subnetDef, ok := net.Subnets[subnet]
			if !ok {
				return nil, fmt.Errorf("router %s should be connected to subnet %s but it doesn't exist",
					routerName, subnet)
			}
			for _, host := range subnetDef.Hosts {
				netTopology.AddMappedArc(routerName, host, 1)
				netTopology.AddMappedArc(host, routerName, 1)
			}
		}
	}

	return netTopology, nil
}

func findRoutes(netGraph *dijkstra.Graph, netDefinition netDef, srcSubnet subnetDef) (map[string][]string, error) {
	shortestPaths := map[string][]string{}
	for dstSubnetName, dstSubnet := range netDefinition.Subnets {
		if srcSubnet.CIDRBlock.String() == dstSubnet.CIDRBlock.String() {
			continue
		}
		src := srcSubnet.Hosts[0]
		dst := dstSubnet.Hosts[0]
		shortestPath, err := netGraph.Shortest(
			netGraph.AddMappedVertex(src), netGraph.AddMappedVertex(dst))
		if err != nil {
			log.error("couldn't find shortest path from %s to %s: %v\n", src, dst, err)
			return nil, fmt.Errorf("couldn't find shortest path from %s to %s: %v", src, dst, err)
		}
		shortestPathMapped := []string{}
		for _, vertex := range shortestPath.Path {
			shortestPathMapped = append(
				shortestPathMapped, func(vID int) string { vMID, _ := netGraph.GetMapped(vID); return vMID }(vertex))
		}
		log.debug("shortest path from %s to %s: %v\n", src, dst, shortestPathMapped)
		shortestPaths[dstSubnetName] = shortestPathMapped[1 : len(shortestPathMapped)-1]
	}
	log.debug("discovered shortest paths from subnet %s: %v\n", srcSubnet.CIDRBlock.String(), shortestPaths)
	return shortestPaths, nil
}
