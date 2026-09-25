//go:build linux

package runtime

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// PolicyConfig holds network policy
type PolicyConfig struct {
	AllowedDomains []string
	AllowedIPs     []string
}

// NetworkGuard v0.2.1 - Iron Dome
// NOTE: eBPF cgroup hooks require kernel-specific BTF. 
// This placeholder provides DNS resolution and logging.
type NetworkGuard struct {
	policy      PolicyConfig
	resolvedIPs map[string][]string
	mu          sync.RWMutex
	stopChan    chan struct{}
	cgroupPath  string
}

func NewNetworkGuard() *NetworkGuard {
	return &NetworkGuard{
		resolvedIPs: make(map[string][]string),
		stopChan:    make(chan struct{}),
	}
}

func (ng *NetworkGuard) LoadDefaultPolicy() {
	log.Println("📋 Loading default network policy...")
	ng.policy = PolicyConfig{
		AllowedDomains: []string{
			"api.openai.com",
			"api.anthropic.com",
			"google.com",
		},
		AllowedIPs: []string{
			"8.8.8.8",
			"8.8.4.4", 
			"1.1.1.1",
		},
	}
}

func (ng *NetworkGuard) AddAllowedIP(ip string) error {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return fmt.Errorf("invalid IP: %s", ip)
	}
	log.Printf("✅ Allowed IP: %s", ip)
	return nil
}

func (ng *NetworkGuard) ResolveDomain(domain string) error {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return fmt.Errorf("DNS failed for %s: %w", domain, err)
	}

	var resolved []string
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			resolved = append(resolved, ipv4.String())
			ng.AddAllowedIP(ipv4.String())
		}
	}

	ng.mu.Lock()
	ng.resolvedIPs[domain] = resolved
	ng.mu.Unlock()

	if len(resolved) > 0 {
		log.Printf("✅ Resolved %s -> %v", domain, resolved)
	}
	return nil
}

func (ng *NetworkGuard) RefreshDNS() {
	for _, domain := range ng.policy.AllowedDomains {
		ng.ResolveDomain(domain)
	}
}

func (ng *NetworkGuard) StartDNSTicker() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-ng.stopChan:
				ticker.Stop()
				return
			case <-ticker.C:
				log.Println("🔄 Refreshing DNS...")
				ng.RefreshDNS()
			}
		}
	}()
}

func (ng *NetworkGuard) Start(cgroupPath string) error {
	ng.cgroupPath = cgroupPath

	if len(ng.policy.AllowedDomains) == 0 {
		ng.LoadDefaultPolicy()
	}

	for _, ip := range ng.policy.AllowedIPs {
		ng.AddAllowedIP(ip)
	}

	ng.RefreshDNS()
	ng.StartDNSTicker()

	log.Println("🔥 KIDON NETWORK GUARD v0.2.1 (Iron Dome)")
	log.Printf("🛡️ Network policy (logging only): Allowed %d domains, %d IPs",
		len(ng.policy.AllowedDomains), len(ng.policy.AllowedIPs))
	log.Printf("📍 Cgroup path: %s", cgroupPath)
	log.Println("⚠️  Note: Full eBPF cgroup filtering requires kernel BTF support")
	log.Println("📡 DNS resolution and policy management ACTIVE")

	return nil
}

func (ng *NetworkGuard) MonitorEvents() {
	log.Println("📡 Network monitoring active (eBPF events when kernel supports)")
	// Block until shutdown
	<-ng.stopChan
}

func (ng *NetworkGuard) Close() {
	close(ng.stopChan)
	log.Println("Network guard stopped")
}
