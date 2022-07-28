package dvnet

import (
	"encoding/json"
	"io/ioutil"

	"github.com/go-playground/validator/v10"
)

type netDef struct {
	Name            string               `json:"name" validate:"required"`
	OutboundAccess  bool                 `json:"outbound_access"`
	UpdateHostsFile bool                 `json:"update_hosts"`
	Subnets         map[string]subnetDef `json:"subnets" validate:"required"`
	Routers         map[string]routerDef `json:"routers" validate:"required"`
}

type subnetDef struct {
	CIDRBlock string   `json:"cidr" validate:"required,cidr4"`
	Hosts     []string `json:"hosts" validate:"required,unique,dive,required"`
}

type routerDef struct {
	Subnets []string  `json:"subnets" validate:"required,unique,dive,required"`
	FWRules fwRuleDef `json:"fw_rules"`
}

type fwRuleDef struct {
	Policy string          `json:"policy"`
	Accept [][]fwTargetDef `json:"accept"`
	Drop   [][]fwTargetDef `json:"drop"`
}

type fwTargetDef interface{}

func loadDef(fPath string) (netDef, error) {
	rawDef, err := ioutil.ReadFile(fPath)
	if err != nil {
		return netDef{}, err
	}

	return parseDef(rawDef)
}

func parseDef(rawDef []byte) (netDef, error) {
	var def netDef

	if err := json.Unmarshal(rawDef, &def); err != nil {
		return netDef{}, err
	}
	return def, nil
}

func validateDef(def netDef) error {
	validate := validator.New()
	return validate.Struct(def)
}
