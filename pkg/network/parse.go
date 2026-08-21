package network

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// This file holds only pure parsers: every function takes command output as a
// string and returns parsed data. Keeping them free of exec calls is what makes
// the locale handling testable (see network_test.go).

// columnGap matches the padding netsh puts between two table columns.
var columnGap = regexp.MustCompile(`[ \t]{2,}`)

// macServiceEntry matches a line of `networksetup -listnetworkserviceorder`,
// e.g. "(1) Wi-Fi" or "(*) Bluetooth PAN" where an asterisk means disabled. The
// asterisk is accepted inside or in front of the index, since only its presence
// is documented, not its exact position.
var macServiceEntry = regexp.MustCompile(`^(\*?\((?:\*|\d+)\))[ \t]+(.+)$`)

// macDeviceLine matches "(Hardware Port: Wi-Fi, Device: en0)".
var macDeviceLine = regexp.MustCompile(`Device:[ \t]*([^,)\s]+)`)

// normalizeLines splits command output into lines, tolerating CRLF and a BOM.
func normalizeLines(out string) []string {
	out = strings.TrimPrefix(out, "\ufeff")
	out = strings.ReplaceAll(out, "\r\n", "\n")
	out = strings.ReplaceAll(out, "\r", "\n")
	return strings.Split(out, "\n")
}

// ---------------------------------------------------------------------------
// Windows: PowerShell (primary path)
// ---------------------------------------------------------------------------

// psStatus accepts both the string and the numeric rendering of an enum, since
// ConvertTo-Json is not consistent about it across PowerShell versions.
type psStatus string

func (s *psStatus) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err == nil {
		*s = psStatus(str)
		return nil
	}
	var num json.Number
	if err := json.Unmarshal(b, &num); err == nil {
		*s = psStatus(num.String())
		return nil
	}
	*s = ""
	return nil
}

type psAdapter struct {
	Name   string   `json:"Name"`
	Status psStatus `json:"Status"`
}

type psAddress struct {
	InterfaceAlias string `json:"InterfaceAlias"`
	IPAddress      string `json:"IPAddress"`
}

// unmarshalPSList decodes `ConvertTo-Json -Compress` output into a slice.
// PowerShell emits a bare object rather than a one-element array when the
// pipeline produced exactly one item, so both shapes are accepted.
func unmarshalPSList[T any](out string) ([]T, error) {
	trimmed := strings.TrimSpace(strings.TrimPrefix(out, "\ufeff"))
	if trimmed == "" || trimmed == "null" {
		return nil, nil
	}

	var list []T
	if err := json.Unmarshal([]byte(trimmed), &list); err == nil {
		return list, nil
	}

	var single T
	if err := json.Unmarshal([]byte(trimmed), &single); err != nil {
		return nil, fmt.Errorf("PowerShell çıktısı okunamadı: %w", err)
	}
	return []T{single}, nil
}

// parsePowerShellAdapters joins the two PowerShell queries into the adapter
// list. Enabled means "not administratively disabled", so a plugged-out cable
// (Status "Disconnected") still counts as enabled.
func parsePowerShellAdapters(adapterJSON, addressJSON string) ([]Adapter, error) {
	psAdapters, err := unmarshalPSList[psAdapter](adapterJSON)
	if err != nil {
		return nil, err
	}

	// Addresses are best-effort: an adapter with no IP is still worth listing.
	ips := map[string]string{}
	if psAddresses, err := unmarshalPSList[psAddress](addressJSON); err == nil {
		for _, a := range psAddresses {
			alias := strings.TrimSpace(a.InterfaceAlias)
			ip := strings.TrimSpace(a.IPAddress)
			if alias == "" || !isIPv4(ip) {
				continue
			}
			// Prefer a routable address over a 169.254.x.x self-assigned one.
			if existing, ok := ips[alias]; ok && !isLinkLocal(existing) {
				continue
			}
			ips[alias] = ip
		}
	}

	adapters := make([]Adapter, 0, len(psAdapters))
	for _, a := range psAdapters {
		name := strings.TrimSpace(a.Name)
		if name == "" {
			continue
		}
		adapters = append(adapters, Adapter{
			Name:      name,
			IPAddress: ips[name],
			Enabled:   !strings.EqualFold(strings.TrimSpace(string(a.Status)), "Disabled"),
		})
	}
	return adapters, nil
}

func isLinkLocal(ip string) bool {
	return strings.HasPrefix(ip, "169.254.")
}

// ---------------------------------------------------------------------------
// Windows: netsh (fallback path, used when PowerShell is unavailable)
// ---------------------------------------------------------------------------

