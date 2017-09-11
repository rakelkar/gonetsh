// haha+build integration

package netsh_test

import (
	"testing"

	netsh "github.com/rakelkar/gonetsh"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/exec"
)

func TestGetInterfaces(t *testing.T) {
	h := netsh.New(exec.New())
	interfaces, err := h.GetInterfaces()
	assert.NoError(t, err)
	t.Logf("%+v", interfaces)
}

func TestGetDefaultGatewayIfaceName(t *testing.T) {
	h := netsh.New(exec.New())
	ifname, err := h.GetDefaultGatewayIfaceName()
	assert.NoError(t, err)
	t.Logf("Default interface is: '%v'", ifname)
}
