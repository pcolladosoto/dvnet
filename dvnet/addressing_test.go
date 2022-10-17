package dvnet

import (
	"testing"
)

func TestAddresserInstantiation(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"192.168.0.0/24", "192.168.0.1/24"},
		{"10.0.0.0/24", "10.0.0.1/24"},
	}

	for _, test := range tests {
		addresser, _ := newSubnetAddresser("addresserTest", cidrParserWrapper(test.in))
		if nextCIDR := addresser.nextCIDR(); nextCIDR != test.want {
			t.Errorf("nextCIDR(%s); netDef = %s; wanted %s", test.in, nextCIDR, test.want)
		}
	}
}
