// +build integration

package netsh

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/exec"
)

func TestGetInterfaces(t *testing.T) {
	h := New(exec.New())
	interfaces, err := h.GetInterfaces()
	assert.NoError(t, err)
	t.Logf("%+v", interfaces)
}

func TestGetInterfaceByName(t *testing.T) {
	h := New(exec.New())
	netInterface, err := h.GetInterfaceByName("Ethernet")
	assert.NoError(t, err)
	t.Logf("%+v", netInterface)
}

func TestGetDefaultGatewayIfaceName(t *testing.T) {
	h := New(exec.New())
	ifname, err := h.GetDefaultGatewayIfaceName()
	assert.NoError(t, err)
	t.Logf("Default interface is: '%v'", ifname)
}

func TestForwarding(t *testing.T) {
	h := New(exec.New())
	err := h.EnableForwarding("Wi-Fi")
	assert.NoError(t, err)
}

func TestSetDNSServer(t *testing.T) {
	h := New(exec.New())
	err := h.SetDNSServer("Wi-Fi", "127.0.0.1")
	assert.NoError(t, err)
}
