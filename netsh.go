/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package netsh

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/glog"
	utilexec "k8s.io/utils/exec"
)

// Interface is an injectable interface for running netsh commands.  Implementations must be goroutine-safe.
type Interface interface {
	// EnsurePortProxyRule checks if the specified redirect exists, if not creates it
	EnsurePortProxyRule(args []string) (bool, error)
	// DeletePortProxyRule deletes the specified portproxy rule.  If the rule did not exist, return error.
	DeletePortProxyRule(args []string) error
	// DeleteIPAddress checks if the specified IP address is present and, if so, deletes it.
	DeleteIPAddress(args []string) error
	// Restore runs `netsh exec` to restore portproxy or addresses using a file.
	// TODO Check if this is required, most likely not
	Restore(args []string) error
	// Get the interface name that has the default gateway
	GetDefaultGatewayIfaceName() (string, error)
	// Get a list of interfaces and addresses
	GetInterfaces() ([]Ipv4Interface, error)
	// Gets an interface by name
	GetInterfaceByName(name string) (Ipv4Interface, error)
	// Gets an interface by ip address in the format a.b.c.d
	GetInterfaceByIP(ipAddr string) (Ipv4Interface, error)
}

const (
	cmdNetsh string = "netsh"
)

// runner implements Interface in terms of exec("netsh").
type runner struct {
	mu   sync.Mutex
	exec utilexec.Interface
}

// Ipv4Interface models IPv4 interface output from: netsh interface ipv4 show addresses
type Ipv4Interface struct {
	Name                  string
	InterfaceMetric       int
	DhcpEnabled           bool
	IpAddress             string
	SubnetPrefix          int
	GatewayMetric         int
	DefaultGatewayAddress string
}

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

// GetInterfaces uses the show addresses command and returns a formatted structure
func (runner *runner) GetInterfaces() ([]Ipv4Interface, error) {
	args := []string{
		"interface", "ipv4", "show", "addresses",
	}

	output, err := runner.exec.Command(cmdNetsh, args...).CombinedOutput()
	if err != nil {
		return nil, err
	}
	interfacesString := string(output[:])

	outputLines := strings.Split(interfacesString, "\n")
	var interfaces []Ipv4Interface
	var currentInterface Ipv4Interface
	quotedPattern := regexp.MustCompile("\\\"(.*?)\\\"")
	cidrPattern := regexp.MustCompile("\\/(.*?)\\ ")
	for _, outputLine := range outputLines {
		if strings.Contains(outputLine, "Configuration for interface") {
			if currentInterface != (Ipv4Interface{}) {
				interfaces = append(interfaces, currentInterface)
			}
			match := quotedPattern.FindStringSubmatch(outputLine)
			currentInterface = Ipv4Interface{
				Name: match[1],
			}
		} else {
			parts := strings.SplitN(outputLine, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if strings.HasPrefix(key, "DHCP enabled") {
				if value == "Yes" {
					currentInterface.DhcpEnabled = true
				}
			} else if strings.HasPrefix(key, "InterfaceMetric") {
				if val, err := strconv.Atoi(value); err == nil {
					currentInterface.InterfaceMetric = val
				}
			} else if strings.HasPrefix(key, "Gateway Metric") {
				if val, err := strconv.Atoi(value); err == nil {
					currentInterface.GatewayMetric = val
				}
			} else if strings.HasPrefix(key, "Subnet Prefix") {
				match := cidrPattern.FindStringSubmatch(value)
				if val, err := strconv.Atoi(match[1]); err == nil {
					currentInterface.SubnetPrefix = val
				}
			} else if strings.HasPrefix(key, "IP Address") {
				currentInterface.IpAddress = value
			} else if strings.HasPrefix(key, "Default Gateway") {
				currentInterface.DefaultGatewayAddress = value
			}
		}
	}

	// add the last one
	if currentInterface != (Ipv4Interface{}) {
		interfaces = append(interfaces, currentInterface)
	}

	if len(interfaces) == 0 {
		return nil, fmt.Errorf("no interfaces found in netsh output: %v", interfacesString)
	}

	return interfaces, nil
}

// EnsurePortProxyRule checks if the specified redirect exists, if not creates it.
func (runner *runner) EnsurePortProxyRule(args []string) (bool, error) {
	glog.V(4).Infof("running netsh interface portproxy add v4tov4 %v", args)
	out, err := runner.exec.Command(cmdNetsh, args...).CombinedOutput()

	if err == nil {
		return true, nil
	}
	if ee, ok := err.(utilexec.ExitError); ok {
		// netsh uses exit(0) to indicate a success of the operation,
		// as compared to a malformed commandline, for example.
		if ee.Exited() && ee.ExitStatus() != 0 {
			return false, nil
		}
	}
	return false, fmt.Errorf("error checking portproxy rule: %v: %s", err, out)

}

// DeletePortProxyRule deletes the specified portproxy rule.  If the rule did not exist, return error.
func (runner *runner) DeletePortProxyRule(args []string) error {
	glog.V(4).Infof("running netsh interface portproxy delete v4tov4 %v", args)
	out, err := runner.exec.Command(cmdNetsh, args...).CombinedOutput()

	if err == nil {
		return nil
	}
	if ee, ok := err.(utilexec.ExitError); ok {
		// netsh uses exit(0) to indicate a success of the operation,
		// as compared to a malformed commandline, for example.
		if ee.Exited() && ee.ExitStatus() == 0 {
			return nil
		}
	}
	return fmt.Errorf("error deleting portproxy rule: %v: %s", err, out)
}

// DeleteIPAddress checks if the specified IP address is present and, if so, deletes it.
func (runner *runner) DeleteIPAddress(args []string) error {
	glog.V(4).Infof("running netsh interface ipv4 delete address %v", args)
	out, err := runner.exec.Command(cmdNetsh, args...).CombinedOutput()

	if err == nil {
		return nil
	}
	if ee, ok := err.(utilexec.ExitError); ok {
		// netsh uses exit(0) to indicate a success of the operation,
		// as compared to a malformed commandline, for example.
		if ee.Exited() && ee.ExitStatus() == 0 {
			return nil
		}
	}
	return fmt.Errorf("error deleting ipv4 address: %v: %s", err, out)
}

func (runner *runner) GetDefaultGatewayIfaceName() (string, error) {
	interfaces, err := runner.GetInterfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.DefaultGatewayAddress != "" {
			return iface.Name, nil
		}
	}

	// return "not found"
	return "", fmt.Errorf("Default interface not found")
}

func (runner *runner) GetInterfaceByName(name string) (Ipv4Interface, error) {
	interfaces, err := runner.GetInterfaces()
	if err != nil {
		return Ipv4Interface{}, err
	}

	for _, iface := range interfaces {
		if iface.Name == name {
			return iface, nil
		}
	}

	// return "not found"
	return Ipv4Interface{}, fmt.Errorf("Interface not found: %v", name)
}

func (runner *runner) GetInterfaceByIP(ipAddr string) (Ipv4Interface, error) {
	interfaces, err := runner.GetInterfaces()
	if err != nil {
		return Ipv4Interface{}, err
	}

	for _, iface := range interfaces {
		if iface.IpAddress == ipAddr {
			return iface, nil
		}
	}

	// return "not found"
	return Ipv4Interface{}, fmt.Errorf("Interface not found: %v", ipAddr)
}

// Restore is part of Interface.
func (runner *runner) Restore(args []string) error {
	return nil
}
