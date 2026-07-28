package urlguard

import (
	"net"
	"strings"
	"testing"
)

// TestBlockedIP pins the address ranges no abcd fetch may reach. The NAT64 and
// 6to4 cases are the ones a naive guard misses: they wrap a loopback or metadata
// IPv4 inside an IPv6 address that none of net.IP's own predicates flag.
func TestBlockedIP(t *testing.T) {
	blocked := []string{
		"127.0.0.1", "::1", // loopback
		"169.254.169.254",                       // cloud metadata (link-local)
		"10.0.0.1", "192.168.1.1", "172.16.5.5", // private
		"0.0.0.0", "::", // unspecified
		"224.0.0.1", "ff02::1", // multicast
		"fe80::1",            // link-local unicast
		"fc00::1",            // unique local (private)
		"64:ff9b::a9fe:a9fe", // NAT64 wrapping 169.254.169.254
		"64:ff9b::7f00:1",    // NAT64 wrapping 127.0.0.1
		"2002:a9fe:a9fe::",   // 6to4 wrapping 169.254.169.254
		"::ffff:127.0.0.1",   // v4-mapped loopback
	}
	for _, s := range blocked {
		ip := net.ParseIP(s)
		if ip == nil {
			t.Fatalf("test bug: %q is not an IP", s)
		}
		if !BlockedIP(ip) {
			t.Errorf("BlockedIP(%s) = false, want true", s)
		}
	}

	allowed := []string{"93.184.216.34", "2606:2800:220:1:248:1893:25c8:1946", "8.8.8.8"}
	for _, s := range allowed {
		ip := net.ParseIP(s)
		if ip == nil {
			t.Fatalf("test bug: %q is not an IP", s)
		}
		if BlockedIP(ip) {
			t.Errorf("BlockedIP(%s) = true, want false", s)
		}
	}
}

// TestCheckHostRefusesNamesAndLiterals covers the two refusal paths that resolve
// without any DNS traffic: the internal/metadata NAME guard, and IP literals.
func TestCheckHostRefusesNamesAndLiterals(t *testing.T) {
	cases := []struct{ host, want string }{
		{"", "refusing to fetch a URL with no host"},
		{"metadata", "refusing to fetch internal/metadata host"},
		{"metadata.google.internal", "refusing to fetch internal/metadata host"},
		{"svc.internal", "refusing to fetch internal/metadata host"},
		{"127.0.0.1", "refusing to fetch"},
		{"169.254.169.254", "refusing to fetch"},
		{"10.0.0.1", "refusing to fetch"},
	}
	for _, c := range cases {
		err := CheckHost(c.host)
		if err == nil {
			t.Errorf("CheckHost(%q) = nil, want a refusal", c.host)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("CheckHost(%q) = %q, want it to contain %q", c.host, err, c.want)
		}
	}
}

// TestCheckHostWithPolicyAllowsLoopback is what lets a fetch path be exercised
// against an httptest server: the predicate is a parameter, so a test can relax
// it for 127.0.0.1 without the shipped default ever being relaxed.
func TestCheckHostWithPolicyAllowsLoopback(t *testing.T) {
	permissive := func(ip net.IP) bool { return !ip.IsLoopback() && BlockedIP(ip) }
	if err := CheckHostWith("127.0.0.1", permissive); err != nil {
		t.Errorf("CheckHostWith(127.0.0.1, permissive) = %v, want nil", err)
	}
	if err := CheckHostWith("169.254.169.254", permissive); err == nil {
		t.Error("CheckHostWith(169.254.169.254, permissive) = nil, want a refusal")
	}
	// The name guard is policy-independent: a *.internal name is refused whatever
	// the address predicate says, because it never gets as far as an address.
	if err := CheckHostWith("svc.internal", permissive); err == nil {
		t.Error("CheckHostWith(svc.internal, permissive) = nil, want a refusal")
	}
}

// TestDialControlRefusesBlockedAddress pins the connect-time re-check that closes
// the DNS-rebinding gap between the name guard and the transport's own lookup.
func TestDialControlRefusesBlockedAddress(t *testing.T) {
	control := DialControl(BlockedIP)
	if err := control("tcp", "169.254.169.254:80", nil); err == nil {
		t.Error("DialControl allowed a metadata address")
	} else if !strings.Contains(err.Error(), "refusing to connect") {
		t.Errorf("got %q, want a refusing-to-connect error", err)
	}
	if err := control("tcp", "93.184.216.34:443", nil); err != nil {
		t.Errorf("DialControl refused a public address: %v", err)
	}
}
