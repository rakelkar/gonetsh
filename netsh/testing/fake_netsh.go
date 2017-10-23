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

package testing

import (
	netsh "github.com/rakelkar/gonetsh/netsh"
)

// no-op implementation of netsh Interface
type FakeNetsh struct {
}

func NewFake() *FakeNetsh {
	return &FakeNetsh{}
}

// GetInterfaces uses the show addresses command and returns a formatted structure
func (*FakeNetsh) GetInterfaces() ([]netsh.Ipv4Interface, error) {
	return nil, nil
}

func (*FakeNetsh) GetInterfaceNameToIndexMap() (map[string]int, error) {
	return map[string]int{}, nil
}

func (*FakeNetsh) EnsurePortProxyRule(args []string) (bool, error) {
	return true, nil
}

// DeletePortProxyRule deletes the specified portproxy rule.  If the rule did not exist, return error.
func (*FakeNetsh) DeletePortProxyRule(args []string) error {
	// Do Nothing
	return nil
}

// DeleteIPAddress checks if the specified IP address is present and, if so, deletes it.
func (*FakeNetsh) DeleteIPAddress(args []string) error {
	// Do Nothing
	return nil
}

// Restore runs `netsh exec` to restore portproxy or addresses using a file.
// TODO Check if this is required, most likely not
func (*FakeNetsh) Restore(args []string) error {
	// Do Nothing
	return nil
}

// GetDefaultGatewayIfaceName returns a fake default interface
func (*FakeNetsh) GetDefaultGatewayIfaceName() (string, error) {
	return "Some Default Interface 1", nil
}

// Gets an interface by name
func (*FakeNetsh) GetInterfaceByName(name string) (netsh.Ipv4Interface, error) {
	return netsh.Ipv4Interface{}, nil
}

// Gets an interface by ip address
func (*FakeNetsh) GetInterfaceByIP(ipAddr string) (netsh.Ipv4Interface, error) {
	return netsh.Ipv4Interface{}, nil
}
// Enable forwarding on the interface (name or index)
func (*FakeNetsh) EnableForwarding(iface string) error {
	return nil
}
var _ = netsh.Interface(&FakeNetsh{})
