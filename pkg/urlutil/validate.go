package urlutil

import (
	"fmt"
	"net"
	"net/url"
)

var privateRanges = []net.IPNet{
	{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)},
	{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)},
	{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)},
	{IP: net.ParseIP("127.0.0.0"), Mask: net.CIDRMask(8, 32)},
	{IP: net.ParseIP("169.254.0.0"), Mask: net.CIDRMask(16, 32)},
	{IP: net.ParseIP("::1"), Mask: net.CIDRMask(128, 128)},
	{IP: net.ParseIP("fc00::"), Mask: net.CIDRMask(7, 128)},
}

// ValidatePublicHTTPURL checks that rawURL uses http/https and does not resolve
// to a private or loopback address, preventing SSRF attacks.
func ValidatePublicHTTPURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("url scheme %q not allowed, only http/https", parsed.Scheme)
	}

	hostname := parsed.Hostname()
	ips, err := net.LookupHost(hostname)
	if err != nil {
		return fmt.Errorf("resolve host %q: %w", hostname, err)
	}
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		for _, r := range privateRanges {
			if r.Contains(ip) {
				return fmt.Errorf("host %q resolves to private address %s", hostname, ipStr)
			}
		}
	}
	return nil
}
