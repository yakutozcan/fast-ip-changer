package network

import (
	"fmt"
	"net"
	"strings"
	"unicode"
)

// Field labels used in the user-facing (Turkish) validation messages.
const (
	fieldIP      = "IP adresi"
	fieldSubnet  = "Alt ağ maskesi"
	fieldGateway = "Ağ geçidi"
	fieldDNS     = "DNS adresi"
)

// isIPv4 reports whether value is a plain dotted-quad IPv4 literal. IPv4-mapped
// IPv6 forms such as "::ffff:192.168.1.1" are rejected on purpose: netsh and
// networksetup only accept the dotted-quad form.
func isIPv4(value string) bool {
	v := strings.TrimSpace(value)
	if strings.Count(v, ".") != 3 || strings.Contains(v, ":") {
		return false
	}
	ip := net.ParseIP(v)
	return ip != nil && ip.To4() != nil
}

// validateIPv4 checks a single required IPv4 value and names the offending
// field in the error, so the UI can show something better than "invalid input".
func validateIPv4(field, value string) error {
	if !isIPv4(value) {
		return fmt.Errorf("%s geçersiz: %s", field, strings.TrimSpace(value))
	}
	return nil
}

// validateNetmask checks that value is an IPv4 subnet mask with a contiguous
// run of leading one-bits (255.255.0.255 is a dotted quad but not a netmask).
func validateNetmask(field, value string) error {
	if err := validateIPv4(field, value); err != nil {
		return err
	}
	ip := net.ParseIP(strings.TrimSpace(value)).To4()
	ones, bits := net.IPMask(ip).Size()
	if bits != 32 || ones == 0 {
		return fmt.Errorf("%s geçersiz: %s", field, strings.TrimSpace(value))
	}
	return nil
}

// parseDNSList splits a user-entered DNS field into individual servers. Commas,
// semicolons and any whitespace all count as separators, so "8.8.8.8, 8.8.4.4"
// and "8.8.8.8 8.8.4.4" behave the same. An empty field yields nil.
func parseDNSList(dns string) []string {
	fields := strings.FieldsFunc(dns, func(r rune) bool {
		return r == ',' || r == ';' || unicode.IsSpace(r)
	})
	if len(fields) == 0 {
		return nil
	}
	return fields
}

// validateDNSList splits and validates the DNS field. An empty field is valid
// and yields no servers, which means "leave the DNS configuration untouched".
func validateDNSList(dns string) ([]string, error) {
	servers := parseDNSList(dns)
	for _, s := range servers {
		if err := validateIPv4(fieldDNS, s); err != nil {
			return nil, err
		}
	}
	return servers, nil
}

// validateAdapterName rejects an empty adapter name before any command runs.
func validateAdapterName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("adaptör adı boş olamaz")
	}
	return nil
}
