package route

import (
	"fmt"
	"strconv"
	"strings"

	utilexec "k8s.io/utils/exec"
)

type Interface interface {
	AddRoute(iface string, cidr string, gateway string) error
	AddRoutes(routes []RouteData) error
	DeleteRoute(dst string, mask string) error
	DeleteRoutes(routes []DeleteRouteData) error
}

type runner struct {
	exec utilexec.Interface
}

const (
	cmdRouting string = "route"
)

// New returns a new Interface which will exec netsh.
func New(exec utilexec.Interface) Interface {

	if exec == nil {
		exec = utilexec.New()
	}

	runner := &runner{
		exec: exec,
	}
	return runner
}

type RouteData struct {
	Dst     string
	Mask    string
	Gateway string
}
type DeleteRouteData struct {
	Dst  string
	Mask string
}

// add static route
func (runner *runner) AddRoute(dst string, mask string, gateway string) error {
	args := []string{
		"ADD", strconv.Quote(dst), "MASK", strconv.Quote(mask), strconv.Quote(gateway),
	}
	cmd := strings.Join(args, " ")
	if stdout, err := runner.exec.Command(cmdRouting, args...).CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add route on, error: %v. cmd: %v. stdout: %v", err.Error(), cmd, string(stdout))
	}
	return nil
}

// add static route
func (runner *runner) DeleteRoute(dst string, mask string) error {
	args := []string{
		"DELETE", strconv.Quote(dst), "MASK", strconv.Quote(mask),
	}
	cmd := strings.Join(args, " ")
	if stdout, err := runner.exec.Command(cmdRouting, args...).CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add route on, error: %v. cmd: %v. stdout: %v", err.Error(), cmd, string(stdout))
	}
	return nil
}

// delete multiple routes
func (runner *runner) DeleteRoutes(routes []DeleteRouteData) error {
	for _, route := range routes {
		if err := runner.DeleteRoute(route.Dst, route.Mask); err != nil {
			return err
		}
	}
	return nil
}

func (runner *runner) AddRoutes(routes []RouteData) error {
	for _, route := range routes {
		if err := runner.AddRoute(route.Dst, route.Mask, route.Gateway); err != nil {
			return err
		}
	}
	return nil
}