// disabledAdminStates lists the localised values netsh prints in the "Admin
// State" column for an administratively disabled interface. Anything unknown
// is treated as enabled, which is the safe default for a fallback path.
var disabledAdminStates = map[string]bool{
	"disabled":     true, // en
	"devre dışı":   true, // tr
	"devre disi":   true, // tr, ASCII-folded console output
	"deaktiviert":  true, // de
	"désactivé":    true, // fr
	"desactivado":  true, // es
	"disabilitato": true, // it
	"desativado":   true, // pt
	"отключено":    true, // ru
}

// parseNetshInterfaces parses `netsh interface show interface` without relying
// on the console language. The old field-index approach broke on Turkish, where
// both the admin state ("Devre Dışı") and the connect state ("Bağlantı
// kesildi") are two words. Instead the table body is located via the dashed
// rule and each row is split on runs of two or more spaces, which is how netsh
// pads its columns; the interface name is always the last chunk.
func parseNetshInterfaces(out string) []Adapter {
	lines := normalizeLines(out)

	start := -1
	for i, line := range lines {
		if isDashRule(strings.TrimSpace(line)) {
			start = i + 1
			break
		}
	}
	if start < 0 {
		// No rule found: assume the first non-empty line is the header.
		for i, line := range lines {
			if strings.TrimSpace(line) != "" {
				start = i + 1
				break
			}
		}
	}
	if start < 0 {
		return nil
	}

	var adapters []Adapter
	for _, line := range lines[start:] {
		line = strings.TrimSpace(line)
		if line == "" || isDashRule(line) {
			continue
		}
		cols := columnGap.Split(line, -1)
		if len(cols) < 2 {
			continue
		}
		name := strings.TrimSpace(cols[len(cols)-1])
		if name == "" || isPseudoInterface(name) {
			continue
		}
		adapters = append(adapters, Adapter{
			Name:    name,
			Enabled: !disabledAdminStates[strings.ToLower(strings.TrimSpace(cols[0]))],
		})
	}
	return adapters
}

// parseNetshAddresses parses `netsh interface ipv4 show addresses` for every
// interface at once. Block headers are the only unindented lines and always
// carry the interface name in quotes, which holds in every locale.
func parseNetshAddresses(out string) map[string]string {
	ips := map[string]string{}
	current := ""

	for _, line := range normalizeLines(out) {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			current = ""
			if first, last := strings.Index(line, `"`), strings.LastIndex(line, `"`); first >= 0 && last > first {
				current = strings.TrimSpace(line[first+1 : last])
			}
			continue
		}
		if current == "" || ips[current] != "" {
			continue
		}
		value := line
		if idx := strings.LastIndex(line, ":"); idx >= 0 {
			value = line[idx+1:]
		}
		// Skip "Subnet Prefix: 192.168.1.0/24 (mask ...)" style lines.
		if strings.Contains(value, "/") {
			continue
		}
		for _, field := range strings.Fields(value) {
			if isIPv4(field) {
				ips[current] = field
				break
			}
		}
	}
	return ips
}

// isPseudoInterface filters the virtual interfaces netsh reports but
// Get-NetAdapter does not, so both code paths return comparable lists.
func isPseudoInterface(name string) bool {
	lower := strings.ToLower(name)
	for _, needle := range []string{"loopback", "pseudo-interface", "isatap", "teredo"} {
		if strings.Contains(lower, needle) {
			return true
		}
	}
	return false
}

func isDashRule(line string) bool {
	if len(line) < 3 {
		return false
	}
	return strings.Trim(line, "-") == ""
}

// ---------------------------------------------------------------------------
// macOS
// ---------------------------------------------------------------------------

// macService is one entry of the network service order, with the BSD device it
// is bound to (used to look its IP up in a single ifconfig call).
type macService struct {
	Name    string
	Device  string
	Enabled bool
}

// parseMacServiceOrder parses `networksetup -listnetworkserviceorder`, which
// reports the service name, whether it is disabled and its BSD device in one
// go. A "(*)" index marks a disabled service; the informational header line
// does not match the entry pattern and is skipped implicitly.
func parseMacServiceOrder(out string) []macService {
	var services []macService
	for _, line := range normalizeLines(out) {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if m := macServiceEntry.FindStringSubmatch(trimmed); m != nil {
			name := strings.TrimSpace(m[2])
			if name == "" {
				continue
			}
			services = append(services, macService{
				Name:    name,
				Enabled: !strings.Contains(m[1], "*"),
			})
			continue
		}
		if len(services) == 0 {
			continue
		}
		if m := macDeviceLine.FindStringSubmatch(trimmed); m != nil {
			services[len(services)-1].Device = m[1]
		}
	}
	return services
}

// parseMacServices parses the flat `networksetup -listallnetworkservices` list,
// used as a fallback. A leading "*" means the service is disabled; the header
// line explaining that is recognised by its literal "(*)".
func parseMacServices(out string) []macService {
	var services []macService
	for _, line := range normalizeLines(out) {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "(*)") {
			continue
		}
		enabled := true
		if strings.HasPrefix(line, "*") {
			enabled = false
			line = strings.TrimSpace(strings.TrimPrefix(line, "*"))
		}
		if line == "" {
			continue
		}
		services = append(services, macService{Name: line, Enabled: enabled})
	}
	return services
}

