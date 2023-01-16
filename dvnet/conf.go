package dvnet

import (
	"encoding/json"
	"net"
	"os"

	"github.com/go-playground/validator/v10"
)

type rawNetDef struct {
	Name             string                  `json:"name" validate:"required"`
	OutboundAccess   RawOutboundAccessDef    `json:"outbound_access"`
	UpdateHostsFile  bool                    `json:"update_hosts"`
	AutomaticRouting bool                    `json:"automatic_routing"`
	Subnets          map[string]rawSubnetDef `json:"subnets" validate:"required"`
	Routers          map[string]routerDef    `json:"routers" validate:"required"`
}

type RawOutboundAccessDef struct {
	Enabled bool   `json:"enabled"`
	HopCIDR string `json:"cidr"`
}

type OutboundAccessDef struct {
	Enabled bool      `json:"enabled"`
	HopCIDR net.IPNet `json:"cidr"`
}

type netDef struct {
	Name             string               `json:"name" validate:"required"`
	OutboundAccess   OutboundAccessDef    `json:"outbound_access"`
	UpdateHostsFile  bool                 `json:"update_hosts"`
	AutomaticRouting bool                 `json:"automatic_routing"`
	Subnets          map[string]subnetDef `json:"subnets" validate:"required"`
	Routers          map[string]routerDef `json:"routers" validate:"required"`
}

type rawSubnetDef struct {
	CIDRBlock string             `json:"cidr" validate:"required,cidr4"`
	Hosts     map[string]HostDef `json:"hosts" validate:"required,unique,dive,required"`
}

type HostDef struct {
	Image string `json:"image"`
}

type subnetDef struct {
	CIDRBlock net.IPNet          `json:"cidr" validate:"required,cidr4"`
	Hosts     map[string]HostDef `json:"hosts" validate:"required,unique,dive,required"`
}

type routerDef struct {
	Subnets []string  `json:"subnets" validate:"required,unique,dive,required"`
	FWRules fwRuleDef `json:"fw_rules"`
	Image   string    `json:"image"`
}

type fwRuleDef struct {
	Policy string          `json:"policy"`
	Accept [][]fwTargetDef `json:"accept"`
	Drop   [][]fwTargetDef `json:"drop"`
}

type fwTargetDef interface{}

func cidrParserWrapper(rawCIDR string) net.IPNet {
	_, netAddr, _ := net.ParseCIDR(rawCIDR)
	return *netAddr
}

func loadDef(fPath string) (netDef, error) {
	rawDef, err := os.ReadFile(fPath)
	if err != nil {
		return netDef{}, err
	}
	return parseDef(rawDef)
}

func parseDef(rawDef []byte) (netDef, error) {
	var rDef rawNetDef

	if err := json.Unmarshal(rawDef, &rDef); err != nil {
		return netDef{}, err
	}

	parsedSubnets := map[string]subnetDef{}
	for subnetName, rawSubnet := range rDef.Subnets {
		parsedSubnets[subnetName] = subnetDef{
			CIDRBlock: cidrParserWrapper(rawSubnet.CIDRBlock),
			Hosts:     rawSubnet.Hosts,
		}
	}

	parsedOutboundAccess := OutboundAccessDef{
		Enabled: rDef.OutboundAccess.Enabled,
		HopCIDR: cidrParserWrapper(rDef.OutboundAccess.HopCIDR),
	}

	def := netDef{
		Name:             rDef.Name,
		OutboundAccess:   parsedOutboundAccess,
		UpdateHostsFile:  rDef.UpdateHostsFile,
		AutomaticRouting: rDef.AutomaticRouting,
		Subnets:          parsedSubnets,
		Routers:          rDef.Routers,
	}

	return def, validateDef(def)
}

func validateDef(def netDef) error {
	validate := validator.New()

	if err := validate.Struct(def); err != nil {
		return err
	}
	return nil
}
