// Package network reads and changes the host's IPv4 configuration. Every system
// command goes through pkg/sysexec, which adds a timeout, hides the console
// window on Windows and keeps stderr in the returned error. Commands that
// mutate the configuration additionally go through sysexec.RunPrivileged.
package network

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/yakutozcan/fast-ip-changer/pkg/sysexec"
)

// PowerShell queries used on Windows. Both are locale-independent, unlike the
// human-readable netsh tables, and each returns data for every adapter at once.
const (
	psListAdapters  = "Get-NetAdapter | Select-Object Name,Status | ConvertTo-Json -Compress"
	psListAddresses = "Get-NetIPAddress -AddressFamily IPv4 | Select-Object InterfaceAlias,IPAddress | ConvertTo-Json -Compress"
	// Dhcp is stringified in the query itself: ConvertTo-Json renders a raw enum
	// as its integer value, and the numbering is not worth relying on.
	psListDHCP = "Get-NetIPInterface -AddressFamily IPv4 | Select-Object InterfaceAlias,@{Name='Dhcp';Expression={$_.Dhcp.ToString()}} | ConvertTo-Json -Compress"
)

// macPrefsPath is SystemConfiguration's store, which records how every network
// service gets its address. Reading it needs no privileges and covers all
// services at once, unlike `networksetup -getinfo`, which is one call each.
const macPrefsPath = "/Library/Preferences/SystemConfiguration/preferences.plist"

// Adapter.Mode values. An empty Mode means the configuration method could not
// be determined; the UI shows no badge rather than guessing.
const (
	ModeDHCP   = "dhcp"
	ModeManual = "manual"
)

// Adapter is a network interface as shown in the UI. Enabled means the adapter
// is not administratively disabled; a connected/disconnected cable does not
// change it.
type Adapter struct {
	Name      string `json:"name"`
	IPAddress string `json:"ipAddress"`
	Enabled   bool   `json:"enabled"`
	// Mode is ModeDHCP, ModeManual or "" when it could not be determined.
	Mode string `json:"mode"`
}

// GetAdapters lists the network adapters and their current IPv4 addresses.
func GetAdapters(ctx context.Context) ([]Adapter, error) {
	ctx, cancel := sysexec.WithTimeout(ctx)
	defer cancel()

	switch runtime.GOOS {
	case "windows":
		return getAdaptersWindows(ctx)
	case "darwin":
		return getAdaptersDarwin(ctx)
	default:
		return nil, errUnsupported()
	}
}

// SetStaticIP applies a manual IPv4 configuration. dns may be empty, in which
// case the existing DNS configuration is left untouched; otherwise it accepts a
// comma- or space-separated list of servers.
func SetStaticIP(ctx context.Context, adapterName, ip, subnet, gateway, dns string) error {
	if err := validateAdapterName(adapterName); err != nil {
		return err
	}
	if err := validateIPv4(fieldIP, ip); err != nil {
		return err
	}
	if err := validateNetmask(fieldSubnet, subnet); err != nil {
		return err
	}
	if err := validateIPv4(fieldGateway, gateway); err != nil {
		return err
	}
	dnsServers, err := validateDNSList(dns)
	if err != nil {
		return err
	}

	adapterName = strings.TrimSpace(adapterName)
	ip, subnet, gateway = strings.TrimSpace(ip), strings.TrimSpace(subnet), strings.TrimSpace(gateway)

	ctx, cancel := sysexec.WithTimeout(ctx)
	defer cancel()

	// The address and the DNS list go out as one batch so that macOS asks for
	// authorisation once instead of once per command.
	switch runtime.GOOS {
	case "windows":
		cmds := [][]string{{"netsh", "interface", "ipv4", "set", "address",
			"name=" + adapterName, "static", ip, subnet, gateway}}
		return sysexec.RunPrivilegedBatch(ctx, append(cmds, dnsCommandsWindows(adapterName, dnsServers)...))
	case "darwin":
		cmds := [][]string{{"networksetup", "-setmanual", adapterName, ip, subnet, gateway}}
		return sysexec.RunPrivilegedBatch(ctx, append(cmds, dnsCommandsDarwin(adapterName, dnsServers)...))
	default:
		return errUnsupported()
	}
}

