// Package capsule implements the inspectable configuration and isolation
// plumbing used by the capsule command.
package capsule

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	Version           = "0.1.0"
	DefaultConfigPath = "capsule.json"
)

var hostnamePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*$`)

// Config is the complete, reviewable capsule capability declaration.
type Config struct {
	Version    int      `json:"version"`
	Image      string   `json:"image"`
	Install    string   `json:"install"`
	Run        string   `json:"run"`
	AllowHosts []string `json:"allow_hosts"`
	Ports      []int    `json:"ports"`
}

func DefaultConfig() Config {
	return Config{Version: 1, Image: "docker.io/library/node:22-bookworm", AllowHosts: []string{}, Ports: []int{}}
}

func LoadConfig(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, fmt.Errorf("%s does not exist; run `capsule init` first", path)
		}
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}
	var c Config
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return Config{}, fmt.Errorf("parse %s: trailing JSON value", path)
	}
	if err := c.Validate(); err != nil {
		return Config{}, err
	}
	return c, nil
}

func WriteConfig(path string, c Config, overwrite bool) error {
	if err := c.NormalizeAndValidate(); err != nil {
		return err
	}
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists; pass --force to replace it", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect %s: %w", path, err)
		}
	}
	b, _ := json.MarshalIndent(c, "", "  ")
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func (c *Config) NormalizeAndValidate() error {
	for i := range c.AllowHosts {
		c.AllowHosts[i] = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(c.AllowHosts[i]), "."))
	}
	sort.Strings(c.AllowHosts)
	c.AllowHosts = uniqueStrings(c.AllowHosts)
	sort.Ints(c.Ports)
	c.Ports = uniqueInts(c.Ports)
	return c.Validate()
}

func (c Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported config version %d; expected 1", c.Version)
	}
	if strings.TrimSpace(c.Image) == "" || strings.ContainsAny(c.Image, " \t\r\n") {
		return errors.New("image must be a non-empty OCI image reference without whitespace")
	}
	if strings.TrimSpace(c.Install) == "" {
		return errors.New("install command is required")
	}
	if strings.TrimSpace(c.Run) == "" {
		return errors.New("run command is required")
	}
	for _, host := range c.AllowHosts {
		if net.ParseIP(host) != nil || !hostnamePattern.MatchString(host) || strings.ToLower(host) != host {
			return fmt.Errorf("allow_hosts entry %q must be a lowercase ASCII hostname, not an IP address", host)
		}
	}
	for _, port := range c.Ports {
		if port < 1 || port > 65535 {
			return fmt.Errorf("port %d is outside 1..65535", port)
		}
	}
	return nil
}

func ConfigID(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	return shortHash(abs)
}

func uniqueStrings(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	out := in[:1]
	for _, v := range in[1:] {
		if v != out[len(out)-1] {
			out = append(out, v)
		}
	}
	return out
}

func uniqueInts(in []int) []int {
	if len(in) == 0 {
		return []int{}
	}
	out := in[:1]
	for _, v := range in[1:] {
		if v != out[len(out)-1] {
			out = append(out, v)
		}
	}
	return out
}
