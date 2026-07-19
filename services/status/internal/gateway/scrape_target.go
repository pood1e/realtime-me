package gateway

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"

	mev1 "github.com/pood1e/realtime-me/gen/go/realtime/me/status/v1"
)

// ScrapeTargetPolicy is the only network space an enrolled probe may expose to
// Prometheus. Requiring literal IP addresses avoids DNS rebinding after a target
// has passed validation.
type ScrapeTargetPolicy struct {
	allowedPrefixes []netip.Prefix
	port            uint16
}

func NewScrapeTargetPolicy(cidrs []string, port int) (ScrapeTargetPolicy, error) {
	if len(cidrs) == 0 {
		return ScrapeTargetPolicy{}, errors.New("probe.allowed_cidrs is required")
	}
	if port < 1 || port > 65535 {
		return ScrapeTargetPolicy{}, errors.New("probe.port must be between 1 and 65535")
	}

	prefixes := make([]netip.Prefix, 0, len(cidrs))
	seen := make(map[netip.Prefix]struct{}, len(cidrs))
	for _, cidr := range cidrs {
		cidr = strings.TrimSpace(cidr)
		prefix, err := netip.ParsePrefix(cidr)
		if err != nil {
			return ScrapeTargetPolicy{}, fmt.Errorf("probe.allowed_cidrs contains %q: %w", cidr, err)
		}
		prefix = prefix.Masked()
		if _, exists := seen[prefix]; exists {
			continue
		}
		seen[prefix] = struct{}{}
		prefixes = append(prefixes, prefix)
	}
	return ScrapeTargetPolicy{allowedPrefixes: prefixes, port: uint16(port)}, nil
}

func (policy ScrapeTargetPolicy) validate() error {
	if len(policy.allowedPrefixes) == 0 || policy.port == 0 {
		return errors.New("probe target policy is required")
	}
	return nil
}

func (policy ScrapeTargetPolicy) Validate(target *mev1.ScrapeTarget) error {
	if target.GetJob() != mev1.ScrapeJob_SCRAPE_JOB_PROBE {
		return errors.New("scrape target job must be SCRAPE_JOB_PROBE")
	}
	host, port, err := net.SplitHostPort(target.GetTarget())
	if err != nil {
		return fmt.Errorf("scrape target must be host:port: %w", err)
	}
	address, err := netip.ParseAddr(host)
	if err != nil || address.Zone() != "" {
		return errors.New("scrape target host must be a literal IP address without a zone")
	}
	address = address.Unmap()
	if address.IsUnspecified() || address.IsLoopback() || address.IsMulticast() {
		return errors.New("scrape target host must be a routable unicast address")
	}
	number, err := strconv.ParseUint(port, 10, 16)
	if err != nil || uint16(number) != policy.port {
		return fmt.Errorf("scrape target port must be %d", policy.port)
	}
	for _, prefix := range policy.allowedPrefixes {
		if prefix.Contains(address) {
			return nil
		}
	}
	return errors.New("scrape target host is outside probe.allowed_cidrs")
}
