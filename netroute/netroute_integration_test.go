// +build integration

package netroute

import (
	"testing"
	"time"

	"net"

	"github.com/stretchr/testify/assert"
)

func TestGetNetRoutes(t *testing.T) {
	nr := New()
	routes, err := nr.GetNetRoutesAll()
	assert.NoError(t, err)
	t.Logf("number of routes %v", len(routes))
	for _, r := range routes {
		t.Logf("%+v\n", r)
	}
}

func TestAddRemoteRoute(t *testing.T) {
	nr := New()

	routes, err := nr.GetNetRoutesAll()
	_, sn, _ := net.ParseCIDR("192.168.111.0/24")
	addr := net.ParseIP("1.2.3.4")
	route := Route{
		LinkIndex:         routes[0].LinkIndex,
		DestinationSubnet: sn,
		GatewayAddress:    addr,
	}

	for _, r := range routes {
		if r.Equal(route) {
			t.Logf("ABORTING route already exists.. %+v ", route)
			return
		}
	}

	err = nr.NewNetRoute(route.LinkIndex, route.DestinationSubnet, route.GatewayAddress)
	assert.NoError(t, err)
	t.Logf("added route %v", route)

	err = nr.RemoveNetRoute(route.LinkIndex, route.DestinationSubnet, route.GatewayAddress)
	assert.NoError(t, err)
	t.Logf("removed route %v", route)
}

func TestLeakyShells(t *testing.T) {
	nr := New()
	done := make(chan bool)
	for i := 0; i < 1000; i++ {
		timeout := time.After(10 * time.Second)

		go func() {
			_, err := nr.GetNetRoutesAll()
			if err != nil {
				t.Errorf("error from GetNetRoutesAll: %v", err)
			}
			done <- true
		}()

		select {
		case <-timeout:
			t.Fatalf("Timed out waiting for command after %d iterations", i)
		case <-done:
			continue
		}
	}
}
