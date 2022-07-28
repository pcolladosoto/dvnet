package dvnet

import (
	"testing"

	// Note cmp.Equal() tends to panic! This makes it unsuitable
	// for production, but it's generally okay to use it in tests!
	"github.com/google/go-cmp/cmp"
)

var jsonNetDefs = []string{
	`{
		"name": "Test Net 0",
		"outbound_access": true,
		"update_hosts": true,
		"subnets": {
			"A": {
				"cidr": "10.0.0.0/24",
				"hosts": ["A-1", "A-2"]
			},
			"B": {
				"cidr": "10.0.1.0/24",
				"hosts": ["B-1", "B-2"]
			}
		},
		"routers": {
			"R-1": {
				"fw_rules": {"POLICY": "ACCEPT", "ACCEPT": [], "DROP": [["A-1", "B-1", true]]},
				"subnets": ["A", "B"]
			},
			"R-2": {
				"fw_rules": {},
				"subnets": ["A", "B"]
			}
		}
	}
	`,
	`{
		"name": "Test Net 1",
		"outbound_access": true,
		"update_hosts": false,
		"subnets": {
			"A": {
				"cidr": "10.0.0.0/24",
				"hosts": ["A-1", "A-2", "A-3"]
			},
			"B": {
				"cidr": "10.0.1.0/24",
				"hosts": ["B-1", "B-2", "B-3"]
			},
			"C": {
				"cidr": "10.0.2.0/24",
				"hosts": ["C-1", "C-2", "C-3"]
			}
		},
		"routers": {
			"R-1": {
				"fw_rules": {},
				"subnets": ["A"]
			},
			"R-2": {
				"fw_rules": {},
				"subnets": ["A", "B"]
			},
			"R-3": {
				"fw_rules": {},
				"subnets": ["B", "C"]
			}
		}
	}
	`,
	`{
		"name": "Test Net 2",
		"outbound_access": true,
		"update_hosts": true,
		"subnets": {
			"A": {
				"cidr": "10.0.0.0/24",
				"hosts": ["A-1", "A-2", "A-3"]
			},
			"B": {
				"cidr": "10.0.1.0/24",
				"hosts": ["B-1", "B-2", "B-3"]
			},
			"C": {
				"cidr": "10.0.2.0/24",
				"hosts": ["C-1", "C-2", "C-3"]
			},
			"D": {
				"cidr": "10.0.3.0/24",
				"hosts": ["D-1", "D-2", "D-3"]
			},
			"E": {
				"cidr": "10.0.4.0/24",
				"hosts": ["E-1", "E-2", "E-3"]
			},
			"F": {
				"cidr": "10.0.5.0/24",
				"hosts": ["F-1", "F-2", "F-3"]
			},
			"G": {
				"cidr": "10.0.6.0/24",
				"hosts": ["G-1", "G-2", "G-3"]
			},
			"H": {
				"cidr": "10.0.7.0/24",
				"hosts": ["H-1", "H-2", "H-3"]
			},
			"I": {
				"cidr": "10.0.8.0/24",
				"hosts": ["I-1", "I-2", "I-3"]
			},
			"J": {
				"cidr": "10.0.9.0/24",
				"hosts": ["J-1", "J-2", "J-3"]
			},
			"K": {
				"cidr": "10.0.10.0/24",
				"hosts": ["K-1", "K-2", "K-3"]
			},
			"L": {
				"cidr": "10.0.11.0/24",
				"hosts": ["L-1", "L-2", "L-3"]
			}
		},
		"routers": {
			"R-1": {
				"fw_rules": {},
				"subnets": ["A"]
			},
			"R-2": {
				"fw_rules": {},
				"subnets": [
					"A", "B", "C", "D", "E", "F",
					"G", "H", "I", "J", "K", "L"
				]
			}
		}
	}
	`,
}

