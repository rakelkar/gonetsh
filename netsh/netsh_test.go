package netsh

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/exec"
	fakeexec "k8s.io/utils/exec/testing"
)

func getFakeExecTemplate(fakeCmd *fakeexec.FakeCmd) fakeexec.FakeExec {
	var fakeTemplate []fakeexec.FakeCommandAction
	for i := 0; i < len(fakeCmd.CombinedOutputScript); i++ {
		fakeTemplate = append(fakeTemplate, func(cmd string, args ...string) exec.Cmd { return fakeexec.InitFakeCmd(fakeCmd, cmd, args...) })
	}
	return fakeexec.FakeExec{
		CommandScript: fakeTemplate,
	}
}

func TestGetInterfacesGoldenPath(t *testing.T) {
	fakeCmd := fakeexec.FakeCmd{
		CombinedOutputScript: []fakeexec.FakeAction{
			func() ([]byte, []byte, error) {
				return []byte(`

Configuration for interface "Ethernet"
    DHCP enabled:                         Yes
    InterfaceMetric:                      5

Configuration for interface "Local Area Connection* 1"
    DHCP enabled:                         Yes
    InterfaceMetric:                      25

Configuration for interface "Wi-Fi"
    DHCP enabled:                         Yes
    IP Address:                           10.88.48.68
    Subnet Prefix:                        10.88.48.0/22 (mask 255.255.252.0)
    Default Gateway:                      10.88.48.1
    Gateway Metric:                       0
    InterfaceMetric:                      35

Configuration for interface "Loopback Pseudo-Interface 1"
    DHCP enabled:                         No
    IP Address:                           127.0.0.1
    Subnet Prefix:                        127.0.0.0/8 (mask 255.0.0.0)
    InterfaceMetric:                      75

	`), nil, nil
			},
			func() ([]byte, []byte, error) {
				return []byte(`
			Idx     Met         MTU          State                Name
---  ----------  ----------  ------------  ---------------------------
  9          25        1500  connected     Ethernet
  1          75  4294967295  connected     Loopback Pseudo-Interface 1
  2          15        1500  connected     Local Area Connection* 1
 14          15        1500  connected     Wi-Fi`), nil, nil
			},
		},
	}

	fakeExec := getFakeExecTemplate(&fakeCmd)

	runner := runner{
		exec: &fakeExec,
	}

	interfaces, err := runner.GetInterfaces()
	assert.NoError(t, err)
	assert.EqualValues(t, 2, fakeCmd.CombinedOutputCalls)
	assert.EqualValues(t, strings.Split("netsh interface ipv4 show config", " "), fakeCmd.CombinedOutputLog[0])
	assert.EqualValues(t, 4, len(interfaces))
	assert.EqualValues(t, Ipv4Interface{
		Idx:                   14,
		DhcpEnabled:           true,
		IpAddress:             "10.88.48.68",
		SubnetPrefix:          22,
		DefaultGatewayAddress: "10.88.48.1",
		GatewayMetric:         0,
		InterfaceMetric:       35,
		Name:                  "Wi-Fi",
	}, interfaces[2])
}

func TestGetInterfacesFailsGracefully(t *testing.T) {

	fakeCmd := fakeexec.FakeCmd{
		CombinedOutputScript: []fakeexec.FakeAction{
			// Failure.
			func() ([]byte, []byte, error) { return nil, nil, &fakeexec.FakeExitError{Status: 2} },
			// Empty Response.
			func() ([]byte, []byte, error) { return []byte{}, nil, nil },
			// Junk Response.
			func() ([]byte, []byte, error) { return []byte("fake error from netsh"), nil, nil },
		},
	}

	fakeExec := getFakeExecTemplate(&fakeCmd)

	runner := runner{
		exec: &fakeExec,
	}

	interfaces, err := runner.getIpAddressConfigurations()
	assert.Error(t, err)
	assert.Nil(t, interfaces)

	interfaces, err = runner.getIpAddressConfigurations()
	assert.Error(t, err)
	assert.Nil(t, interfaces)

	interfaces, err = runner.getIpAddressConfigurations()
	assert.Error(t, err)
	assert.Nil(t, interfaces)

	assert.EqualValues(t, 3, fakeCmd.CombinedOutputCalls)
	assert.EqualValues(t, strings.Split("netsh interface ipv4 show config", " "), fakeCmd.CombinedOutputLog[0])
}

func TestGetInterfaceNameToIndexMap(t *testing.T) {
	fake := fakeexec.FakeCmd{
		CombinedOutputScript: []fakeexec.FakeAction{
			func() ([]byte, []byte, error) { return []byte(`badinput`), nil, nil },
			func() ([]byte, []byte, error) {
				return []byte(`
			Idx     Met         MTU          State                Name
---  ----------  ----------  ------------  ---------------------------
  9          25        1500  connected     Ethernet
  1          75  4294967295  connected     Loopback Pseudo-Interface 1
  2          15        1500  connected     vEthernet (New Virtual Switch)
 14          15        1500  connected     vEthernet (HNS Internal NIC)`), nil, nil
			},
		},
	}

	fakeExec := getFakeExecTemplate(&fake)

	runner := runner{
		exec: &fakeExec,
	}

	// Test bad input
	idxMap, err := runner.getNetworkInterfaceParameters()

	assert.NotNil(t, err)
	assert.Nil(t, idxMap)

	// Test good input
	idxMap, err = runner.getNetworkInterfaceParameters()

	assert.Nil(t, err)
	assert.NotNil(t, idxMap)
	assert.Equal(t, 9, idxMap["Ethernet"])
	assert.Equal(t, 14, idxMap["vEthernet (HNS Internal NIC)"])
}
