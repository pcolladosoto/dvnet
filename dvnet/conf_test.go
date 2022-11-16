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
				Name:            "Test Net 0",
				OutboundAccess:  OutboundAccessDef{Enabled: true, HopCIDR: cidrParserWrapper("192.168.240.0/24")},
				UpdateHostsFile: true,
				Subnets: map[string]subnetDef{
					"A": {CIDRBlock: cidrParserWrapper("10.0.0.0/24"), Hosts: map[string]HostDef{
						"A-1": {"pcollado/dhost"}, "A-2": {"pcollado/dhost"},
					}},
					"B": {CIDRBlock: cidrParserWrapper("10.0.1.0/24"), Hosts: map[string]HostDef{
						"B-1": {"pcollado/dhost"}, "B-2": {"pcollado/dhost"},
					}}},
				Routers: map[string]routerDef{
					"R-1": {
						FWRules: fwRuleDef{Policy: "ACCEPT", Accept: [][]fwTargetDef{}, Drop: [][]fwTargetDef{{"A-1", "B-1", true}}},
						Subnets: []string{"A", "B"},
						Image:   "pcollado/drouter",
					},
					"R-2": {
						FWRules: fwRuleDef{Policy: "", Accept: [][]fwTargetDef(nil), Drop: [][]fwTargetDef(nil)},
						Subnets: []string{"A", "B"},
						Image:   "pcollado/drouter",
					},
				},
			},
		},
		{
			jsonNetDefs[1], netDef{
				Name:            "Test Net 1",
				OutboundAccess:  OutboundAccessDef{Enabled: true, HopCIDR: cidrParserWrapper("192.168.240.0/24")},
				UpdateHostsFile: false,
				Subnets: map[string]subnetDef{
					"A": {CIDRBlock: cidrParserWrapper("10.0.0.0/24"), Hosts: map[string]HostDef{
						"A-1": {"pcollado/dhost"}, "A-2": {"pcollado/dhost"}, "A-3": {"pcollado/dhost"},
					}},
					"B": {CIDRBlock: cidrParserWrapper("10.0.1.0/24"), Hosts: map[string]HostDef{
						"B-1": {"pcollado/dhost"}, "B-2": {"pcollado/dhost"}, "B-3": {"pcollado/dhost"},
					}},
					"C": {CIDRBlock: cidrParserWrapper("10.0.2.0/24"), Hosts: map[string]HostDef{
						"C-1": {"pcollado/dhost"}, "C-2": {"pcollado/dhost"}, "C-3": {"pcollado/dhost"},
					}}},
				Routers: map[string]routerDef{
					"R-1": {
						FWRules: fwRuleDef{Policy: "", Accept: [][]fwTargetDef(nil), Drop: [][]fwTargetDef(nil)},
						Subnets: []string{"A"},
						Image:   "pcollado/drouter",
					},
					"R-2": {
						FWRules: fwRuleDef{Policy: "", Accept: [][]fwTargetDef(nil), Drop: [][]fwTargetDef(nil)},
						Subnets: []string{"A", "B"},
						Image:   "pcollado/drouter",
					},
					"R-3": {
						FWRules: fwRuleDef{Policy: "", Accept: [][]fwTargetDef(nil), Drop: [][]fwTargetDef(nil)},
						Subnets: []string{"B", "C"},
						Image:   "pcollado/drouter",
					},
				},
			},
		},
		{
			jsonNetDefs[2], netDef{
				Name:            "Test Net 2",
				OutboundAccess:  OutboundAccessDef{Enabled: true, HopCIDR: cidrParserWrapper("192.168.240.0/24")},
				UpdateHostsFile: true,
				Subnets: map[string]subnetDef{
					"A": {CIDRBlock: cidrParserWrapper("10.0.0.0/24"), Hosts: map[string]HostDef{
						"A-1": {"pcollado/dhost"}, "A-2": {"pcollado/dhost"}, "A-3": {"pcollado/dhost"},
					}},
					"B": {CIDRBlock: cidrParserWrapper("10.0.1.0/24"), Hosts: map[string]HostDef{
						"B-1": {"pcollado/dhost"}, "B-2": {"pcollado/dhost"}, "B-3": {"pcollado/dhost"},
					}},
					"C": {CIDRBlock: cidrParserWrapper("10.0.2.0/24"), Hosts: map[string]HostDef{
						"C-1": {"pcollado/dhost"}, "C-2": {"pcollado/dhost"}, "C-3": {"pcollado/dhost"},
					}},
					"D": {CIDRBlock: cidrParserWrapper("10.0.3.0/24"), Hosts: map[string]HostDef{
						"D-1": {"pcollado/dhost"}, "D-2": {"pcollado/dhost"}, "D-3": {"pcollado/dhost"},
					}},
					"E": {CIDRBlock: cidrParserWrapper("10.0.4.0/24"), Hosts: map[string]HostDef{
						"E-1": {"pcollado/dhost"}, "E-2": {"pcollado/dhost"}, "E-3": {"pcollado/dhost"},
					}},
					"F": {CIDRBlock: cidrParserWrapper("10.0.5.0/24"), Hosts: map[string]HostDef{
						"F-1": {"pcollado/dhost"}, "F-2": {"pcollado/dhost"}, "F-3": {"pcollado/dhost"},
					}},
					"G": {CIDRBlock: cidrParserWrapper("10.0.6.0/24"), Hosts: map[string]HostDef{
						"G-1": {"pcollado/dhost"}, "G-2": {"pcollado/dhost"}, "G-3": {"pcollado/dhost"},
					}},
					"H": {CIDRBlock: cidrParserWrapper("10.0.7.0/24"), Hosts: map[string]HostDef{
						"H-1": {"pcollado/dhost"}, "H-2": {"pcollado/dhost"}, "H-3": {"pcollado/dhost"},
					}},
					"I": {CIDRBlock: cidrParserWrapper("10.0.8.0/24"), Hosts: map[string]HostDef{
						"I-1": {"pcollado/dhost"}, "I-2": {"pcollado/dhost"}, "I-3": {"pcollado/dhost"},
					}},
					"J": {CIDRBlock: cidrParserWrapper("10.0.9.0/24"), Hosts: map[string]HostDef{
						"J-1": {"pcollado/dhost"}, "J-2": {"pcollado/dhost"}, "J-3": {"pcollado/dhost"},
					}},
					"K": {CIDRBlock: cidrParserWrapper("10.0.10.0/24"), Hosts: map[string]HostDef{
						"K-1": {"pcollado/dhost"}, "K-2": {"pcollado/dhost"}, "K-3": {"pcollado/dhost"},
					}},
					"L": {CIDRBlock: cidrParserWrapper("10.0.11.0/24"), Hosts: map[string]HostDef{
						"L-1": {"pcollado/dhost"}, "L-2": {"pcollado/dhost"}, "L-3": {"pcollado/dhost"},
					}}},
				Routers: map[string]routerDef{
					"R-1": {
						FWRules: fwRuleDef{Policy: "", Accept: [][]fwTargetDef(nil), Drop: [][]fwTargetDef(nil)},
						Subnets: []string{"A"},
						Image:   "pcollado/drouter",
					},
					"R-2": {
						FWRules: fwRuleDef{Policy: "", Accept: [][]fwTargetDef(nil), Drop: [][]fwTargetDef(nil)},
						Subnets: []string{
							"A", "B", "C", "D", "E", "F",
							"G", "H", "I", "J", "K", "L"},
						Image: "pcollado/drouter",
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
		def, err := parseDef([]byte(test.in))
		if err != nil {
			t.Errorf("failed parsing net definition %d", i)
		}
		if err := validateDef(def); err != nil {
			t.Errorf("validateDef(test#%d); netDef = %+v; err %+v", i, def, err)
		}
	}
}