// parseIfconfig maps a BSD device name to its first IPv4 address.
func parseIfconfig(out string) map[string]string {
	ips := map[string]string{}
	device := ""

	for _, line := range normalizeLines(out) {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			device = ""
			if idx := strings.Index(line, ":"); idx > 0 {
				device = line[:idx]
			}
			continue
		}
		if device == "" || ips[device] != "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "inet" && isIPv4(fields[1]) {
			ips[device] = fields[1]
		}
	}
	return ips
}

// parseMacGetInfo pulls the IPv4 address out of `networksetup -getinfo`.
// The "IPv6 IP address:" line does not share the prefix, so it is not matched.
func parseMacGetInfo(out string) string {
	for _, line := range normalizeLines(out) {
		if !strings.HasPrefix(line, "IP address:") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(line, "IP address:"))
		if isIPv4(value) {
			return value
		}
		return ""
	}
	return ""
}

// ---------------------------------------------------------------------------
// Configuration mode (DHCP vs manual)
// ---------------------------------------------------------------------------

// normalizeConfigMethod maps a platform's spelling of an IPv4 configuration
// method onto an Adapter.Mode value. Anything that is neither a plain manual
// address nor plain DHCP (BOOTP, INFORM, PPP, link-local only) is reported as
// unknown rather than guessed at, since the UI only distinguishes those two.
func normalizeConfigMethod(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "manual":
		return ModeManual
	case "dhcp":
		return ModeDHCP
	default:
		return ""
	}
}

// psDHCPInterface is one row of the Get-NetIPInterface query.
type psDHCPInterface struct {
	InterfaceAlias string   `json:"InterfaceAlias"`
	Dhcp           psStatus `json:"Dhcp"`
}

// parsePowerShellDHCP maps an interface alias to its configuration mode. The
// query stringifies the enum, but the numeric rendering is tolerated too.
// Unrecognised values are omitted, leaving the adapter's mode unknown.
func parsePowerShellDHCP(out string) map[string]string {
	modes := map[string]string{}

	entries, err := unmarshalPSList[psDHCPInterface](out)
	if err != nil {
		return modes
	}

	for _, entry := range entries {
		alias := strings.TrimSpace(entry.InterfaceAlias)
		if alias == "" {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(string(entry.Dhcp))) {
		case "enabled", "1":
			modes[alias] = ModeDHCP
		case "disabled", "0":
			modes[alias] = ModeManual
		}
	}
	return modes
}

// macPrefs is the subset of SystemConfiguration's preferences.plist that says
// how each network service obtains its IPv4 address.
type macPrefs struct {
	NetworkServices map[string]struct {
		UserDefinedName string `json:"UserDefinedName"`
		IPv4            struct {
			ConfigMethod string `json:"ConfigMethod"`
		} `json:"IPv4"`
		Interface struct {
			DeviceName string `json:"DeviceName"`
		} `json:"Interface"`
	} `json:"NetworkServices"`
}

// parseMacConfigMethods reads the JSON rendering of preferences.plist and
// returns the configuration mode keyed by BSD device and, as a fallback, by
// service name. Callers should prefer the device: the plist's UserDefinedName
// does not always match what networksetup reports for the same service.
func parseMacConfigMethods(out string) (byDevice, byName map[string]string) {
	byDevice, byName = map[string]string{}, map[string]string{}

	trimmed := strings.TrimSpace(strings.TrimPrefix(out, "\ufeff"))
	if trimmed == "" {
		return byDevice, byName
	}

	var prefs macPrefs
	if err := json.Unmarshal([]byte(trimmed), &prefs); err != nil {
		return byDevice, byName
	}

	for _, svc := range prefs.NetworkServices {
		mode := normalizeConfigMethod(svc.IPv4.ConfigMethod)
		if mode == "" {
			continue
		}
		if device := strings.TrimSpace(svc.Interface.DeviceName); device != "" {
			byDevice[device] = mode
		}
		if name := strings.TrimSpace(svc.UserDefinedName); name != "" {
			byName[name] = mode
		}
	}
	return byDevice, byName
}

// parseMacGetInfoMode reads the configuration mode from `networksetup -getinfo`,
// whose first non-empty line is "DHCP Configuration" or "Manual Configuration".
// Only used for services the service list reported without a BSD device.
func parseMacGetInfoMode(out string) string {
	for _, line := range normalizeLines(out) {
		trimmed := strings.ToLower(strings.TrimSpace(line))
		if trimmed == "" {
			continue
		}
		switch {
		case strings.HasPrefix(trimmed, "manual"):
			return ModeManual
		case strings.HasPrefix(trimmed, "dhcp"):
			return ModeDHCP
		}
		return ""
	}
	return ""
}
