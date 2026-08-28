package capsule

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type CapabilityReview struct {
	Config     string   `json:"config"`
	Image      string   `json:"image"`
	Filesystem []string `json:"filesystem"`
	Network    []string `json:"network"`
	Ports      []string `json:"ports"`
	Hardening  []string `json:"hardening"`
	Install    string   `json:"install"`
	Run        string   `json:"run"`
}

func Review(path string, c Config) CapabilityReview {
	network := []string{"direct network: denied"}
	for _, host := range c.AllowHosts {
		network = append(network, "HTTPS/HTTP proxy: "+host)
	}
	ports := []string{"host listeners: none"}
	if len(c.Ports) > 0 {
		ports = ports[:0]
		for _, p := range c.Ports {
			ports = append(ports, fmt.Sprintf("127.0.0.1:%d → capsule:%d", p, p))
		}
	}
	return CapabilityReview{
		Config: path,
		Image:  c.Image,
		Filesystem: []string{
			"/workspace: new empty tmpfs",
			"/: read-only container image",
			"host home and project: not mounted",
		},
		Network:   network,
		Ports:     ports,
		Hardening: []string{"rootless engine required", "all capabilities dropped", "no-new-privileges", "PID limit: 256"},
		Install:   c.Install,
		Run:       c.Run,
	}
}

func (r CapabilityReview) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "CAPABILITY REVIEW  %s\n\n", r.Config)
	fmt.Fprintf(&b, "IMAGE\n  %s\n\n", r.Image)
	writeLines(&b, "FILESYSTEM", r.Filesystem)
	writeLines(&b, "NETWORK", r.Network)
	writeLines(&b, "PORTS", r.Ports)
	writeLines(&b, "HARDENING", r.Hardening)
	fmt.Fprintf(&b, "COMMANDS\n  install  %s\n  run      %s\n", r.Install, r.Run)
	return b.String()
}

func writeLines(b *strings.Builder, title string, lines []string) {
	fmt.Fprintln(b, title)
	for _, line := range lines {
		fmt.Fprintln(b, "  "+line)
	}
	fmt.Fprintln(b)
}

func shortHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:12]
}
