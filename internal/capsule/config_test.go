package capsule

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDocumentedConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "capsule.json")
	c := Config{
		Version:    1,
		Image:      "docker.io/library/node:22-alpine",
		Install:    "git clone https://github.com/owner/project.git .",
		Run:        "npm install && npm run dev -- --host 127.0.0.1",
		AllowHosts: []string{"registry.npmjs.org", "github.com", "github.com"},
		Ports:      []int{3000, 3000},
	}
	if err := WriteConfig(path, c, false); err != nil {
		t.Fatal(err)
	}
	got, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.AllowHosts, []string{"github.com", "registry.npmjs.org"}) {
		t.Fatalf("hosts = %#v", got.AllowHosts)
	}
	if !reflect.DeepEqual(got.Ports, []int{3000}) {
		t.Fatalf("ports = %#v", got.Ports)
	}
	if err := WriteConfig(path, c, false); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected overwrite refusal, got %v", err)
	}
}

func TestConfigRejectsUnsafeCapabilities(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{"missing install", func(c *Config) { c.Install = "" }, "install command"},
		{"IP allowlist", func(c *Config) { c.AllowHosts = []string{"127.0.0.1"} }, "not an IP"},
		{"uppercase host", func(c *Config) { c.AllowHosts = []string{"GitHub.com"} }, "lowercase"},
		{"bad port", func(c *Config) { c.Ports = []int{70000} }, "outside"},
		{"space in image", func(c *Config) { c.Image = "node:22 --privileged" }, "whitespace"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Config{Version: 1, Image: "node:22", Install: "npm ci", Run: "npm start"}
			tt.edit(&c)
			if err := c.Validate(); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v, want error containing %q", err, tt.want)
			}
		})
	}
}

func TestLoadRejectsUnknownAndTrailingFields(t *testing.T) {
	for _, body := range []string{
		`{"version":1,"image":"node:22","install":"npm ci","run":"npm start","oops":true}`,
		`{"version":1,"image":"node:22","install":"npm ci","run":"npm start"} {}`,
	} {
		path := filepath.Join(t.TempDir(), "capsule.json")
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadConfig(path); err == nil {
			t.Fatalf("expected %q to fail", body)
		}
	}
}
