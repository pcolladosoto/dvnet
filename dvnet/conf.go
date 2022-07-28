package dvnet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/go-playground/validator/v10"
	"github.com/xeipuuv/gojsonschema"
)

var validationSchema = `{
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "$id": "file:docker_virt_net/net.schema",
    "title": "NetSchema",
    "description": "Rules virtual network definitions should adhere to",
    "type": "object",
    "properties": {
        "name": {
            "description": "The network's name. Used only for displaying purposes.",
            "type": "string"
        },
        "outbound_access": {
            "description": "Whether to allow network nodes to access the internet",
            "type": "boolean"
        },
        "update_hosts": {
            "description": "Whether to update the /etc/hosts file at each node",
            "type": "boolean"
        },
        "host_image": {
            "description": "Image to be run by host containers",
            "type": "string"
        },
        "router_image": {
            "description": "Image to be run by router containers",
            "type": "string"
        },
        "subnets": {
            "description": "Collection of subnets the network is composed of",
            "type": "object",
            "minProperties": 1,
            "patternProperties": {
                "^.+$": {
                    "description": "Contents of a subnet",
                    "type": "object",
                    "properties": {
                        "cidr": {
                            "description": "Subnet address as a CIDR block (A.B.C.D/X)",
                            "type": "string"
                        },
                        "hosts": {
                            "description": "List of hosts belonging to this subnet",
                            "type": "array",
                            "prefixItems": {"type": "string"},
                            "uniqueItems": true
                        }
                    },
                    "required": ["cidr", "hosts"]
                }
            }
        },
        "routers": {
            "description": "Set of routers belonging to the network",
            "type": "object",
            "patternProperties": {
                "^.+$": {
                    "description": "Router configuration",
                    "type": "object",
                    "properties": {
                        "fw_rules": {
                            "description": "Firewall rules to be applied to this router",
                            "type": "object",
                            "properties": {
                                "POLICY": {
                                    "description": "Defautl policy for the FORWARD chain",
                                    "type": "string",
                                    "enum": ["ACCEPT", "DROP"]
                                },
                                "ACCEPT": {
                                    "description": "Rules for the FORWARD chain describing packets to be ACCEPTed",
                                    "type": "array",
                                    "prefixItems": {
                                        "type": "array",
                                        "prefixItems": [
                                            {"type": "string"},
                                            {"type": "string"},
                                            {"type": "boolean"}
                                        ],
                                        "items": false
                                    },
                                    "uniqueItems": true
                                },
                                "DROP": {
                                    "description": "Rules for the FORWARD chain describing packets to be DROPped",
                                    "type": "array",
                                    "prefixItems": {
                                        "type": "array",
                                        "prefixItems": [
                                            {"type": "string"},
                                            {"type": "string"},
                                            {"type": "boolean"}
                                        ],
                                        "items": false
                                    },
                                    "uniqueItems": true
                                }
                            },
                            "oneOf": [
                                {"required": ["POLICY", "ACCEPT", "DROP"]},
                                {"maxProperties": 0}
                            ]
                        },
                        "subnets": {
                            "description": "Subnets this router attaches to",
                            "type": "array",
                            "prefixItems": {"type": "string"},
                            "uniqueItems": true
                        }
                    },
                    "required": ["fw_rules", "subnets"]
                }
            }
        }
    },
    "required": ["name", "subnets", "routers"]
}
`

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

func validateDefFoo(def netDef) error {
	validate := validator.New()
	return validate.Struct(def)
}

func validateDef(netDef string) error {
	schemaLoader := gojsonschema.NewStringLoader(validationSchema)
	netDefLoader := gojsonschema.NewStringLoader(netDef)

	if result, err := gojsonschema.Validate(schemaLoader, netDefLoader); err != nil {
		return err
	} else if result.Valid() {
		return nil
	} else {
		for _, errDesc := range result.Errors() {
			log.error("validation error: %s\n", errDesc)
		}
		return fmt.Errorf("validation failed. Check the log for more info")
	}
}
