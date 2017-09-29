package netroute

import (
	"regexp"
	"net"
	"strconv"
	"strings"
	"bufio"
	"bytes"
	ps "github.com/gorillalabs/go-powershell"
	psbe "github.com/gorillalabs/go-powershell/backend"

	"fmt"
)

// Interface is an injectable interface for running MSFT_NetRoute commands. Implementations must be goroutine-safe.
type Interface interface {
	// Get all net routes on the host
	GetNetRoutesAll() ([]Route, error)

	// Get net routes by link and destination subnet
	GetNetRoutes(linkIndex int, destinationSubnet *net.IPNet) ([]Route, error)

	// Create a new route
	NewNetRoute(linkIndex int, destinationSubnet *net.IPNet, gatewayAddress net.IP) error

	// Remove an existing route
	RemoveNetRoute(linkIndex int, destinationSubnet *net.IPNet, gatewayAddress net.IP) error
}

type Route struct {
	linkIndex         int
	destinationSubnet *net.IPNet
	gatewayAddress    net.IP
	routeMetric       int
	ifMetric          int
}

type shell struct {
	shellInstance ps.Shell
}

func New() Interface {

	s, _ := ps.New(&psbe.Local{})

	runner := &shell{
		shellInstance: s,
	}

	return runner
}

func (shell *shell) Exit() {
	shell.shellInstance.Exit()
}

func (shell *shell) GetNetRoutesAll() ([]Route, error) {
	getRouteCmdLine := "get-netroute -erroraction Ignore"
	stdout, err := shell.runScript(getRouteCmdLine)
	if err != nil {
		return nil, err
	}
	return parseRoutesList(stdout), nil
}
func (shell *shell) GetNetRoutes(linkIndex int, destinationSubnet *net.IPNet) ([]Route, error) {
	getRouteCmdLine := fmt.Sprintf("get-netroute -InterfaceIndex %v -DestinationPrefix %v -erroraction Ignore", linkIndex, destinationSubnet.String())
	stdout, err := shell.runScript(getRouteCmdLine)
	if err != nil {
		return nil, err
	}
	return parseRoutesList(stdout), nil
}

func (shell *shell) RemoveNetRoute(linkIndex int, destinationSubnet *net.IPNet, gatewayAddress net.IP) error {
	removeRouteCmdLine := fmt.Sprint("remove-netroute -InterfaceIndex %v -DestinationPrefix %v -NextHop  %v -Verbose", linkIndex, destinationSubnet.String(), gatewayAddress.String())
	_, err := shell.runScript(removeRouteCmdLine)

	return err
}

func (shell *shell) NewNetRoute(linkIndex int, destinationSubnet *net.IPNet, gatewayAddress net.IP) error {
	newRouteCmdLine := fmt.Sprintf("new-netroute -InterfaceIndex %v -DestinationPrefix %v -NextHop  %v -Verbose", linkIndex, destinationSubnet.String(), gatewayAddress.String())
	_, err := shell.runScript(newRouteCmdLine)

	return err
}

func parseRoutesList(stdout string) []Route {
	internalWhitespaceRegEx := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	var routes []Route
	for scanner.Scan() {
		line := internalWhitespaceRegEx.ReplaceAllString(scanner.Text(), "|")
		if strings.HasPrefix(line, "ifIndex") || strings.HasPrefix(line, "----") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) != 5 {
			continue
		}

		linkIndex, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		gatewayAddress := net.ParseIP(parts[2])
		if gatewayAddress == nil {
			continue
		}

		_, destinationSubnet, err := net.ParseCIDR(parts[1])
		if err != nil {
			continue
		}
		route := Route{
			destinationSubnet: destinationSubnet,
			gatewayAddress:    gatewayAddress,
			linkIndex:         linkIndex,
		}

		routes = append(routes, route)
	}

	return routes
}

func (r *Route) Equal(route Route) bool {
	if r.destinationSubnet.IP.Equal(route.destinationSubnet.IP) && r.gatewayAddress.Equal(route.gatewayAddress) && bytes.Equal(r.destinationSubnet.Mask, route.destinationSubnet.Mask) {
		return true
	}

	return false
}

func (shell *shell) runScript(cmdLine string) (string, error) {

	stdout, _, err := shell.shellInstance.Execute(cmdLine)
	if err != nil {
		return "", err
	}

	return stdout, nil
}