func TestNetDefLoading(t *testing.T) {
	tests := []struct {
		in   string
		want netDef
	}{
		{
			jsonNetDefs[0], netDef{
				Name: "Test Net 0", OutboundAccess: true, UpdateHostsFile: true,
				Subnets: map[string]subnetDef{
					"A": {CIDRBlock: "10.0.0.0/24", Hosts: []string{"A-1", "A-2"}},
					"B": {CIDRBlock: "10.0.1.0/24", Hosts: []string{"B-1", "B-2"}}},
				Routers: map[string]routerDef{
					"R-1": {
						FWRules: fwRuleDef{Policy: "ACCEPT", Accept: [][]fwTargetDef{}, Drop: [][]fwTargetDef{{"A-1", "B-1", true}}},
						Subnets: []string{"A", "B"},
					},
					"R-2": {
						FWRules: fwRuleDef{Policy: "", Accept: [][]fwTargetDef(nil), Drop: [][]fwTargetDef(nil)},
						Subnets: []string{"A", "B"},
					},
				},
			},
		},
		{
			jsonNetDefs[1], netDef{
				Name: "Test Net 1", OutboundAccess: true, UpdateHostsFile: false,
				Subnets: map[string]subnetDef{
					"A": {CIDRBlock: "10.0.0.0/24", Hosts: []string{"A-1", "A-2", "A-3"}},
					"B": {CIDRBlock: "10.0.1.0/24", Hosts: []string{"B-1", "B-2", "B-3"}},
					"C": {CIDRBlock: "10.0.2.0/24", Hosts: []string{"C-1", "C-2", "C-3"}}},
				Routers: map[string]routerDef{
					"R-1": {
						FWRules: fwRuleDef{Policy: "", Accept: [][]fwTargetDef(nil), Drop: [][]fwTargetDef(nil)},
						Subnets: []string{"A"},
					},
					"R-2": {
						FWRules: fwRuleDef{Policy: "", Accept: [][]fwTargetDef(nil), Drop: [][]fwTargetDef(nil)},
						Subnets: []string{"A", "B"},
					},
					"R-3": {
						FWRules: fwRuleDef{Policy: "", Accept: [][]fwTargetDef(nil), Drop: [][]fwTargetDef(nil)},
						Subnets: []string{"B", "C"},
					},
				},
			},
		},
		{
			jsonNetDefs[2], netDef{
				Name: "Test Net 2", OutboundAccess: true, UpdateHostsFile: true,
				Subnets: map[string]subnetDef{
					"A": {CIDRBlock: "10.0.0.0/24", Hosts: []string{"A-1", "A-2", "A-3"}},
					"B": {CIDRBlock: "10.0.1.0/24", Hosts: []string{"B-1", "B-2", "B-3"}},
					"C": {CIDRBlock: "10.0.2.0/24", Hosts: []string{"C-1", "C-2", "C-3"}},
					"D": {CIDRBlock: "10.0.3.0/24", Hosts: []string{"D-1", "D-2", "D-3"}},
					"E": {CIDRBlock: "10.0.4.0/24", Hosts: []string{"E-1", "E-2", "E-3"}},
					"F": {CIDRBlock: "10.0.5.0/24", Hosts: []string{"F-1", "F-2", "F-3"}},
					"G": {CIDRBlock: "10.0.6.0/24", Hosts: []string{"G-1", "G-2", "G-3"}},
					"H": {CIDRBlock: "10.0.7.0/24", Hosts: []string{"H-1", "H-2", "H-3"}},
					"I": {CIDRBlock: "10.0.8.0/24", Hosts: []string{"I-1", "I-2", "I-3"}},
					"J": {CIDRBlock: "10.0.9.0/24", Hosts: []string{"J-1", "J-2", "J-3"}},
					"K": {CIDRBlock: "10.0.10.0/24", Hosts: []string{"K-1", "K-2", "K-3"}},
					"L": {CIDRBlock: "10.0.11.0/24", Hosts: []string{"L-1", "L-2", "L-3"}}},
				Routers: map[string]routerDef{
					"R-1": {
						FWRules: fwRuleDef{Policy: "", Accept: [][]fwTargetDef(nil), Drop: [][]fwTargetDef(nil)},
						Subnets: []string{"A"},
					},
					"R-2": {
						FWRules: fwRuleDef{Policy: "", Accept: [][]fwTargetDef(nil), Drop: [][]fwTargetDef(nil)},
						Subnets: []string{
							"A", "B", "C", "D", "E", "F",
							"G", "H", "I", "J", "K", "L"},
					},
				},
			},
		},
	}

	for i, test := range tests {
		got, err := parseDef([]byte(test.in))
		if err != nil || !cmp.Equal(got, test.want) {
			t.Errorf("parseDef(test#%d) = %#v; wanted %#v; err %v", i, got, test.want, err)
		}
	}
}

func TestNetDefValidation(t *testing.T) {
	tests := []struct {
		in string
	}{{jsonNetDefs[0]}, {jsonNetDefs[1]}, {jsonNetDefs[2]}}

	for i, test := range tests {
		if err := validateDef(test.in); err != nil {
			t.Errorf("validateDef(test#%d); err %+v", i, err)
		}
	}
}

func TestNetDefValidationFoo(t *testing.T) {
	tests := []struct {
		in string
	}{{jsonNetDefs[0]}, {jsonNetDefs[1]}, {jsonNetDefs[2]}}

	for i, test := range tests {
		def, err := parseDef([]byte(test.in))
		if err != nil {
			t.Errorf("failed parsing net definition %d", i)
		}
		if err := validateDefFoo(def); err != nil {
			t.Errorf("validateDef(test#%d); netDef = %+v; err %+v", i, def, err)
		}
	}
}