// SetDHCP switches the adapter back to automatic addressing. DNS is reset to
// automatic as well: without that, a static DNS set earlier would silently
// survive the switch and keep overriding whatever DHCP hands out.
func SetDHCP(ctx context.Context, adapterName string) error {
	if err := validateAdapterName(adapterName); err != nil {
		return err
	}
	adapterName = strings.TrimSpace(adapterName)

	ctx, cancel := sysexec.WithTimeout(ctx)
	defer cancel()

	switch runtime.GOOS {
	case "windows":
		return sysexec.RunPrivilegedBatch(ctx, [][]string{
			{"netsh", "interface", "ipv4", "set", "address", "name=" + adapterName, "source=dhcp"},
			{"netsh", "interface", "ipv4", "set", "dns", "name=" + adapterName, "source=dhcp"},
		})
	case "darwin":
		return sysexec.RunPrivilegedBatch(ctx, [][]string{
			{"networksetup", "-setdhcp", adapterName},
			// "Empty" is how networksetup clears a static DNS list.
			{"networksetup", "-setdnsservers", adapterName, "Empty"},
		})
	default:
		return errUnsupported()
	}
}

// EnableAdapter administratively enables the adapter.
func EnableAdapter(ctx context.Context, adapterName string) error {
	return setAdapterState(ctx, adapterName, true)
}

// DisableAdapter administratively disables the adapter.
func DisableAdapter(ctx context.Context, adapterName string) error {
	return setAdapterState(ctx, adapterName, false)
}

func setAdapterState(ctx context.Context, adapterName string, enable bool) error {
	if err := validateAdapterName(adapterName); err != nil {
		return err
	}
	adapterName = strings.TrimSpace(adapterName)

	ctx, cancel := sysexec.WithTimeout(ctx)
	defer cancel()

	switch runtime.GOOS {
	case "windows":
		state := "admin=disabled"
		if enable {
			state = "admin=enabled"
		}
		return sysexec.RunPrivileged(ctx, "netsh", "interface", "set", "interface",
			"name="+adapterName, state)
	case "darwin":
		state := "off"
		if enable {
			state = "on"
		}
		return sysexec.RunPrivileged(ctx, "networksetup", "-setnetworkserviceenabled",
			adapterName, state)
	default:
		return errUnsupported()
	}
}

// dnsCommandsWindows builds the netsh calls that write the DNS list: the first
// server replaces the existing configuration, the rest are appended with an
// explicit index. No servers yields no commands, i.e. "leave as is".
func dnsCommandsWindows(adapterName string, servers []string) [][]string {
	if len(servers) == 0 {
		return nil
	}
	cmds := [][]string{{"netsh", "interface", "ipv4", "set", "dns",
		"name=" + adapterName, "static", servers[0]}}
	for i, server := range servers[1:] {
		cmds = append(cmds, []string{"netsh", "interface", "ipv4", "add", "dns",
			"name=" + adapterName, server, "index=" + strconv.Itoa(i+2)})
	}
	return cmds
}

// dnsCommandsDarwin writes the whole DNS list in a single networksetup call.
func dnsCommandsDarwin(adapterName string, servers []string) [][]string {
	if len(servers) == 0 {
		return nil
	}
	return [][]string{append([]string{"networksetup", "-setdnsservers", adapterName}, servers...)}
}

// ---------------------------------------------------------------------------
// Windows enumeration
// ---------------------------------------------------------------------------

// getAdaptersWindows prefers PowerShell, whose structured output is immune to
// the console language. Where PowerShell is blocked by policy it falls back to
// parsing netsh, which still works but has to guess localised state words.
func getAdaptersWindows(ctx context.Context) ([]Adapter, error) {
	adapterJSON, err := powerShell(ctx, psListAdapters)
	if err == nil {
		addressJSON, addrErr := powerShell(ctx, psListAddresses)
		if addrErr != nil {
			addressJSON = ""
		}
		adapters, parseErr := parsePowerShellAdapters(adapterJSON, addressJSON)
		if parseErr == nil && len(adapters) > 0 {
			applyWindowsModes(ctx, adapters)
			return adapters, nil
		}
	}
	return getAdaptersNetsh(ctx)
}

