// haha+build integration

package netsh_test

import (
	"testing"

	netsh "github.com/rakelkar/gonetsh"
	"k8s.io/utils/exec"
)

func TestGetInterfaceToAddIP(t *testing.T) {
	execer := exec.New()
	h := netsh.New(execer)
	ifname := h.GetInterfaceToAddIP()
	t.Log(ifname)
}

func TestGetDefaultGatewayIfaceName(t *testing.T) {
	execer := exec.New()
	h := netsh.New(execer)
	ifname, err := h.GetDefaultGatewayIfaceName()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Default interface is: '%v'", ifname)
}
