package network

import (
	"reflect"
	"strings"
	"testing"
)

func TestIsIPv4(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"plain", "192.168.1.10", true},
		{"zero", "0.0.0.0", true},
		{"broadcast", "255.255.255.255", true},
		{"surrounding whitespace", "  10.0.0.1  ", true},
		{"empty", "", false},
		{"octet out of range", "192.168.1.999", false},
		{"three octets", "192.168.1", false},
		{"five octets", "1.2.3.4.5", false},
		{"leading zeros", "192.168.001.1", false},
		{"trailing dot", "192.168.1.1.", false},
		{"ipv6", "fe80::1", false},
		{"ipv4 mapped ipv6", "::ffff:192.168.1.1", false},
		{"with prefix", "192.168.1.0/24", false},
		{"hostname", "router.local", false},
		{"negative octet", "192.168.-1.1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIPv4(tt.value); got != tt.want {
				t.Fatalf("isIPv4(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestValidateIPv4Message(t *testing.T) {
	err := validateIPv4(fieldIP, "192.168.1.999")
	if err == nil {
		t.Fatal("expected an error for 192.168.1.999")
	}
	if got, want := err.Error(), "IP adresi geçersiz: 192.168.1.999"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
	if err := validateIPv4(fieldGateway, "192.168.1.1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateNetmask(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"class c", "255.255.255.0", false},
		{"class b", "255.255.0.0", false},
		{"class a", "255.0.0.0", false},
		{"/23", "255.255.254.0", false},
		{"/30", "255.255.255.252", false},
		{"/32", "255.255.255.255", false},
		{"non contiguous middle", "255.0.255.0", true},
		{"non contiguous last octet", "255.255.255.253", true},
		{"non contiguous 255.255.4.0", "255.255.4.0", true},
		{"all zero", "0.0.0.0", true},
		{"not an ip", "255.255.255", true},
		{"empty", "", true},
		{"octet out of range", "255.255.255.256", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNetmask(fieldSubnet, tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateNetmask(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
			if err != nil && !strings.HasPrefix(err.Error(), fieldSubnet) {
				t.Fatalf("error %q does not name the field", err)
			}
		})
	}
}

func TestParseDNSList(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"whitespace only", "   \t ", nil},
		{"single", "8.8.8.8", []string{"8.8.8.8"}},
		{"comma", "8.8.8.8,8.8.4.4", []string{"8.8.8.8", "8.8.4.4"}},
		{"comma and spaces", " 8.8.8.8 ,  8.8.4.4 ", []string{"8.8.8.8", "8.8.4.4"}},
		{"space separated", "8.8.8.8 1.1.1.1", []string{"8.8.8.8", "1.1.1.1"}},
		{"semicolon", "8.8.8.8;1.1.1.1", []string{"8.8.8.8", "1.1.1.1"}},
		{"newline", "8.8.8.8\n1.1.1.1", []string{"8.8.8.8", "1.1.1.1"}},
		{"three servers", "8.8.8.8, 8.8.4.4, 1.1.1.1", []string{"8.8.8.8", "8.8.4.4", "1.1.1.1"}},
		{"trailing separator", "8.8.8.8,", []string{"8.8.8.8"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDNSList(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseDNSList(%q) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}

func TestValidateDNSList(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []string
		wantErr string
	}{
		{"empty keeps dns untouched", "", nil, ""},
		{"valid pair", "8.8.8.8, 8.8.4.4", []string{"8.8.8.8", "8.8.4.4"}, ""},
		{"invalid second entry", "8.8.8.8, 8.8.4.300", nil, "DNS adresi geçersiz: 8.8.4.300"},
		{"hostname rejected", "dns.google", nil, "DNS adresi geçersiz: dns.google"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateDNSList(tt.in)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("servers = %#v, want %#v", got, tt.want)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAdapterName(t *testing.T) {
	if err := validateAdapterName(""); err == nil {
		t.Fatal("expected an error for an empty adapter name")
	}
	if err := validateAdapterName("   "); err == nil {
		t.Fatal("expected an error for a blank adapter name")
	}
	if err := validateAdapterName("Wi-Fi"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParsePowerShellAdaptersArray(t *testing.T) {
	adapterJSON := `[{"Name":"Ethernet","Status":"Up"},{"Name":"Wi-Fi","Status":"Disabled"},{"Name":"Ethernet 2","Status":"Disconnected"}]`
	addressJSON := `[{"InterfaceAlias":"Ethernet","IPAddress":"192.168.1.24"},{"InterfaceAlias":"Ethernet 2","IPAddress":"169.254.10.1"},{"InterfaceAlias":"Loopback Pseudo-Interface 1","IPAddress":"127.0.0.1"}]`

	got, err := parsePowerShellAdapters(adapterJSON, addressJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Adapter{
		{Name: "Ethernet", IPAddress: "192.168.1.24", Enabled: true},
		{Name: "Wi-Fi", IPAddress: "", Enabled: false},
		{Name: "Ethernet 2", IPAddress: "169.254.10.1", Enabled: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("adapters = %#v, want %#v", got, want)
	}
}

func TestParsePowerShellAdaptersSingleObject(t *testing.T) {
	// PowerShell drops the array wrapper when the pipeline yields one item.
	adapterJSON := `{"Name":"Ethernet","Status":"Up"}`
	addressJSON := `{"InterfaceAlias":"Ethernet","IPAddress":"10.0.0.7"}`

	got, err := parsePowerShellAdapters(adapterJSON, addressJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Adapter{{Name: "Ethernet", IPAddress: "10.0.0.7", Enabled: true}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("adapters = %#v, want %#v", got, want)
	}
}

func TestParsePowerShellAdaptersEdgeCases(t *testing.T) {
	t.Run("numeric status enum", func(t *testing.T) {
		got, err := parsePowerShellAdapters(`{"Name":"Ethernet","Status":2}`, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || !got[0].Enabled {
			t.Fatalf("adapters = %#v, want one enabled adapter", got)
		}
	})

	t.Run("empty output", func(t *testing.T) {
		got, err := parsePowerShellAdapters("", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("adapters = %#v, want none", got)
		}
	})

	t.Run("garbage output", func(t *testing.T) {
		if _, err := parsePowerShellAdapters("Get-NetAdapter is not recognized", ""); err == nil {
			t.Fatal("expected an error for non-JSON output")
		}
	})

	t.Run("bad address json is tolerated", func(t *testing.T) {
		got, err := parsePowerShellAdapters(`{"Name":"Ethernet","Status":"Up"}`, "not json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].IPAddress != "" {
			t.Fatalf("adapters = %#v, want one adapter with no IP", got)
		}
	})

	t.Run("routable address wins over link local", func(t *testing.T) {
		addressJSON := `[{"InterfaceAlias":"Ethernet","IPAddress":"169.254.3.4"},{"InterfaceAlias":"Ethernet","IPAddress":"192.168.1.5"}]`
		got, err := parsePowerShellAdapters(`{"Name":"Ethernet","Status":"Up"}`, addressJSON)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got[0].IPAddress != "192.168.1.5" {
			t.Fatalf("IPAddress = %q, want 192.168.1.5", got[0].IPAddress)
		}
	})
}

// The regression test for the Turkish locale bug: the adapter name used to be
// taken from strings.Fields(line)[3:], which shifts as soon as a localised
// state value contains a space ("Devre Dışı", "Bağlantı kesildi").
func TestParseNetshInterfacesLocales(t *testing.T) {
	englishOut := "\r\n" +
		"Admin State    State          Type             Interface Name\r\n" +
		"-------------------------------------------------------------------------\r\n" +
		"Enabled        Connected      Dedicated        Ethernet\r\n" +
		"Disabled       Disconnected   Dedicated        Wi-Fi 2\r\n" +
		"Enabled        Connected      Internal         Loopback Pseudo-Interface 1\r\n"

	turkishOut := "\r\n" +
		"Yönetici Durumu Durum          Tür              Arabirim Adı\r\n" +
		"-------------------------------------------------------------------------\r\n" +
		"Etkin          Bağlı          Ayrılmış         Ethernet\r\n" +
		"Devre Dışı     Bağlantı kesildi Ayrılmış       Yerel Ağ Bağlantısı 2\r\n" +
		"Etkin          Bağlantı kesildi Ayrılmış       Wi-Fi\r\n"

	tests := []struct {
		name string
		out  string
		want []Adapter
	}{
		{
			name: "english",
			out:  englishOut,
			want: []Adapter{
				{Name: "Ethernet", Enabled: true},
				{Name: "Wi-Fi 2", Enabled: false},
			},
		},
		{
			name: "turkish",
			out:  turkishOut,
			want: []Adapter{
				{Name: "Ethernet", Enabled: true},
				{Name: "Yerel Ağ Bağlantısı 2", Enabled: false},
				{Name: "Wi-Fi", Enabled: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseNetshInterfaces(tt.out); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("adapters = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseNetshInterfacesNoRule(t *testing.T) {
	// Without the dashed rule the header line is skipped by position.
	out := "Admin State    State          Type             Interface Name\n" +
		"Enabled        Connected      Dedicated        Ethernet\n"
	want := []Adapter{{Name: "Ethernet", Enabled: true}}
	if got := parseNetshInterfaces(out); !reflect.DeepEqual(got, want) {
		t.Fatalf("adapters = %#v, want %#v", got, want)
	}
}

func TestParseNetshInterfacesEmpty(t *testing.T) {
	if got := parseNetshInterfaces(""); got != nil {
		t.Fatalf("adapters = %#v, want nil", got)
	}
}

func TestParseNetshAddresses(t *testing.T) {
	out := "\r\n" +
		"Configuration for interface \"Ethernet\"\r\n" +
		"    DHCP enabled:                         No\r\n" +
		"    IP Address:                           192.168.1.24\r\n" +
		"    Subnet Prefix:                        192.168.1.0/24 (mask 255.255.255.0)\r\n" +
		"    Default Gateway:                      192.168.1.1\r\n" +
		"    Gateway Metric:                       0\r\n" +
		"    InterfaceMetric:                      25\r\n" +
		"\r\n" +
		"\"Yerel Ağ Bağlantısı 2\" arabirimi için yapılandırma\r\n" +
		"    DHCP etkin:                           Evet\r\n" +
		"    IP Adresi:                            10.10.0.5\r\n" +
		"    Alt Ağ Öneki:                         10.10.0.0/16 (mask 255.255.0.0)\r\n" +
		"\r\n"

	got := parseNetshAddresses(out)
	want := map[string]string{
		"Ethernet":              "192.168.1.24",
		"Yerel Ağ Bağlantısı 2": "10.10.0.5",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("addresses = %#v, want %#v", got, want)
	}
}

func TestParseMacServiceOrder(t *testing.T) {
	out := "An asterisk (*) denotes that a network service is disabled.\n" +
		"(1) Wi-Fi\n" +
		"(Hardware Port: Wi-Fi, Device: en0)\n" +
		"\n" +
		"(2) Thunderbolt Ethernet\n" +
		"(Hardware Port: Thunderbolt Ethernet, Device: en4)\n" +
		"\n" +
		"(*) Bluetooth PAN\n" +
		"(Hardware Port: Bluetooth PAN, Device: en5)\n" +
		"\n" +
		"*(4) iPhone USB\n" +
		"(Hardware Port: iPhone USB, Device: en6)\n" +
		"\n"

	got := parseMacServiceOrder(out)
	want := []macService{
		{Name: "Wi-Fi", Device: "en0", Enabled: true},
		{Name: "Thunderbolt Ethernet", Device: "en4", Enabled: true},
		{Name: "Bluetooth PAN", Device: "en5", Enabled: false},
		{Name: "iPhone USB", Device: "en6", Enabled: false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("services = %#v, want %#v", got, want)
	}
}

func TestParseMacServices(t *testing.T) {
	out := "An asterisk (*) denotes that a network service is disabled.\n" +
		"Wi-Fi\n" +
		"Thunderbolt Ethernet\n" +
		"*Bluetooth PAN\n" +
		"\n"

	got := parseMacServices(out)
	want := []macService{
		{Name: "Wi-Fi", Enabled: true},
		{Name: "Thunderbolt Ethernet", Enabled: true},
		{Name: "Bluetooth PAN", Enabled: false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("services = %#v, want %#v", got, want)
	}
}

func TestParseIfconfig(t *testing.T) {
	out := "lo0: flags=8049<UP,LOOPBACK,RUNNING,MULTICAST> mtu 16384\n" +
		"\tinet 127.0.0.1 netmask 0xff000000\n" +
		"en0: flags=8863<UP,BROADCAST,SMART,RUNNING,SIMPLEX,MULTICAST> mtu 1500\n" +
		"\tether aa:bb:cc:dd:ee:ff\n" +
		"\tinet6 fe80::1%en0 prefixlen 64 secured scopeid 0xc\n" +
		"\tinet 192.168.1.37 netmask 0xffffff00 broadcast 192.168.1.255\n" +
		"en4: flags=8863<UP,BROADCAST,SMART,RUNNING,SIMPLEX,MULTICAST> mtu 1500\n" +
		"\tether 11:22:33:44:55:66\n" +
		"en5: flags=8822<BROADCAST,SMART,SIMPLEX,MULTICAST> mtu 1500\n"

	got := parseIfconfig(out)
	want := map[string]string{"lo0": "127.0.0.1", "en0": "192.168.1.37"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ips = %#v, want %#v", got, want)
	}
}

func TestParseMacGetInfo(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want string
	}{
		{
			name: "configured",
			out: "DHCP Configuration\n" +
				"IP address: 192.168.1.37\n" +
				"Subnet mask: 255.255.255.0\n" +
				"Router: 192.168.1.1\n" +
				"IPv6 IP address: fe80::1\n",
			want: "192.168.1.37",
		},
		{
			name: "no address",
			out:  "DHCP Configuration\nIP address: none\nSubnet mask: none\n",
			want: "",
		},
		{
			name: "not configured at all",
			out:  "DHCP Configuration\nClient ID:\n",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseMacGetInfo(tt.out); got != tt.want {
				t.Fatalf("parseMacGetInfo() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSetStaticIPValidatesBeforeRunning(t *testing.T) {
	// No command may run when the input is invalid, so these are safe to call
	// on any machine and must fail on validation, not on exec.
	tests := []struct {
		name    string
		adapter string
		ip      string
		subnet  string
		gateway string
		dns     string
		wantErr string
	}{
		{"empty adapter", "", "192.168.1.5", "255.255.255.0", "192.168.1.1", "", "adaptör adı boş olamaz"},
		{"bad ip", "Ethernet", "192.168.1.999", "255.255.255.0", "192.168.1.1", "", "IP adresi geçersiz: 192.168.1.999"},
		{"bad mask", "Ethernet", "192.168.1.5", "255.0.255.0", "192.168.1.1", "", "Alt ağ maskesi geçersiz: 255.0.255.0"},
		{"bad gateway", "Ethernet", "192.168.1.5", "255.255.255.0", "gw", "", "Ağ geçidi geçersiz: gw"},
		{"bad dns", "Ethernet", "192.168.1.5", "255.255.255.0", "192.168.1.1", "8.8.8.8 8.8.4", "DNS adresi geçersiz: 8.8.4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetStaticIP(nil, tt.adapter, tt.ip, tt.subnet, tt.gateway, tt.dns)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestAdapterNameRequiredEverywhere(t *testing.T) {
	if err := SetDHCP(nil, " "); err == nil {
		t.Fatal("SetDHCP: expected an error for a blank adapter name")
	}
	if err := EnableAdapter(nil, ""); err == nil {
		t.Fatal("EnableAdapter: expected an error for an empty adapter name")
	}
	if err := DisableAdapter(nil, ""); err == nil {
		t.Fatal("DisableAdapter: expected an error for an empty adapter name")
	}
}

func TestDNSCommandsWindows(t *testing.T) {
	tests := []struct {
		name    string
		adapter string
		servers []string
		want    [][]string
	}{
		{
			name:    "no servers leaves configuration untouched",
			adapter: "Ethernet",
			servers: nil,
			want:    nil,
		},
		{
			name:    "single server",
			adapter: "Ethernet",
			servers: []string{"1.1.1.1"},
			want: [][]string{
				{"netsh", "interface", "ipv4", "set", "dns", "name=Ethernet", "static", "1.1.1.1"},
			},
		},
		{
			name:    "three servers are indexed from two",
			adapter: "Yerel Ağ Bağlantısı",
			servers: []string{"1.1.1.1", "8.8.8.8", "9.9.9.9"},
			want: [][]string{
				{"netsh", "interface", "ipv4", "set", "dns", "name=Yerel Ağ Bağlantısı", "static", "1.1.1.1"},
				{"netsh", "interface", "ipv4", "add", "dns", "name=Yerel Ağ Bağlantısı", "8.8.8.8", "index=2"},
				{"netsh", "interface", "ipv4", "add", "dns", "name=Yerel Ağ Bağlantısı", "9.9.9.9", "index=3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dnsCommandsWindows(tt.adapter, tt.servers)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dnsCommandsWindows()\n got: %v\nwant: %v", got, tt.want)
			}
		})
	}
}

func TestDNSCommandsDarwin(t *testing.T) {
	tests := []struct {
		name    string
		adapter string
		servers []string
		want    [][]string
	}{
		{
			name:    "no servers leaves configuration untouched",
			adapter: "Wi-Fi",
			servers: nil,
			want:    nil,
		},
		{
			name:    "all servers in one call",
			adapter: "Wi-Fi",
			servers: []string{"1.1.1.1", "8.8.8.8"},
			want: [][]string{
				{"networksetup", "-setdnsservers", "Wi-Fi", "1.1.1.1", "8.8.8.8"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dnsCommandsDarwin(tt.adapter, tt.servers)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dnsCommandsDarwin()\n got: %v\nwant: %v", got, tt.want)
			}
		})
	}
}

// The adapter name must reach netsh as a bare "name=<value>" argv element. The
// old code wrapped the value in literal quotes, which exec passes through
// verbatim because no shell is involved, so netsh saw the quotes as part of the
// name and could not find the adapter.
func TestDNSCommandsWindowsDoesNotQuoteAdapterName(t *testing.T) {
	cmds := dnsCommandsWindows(`Wi-Fi 2`, []string{"1.1.1.1"})
	if len(cmds) != 1 {
		t.Fatalf("got %d commands, want 1", len(cmds))
	}
	for _, arg := range cmds[0] {
		if strings.HasPrefix(arg, "name=") {
			if strings.Contains(arg, `"`) {
				t.Errorf("adapter argument %q contains literal quotes", arg)
			}
			if arg != "name=Wi-Fi 2" {
				t.Errorf("adapter argument = %q, want %q", arg, "name=Wi-Fi 2")
			}
			return
		}
	}
	t.Error("no name= argument found")
}

func TestNormalizeConfigMethod(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{"DHCP", ModeDHCP},
		{"dhcp", ModeDHCP},
		{"  DHCP  ", ModeDHCP},
		{"Manual", ModeManual},
		{"manual", ModeManual},
		// Neither a plain manual address nor plain DHCP: reported as unknown
		// rather than guessed at.
		{"INFORM", ""},
		{"BOOTP", ""},
		{"PPP", ""},
		{"LinkLocal", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := normalizeConfigMethod(tt.value); got != tt.want {
				t.Errorf("normalizeConfigMethod(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestParsePowerShellDHCP(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want map[string]string
	}{
		{
			name: "array of stringified enums",
			out:  `[{"InterfaceAlias":"Ethernet","Dhcp":"Disabled"},{"InterfaceAlias":"Wi-Fi","Dhcp":"Enabled"}]`,
			want: map[string]string{"Ethernet": ModeManual, "Wi-Fi": ModeDHCP},
		},
		{
			// ConvertTo-Json emits a bare object, not an array, for one row.
			name: "single object",
			out:  `{"InterfaceAlias":"Ethernet","Dhcp":"Enabled"}`,
			want: map[string]string{"Ethernet": ModeDHCP},
		},
		{
			// Belt and braces: the query stringifies the enum, but an older
			// PowerShell serialising it as a number must still work.
			name: "numeric enum",
			out:  `[{"InterfaceAlias":"Ethernet","Dhcp":1},{"InterfaceAlias":"Wi-Fi","Dhcp":0}]`,
			want: map[string]string{"Ethernet": ModeDHCP, "Wi-Fi": ModeManual},
		},
		{
			name: "turkish adapter name is not special-cased",
			out:  `[{"InterfaceAlias":"Yerel Ağ Bağlantısı","Dhcp":"Disabled"}]`,
			want: map[string]string{"Yerel Ağ Bağlantısı": ModeManual},
		},
		{
			name: "unrecognised value is omitted rather than guessed",
			out:  `[{"InterfaceAlias":"Ethernet","Dhcp":"Weird"}]`,
			want: map[string]string{},
		},
		{"empty alias skipped", `[{"InterfaceAlias":"  ","Dhcp":"Enabled"}]`, map[string]string{}},
		{"empty output", ``, map[string]string{}},
		{"null output", `null`, map[string]string{}},
		{"garbage output", `not json`, map[string]string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePowerShellDHCP(tt.out)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePowerShellDHCP()\n got: %v\nwant: %v", got, tt.want)
			}
		})
	}
}

// Sample shaped like the real `plutil -convert json -o - preferences.plist`
// output, including the mismatch that makes the device the reliable key: the
// service is "iPhone" here but "iPhone USB" in the service list.
const macPrefsJSON = `{
  "NetworkServices": {
    "1111-AAAA": {"UserDefinedName":"Wi-Fi","IPv4":{"ConfigMethod":"Manual"},"Interface":{"DeviceName":"en0"}},
    "2222-BBBB": {"UserDefinedName":"Thunderbolt Bridge","IPv4":{"ConfigMethod":"DHCP"},"Interface":{"DeviceName":"bridge0"}},
    "3333-CCCC": {"UserDefinedName":"iPhone","IPv4":{"ConfigMethod":"DHCP"},"Interface":{"DeviceName":"en5"}},
    "4444-DDDD": {"UserDefinedName":"Bluetooth PAN","IPv4":{"ConfigMethod":"INFORM"},"Interface":{"DeviceName":"en6"}},
    "5555-EEEE": {"UserDefinedName":"Old VPN","Interface":{"DeviceName":"utun0"}}
  }
}`

func TestParseMacConfigMethods(t *testing.T) {
	byDevice, byName := parseMacConfigMethods(macPrefsJSON)

	wantDevice := map[string]string{"en0": ModeManual, "bridge0": ModeDHCP, "en5": ModeDHCP}
	if !reflect.DeepEqual(byDevice, wantDevice) {
		t.Errorf("byDevice\n got: %v\nwant: %v", byDevice, wantDevice)
	}

	wantName := map[string]string{"Wi-Fi": ModeManual, "Thunderbolt Bridge": ModeDHCP, "iPhone": ModeDHCP}
	if !reflect.DeepEqual(byName, wantName) {
		t.Errorf("byName\n got: %v\nwant: %v", byName, wantName)
	}
}

func TestParseMacConfigMethodsBadInput(t *testing.T) {
	for _, out := range []string{``, `   `, `not json`, `{}`, `{"NetworkServices":{}}`, `null`} {
		byDevice, byName := parseMacConfigMethods(out)
		if len(byDevice) != 0 || len(byName) != 0 {
			t.Errorf("parseMacConfigMethods(%q) = %v / %v, want empty maps", out, byDevice, byName)
		}
	}
}

// The plist service name can differ from the name networksetup reports, so a
// device-keyed lookup has to win. This is the regression test for that.
func TestMacConfigMethodsPreferDeviceOverName(t *testing.T) {
	byDevice, byName := parseMacConfigMethods(macPrefsJSON)

	if _, ok := byName["iPhone USB"]; ok {
		t.Error("byName unexpectedly contains the service-list spelling")
	}
	if got := byDevice["en5"]; got != ModeDHCP {
		t.Errorf("byDevice[en5] = %q, want %q — the device lookup must cover the name mismatch", got, ModeDHCP)
	}
}

func TestParseMacGetInfoMode(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want string
	}{
		{"dhcp", "DHCP Configuration\nIP address: 192.168.1.20\n", ModeDHCP},
		{"manual", "Manual Configuration\nIP address: 192.168.1.20\n", ModeManual},
		{"leading blank lines", "\n\nDHCP Configuration\n", ModeDHCP},
		{"crlf", "Manual Configuration\r\nIP address: 10.0.0.2\r\n", ModeManual},
		{"no ip configured", "DHCP Configuration\nIP address: none\n", ModeDHCP},
		{"unrecognised first line", "Something Else\nIP address: 10.0.0.2\n", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseMacGetInfoMode(tt.out); got != tt.want {
				t.Errorf("parseMacGetInfoMode() = %q, want %q", got, tt.want)
			}
		})
	}
}