// applyWindowsModes annotates the adapters with DHCP-vs-manual. It is
// best-effort: on failure Mode stays empty and the UI shows no badge.
func applyWindowsModes(ctx context.Context, adapters []Adapter) {
	out, err := powerShell(ctx, psListDHCP)
	if err != nil {
		return
	}
	modes := parsePowerShellDHCP(out)
	for i := range adapters {
		adapters[i].Mode = modes[adapters[i].Name]
	}
}

func powerShell(ctx context.Context, script string) (string, error) {
	return sysexec.Output(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", script)
}

func getAdaptersNetsh(ctx context.Context) ([]Adapter, error) {
	out, err := sysexec.Output(ctx, "netsh", "interface", "show", "interface")
	if err != nil {
		return nil, err
	}

	adapters := parseNetshInterfaces(out)
	if len(adapters) == 0 {
		return nil, fmt.Errorf("ağ adaptörleri listelenemedi")
	}

	// One call for every interface, instead of one call per interface.
	if addrOut, addrErr := sysexec.Output(ctx, "netsh", "interface", "ipv4", "show", "addresses"); addrErr == nil {
		ips := parseNetshAddresses(addrOut)
		for i := range adapters {
			adapters[i].IPAddress = ips[adapters[i].Name]
		}
	}
	return adapters, nil
}

// ---------------------------------------------------------------------------
// macOS enumeration
// ---------------------------------------------------------------------------

// getAdaptersDarwin uses two subprocesses in total: the service order (which
// reports the name, the disabled marker and the BSD device in one shot) plus a
// single ifconfig for the addresses. The previous implementation ran one
// `networksetup -getinfo` per service, disabled ones included.
func getAdaptersDarwin(ctx context.Context) ([]Adapter, error) {
	orderOut, orderErr := sysexec.Output(ctx, "networksetup", "-listnetworkserviceorder")
	if orderErr == nil {
		if services := parseMacServiceOrder(orderOut); len(services) > 0 {
			return macAdaptersFromServices(ctx, services, true), nil
		}
	}

	listOut, listErr := sysexec.Output(ctx, "networksetup", "-listallnetworkservices")
	if listErr != nil {
		if orderErr != nil {
			return nil, orderErr
		}
		return nil, listErr
	}
	return macAdaptersFromServices(ctx, parseMacServices(listOut), false), nil
}

// macAdaptersFromServices resolves an IP for every enabled service. Disabled
// services have no address, so they are never looked up. When the service list
// carries BSD device names, a single ifconfig covers all of them; only services
// without a device fall back to an individual -getinfo call.
func macAdaptersFromServices(ctx context.Context, services []macService, useIfconfig bool) []Adapter {
	deviceIPs := map[string]string{}
	if useIfconfig {
		if out, err := sysexec.Output(ctx, "ifconfig", "-a"); err == nil {
			deviceIPs = parseIfconfig(out)
		}
	}

	// One plutil call yields the mode for every service at once. Both lookups
	// are best-effort: an unreadable store just leaves Mode empty.
	modesByDevice, modesByName := map[string]string{}, map[string]string{}
	if out, err := sysexec.Output(ctx, "plutil", "-convert", "json", "-o", "-", macPrefsPath); err == nil {
		modesByDevice, modesByName = parseMacConfigMethods(out)
	}

	adapters := make([]Adapter, 0, len(services))
	for _, svc := range services {
		adapter := Adapter{Name: svc.Name, Enabled: svc.Enabled}
		// The device is the reliable key: the plist's service name does not
		// always match the service list (a tethered iPhone is "iPhone" there
		// and "iPhone USB" here), so the name is only a fallback.
		if svc.Device != "" {
			adapter.Mode = modesByDevice[svc.Device]
		}
		if adapter.Mode == "" {
			adapter.Mode = modesByName[svc.Name]
		}
		if svc.Enabled {
			if svc.Device != "" {
				adapter.IPAddress = deviceIPs[svc.Device]
			} else if out, err := sysexec.Output(ctx, "networksetup", "-getinfo", svc.Name); err == nil {
				adapter.IPAddress = parseMacGetInfo(out)
				if adapter.Mode == "" {
					adapter.Mode = parseMacGetInfoMode(out)
				}
			}
		}
		adapters = append(adapters, adapter)
	}
	return adapters
}

func errUnsupported() error {
	return fmt.Errorf("desteklenmeyen platform: %s", runtime.GOOS)
}
