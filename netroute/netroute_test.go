package netroute

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"fmt"
)

type fakeResponse struct {
	StdOut string
	StdErr string
	Err    error
}

const (
	GetAllRoutesCommand = "get-netroute -erroraction Ignore"
	GetRouteStdOut = `
ifIndex DestinationPrefix                              NextHop                                  RouteMetric ifMetric PolicyStore
------- -----------------                              -------                                  ----------- -------- -----------
13      255.255.255.255/32                             0.0.0.0                                          256 25       ActiveStore
12      192.168.10.0/24                                10.244.0.1                                       256 35       ActiveStore
3       255.255.255.255/32                             0.0.0.0                                          256 65       ActiveStore`
)

type fakeShell struct {
	DefaultResponse *fakeResponse
	RequestMap      map[string]fakeResponse
	t               *testing.T
}

func NewFakeShell(t *testing.T) *fakeShell{
	var f fakeShell
	f.t = t
	f.RequestMap = make(map[string]fakeResponse)
	return &f
}

func (fs *fakeShell) Execute(cmd string) (string, string, error) {
	if val, ok := fs.RequestMap[cmd]; ok {
		//do something here
		return val.StdOut, val.StdErr, nil
	}

	if fs.DefaultResponse != nil {
		val := fs.DefaultResponse
		return val.StdOut, val.StdErr, nil
	}

	err := fmt.Errorf("unexpected command %v", cmd)
	assert.Fail(fs.t, err.Error())
	return "", "", err
}

func (fs *fakeShell) Exit() {

}

func GetAllRoutesTest(t *testing.T) {

	fs := NewFakeShell(t)

	fs.RequestMap[GetAllRoutesCommand] = fakeResponse {
		GetRouteStdOut,
		"",
		nil,
	}

	nr := &shell{
		shellInstance: fs,
	}

	routes, err := nr.GetNetRoutesAll()

	assert.Nil(t, err)
	assert.Equal(t,3, len(routes))
	assert.Equal(t, "192.168.10.0/24", routes[2].DestinationSubnet.String())
	assert.Equal(t, 12, routes[2].LinkIndex)
	assert.Equal(t, "10.244.0.1", routes[2].GatewayAddress)
	assert.True(t, routes[0].Equal(routes[0]))
	assert.False(t, routes[0].Equal(routes[1]))
}

func GetAllRoutesEmptyTest(t *testing.T) {

	fs := NewFakeShell(t)

	fs.RequestMap["GetAllRoutesCommand"] = fakeResponse {
		"",
		"",
		nil,
	}

	nr := &shell{
		shellInstance: fs,
	}

	routes, err := nr.GetNetRoutesAll()

	assert.Nil(t, err)
	assert.Equal(t,0, len(routes))
}
