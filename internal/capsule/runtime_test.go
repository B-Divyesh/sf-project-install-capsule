package capsule

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeArgsHaveNoAmbientAccess(t *testing.T) {
	c := Config{Version: 1, Image: "node:22-alpine", Install: "npm ci", Run: "npm start", AllowHosts: []string{"registry.npmjs.org"}, Ports: []int{3000}}
	args := RuntimeArgs(c, "capsule-test", "/tmp/capsule-test", "/usr/bin/capsule")
	joined := strings.Join(args, "\x00")
	for _, required := range []string{"--network=none", "--read-only", "--cap-drop=ALL", "--security-opt=no-new-privileges", "/workspace:rw,nosuid,nodev,exec,size=1g"} {
		if !strings.Contains(joined, required) {
			t.Errorf("runtime args missing %q", required)
		}
	}
	for _, forbidden := range []string{"--privileged", "/var/run/docker.sock", "/home/", "--network=host"} {
		if strings.Contains(joined, forbidden) {
			t.Errorf("runtime args unexpectedly contain %q", forbidden)
		}
	}
	if strings.Contains(joined, "registry.npmjs.org") {
		t.Error("allowlisted hostname must stay in host proxy, not engine arguments")
	}
}

func TestProxyRejectsUndeclaredHostBeforeDNS(t *testing.T) {
	socket := filepath.Join(t.TempDir(), "proxy.sock")
	proxy, err := StartAllowProxy(socket, []string{"github.com"})
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()

	transport := &http.Transport{Proxy: func(*http.Request) (*url.URL, error) { return url.Parse("http://unix.invalid") }, DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socket)
	}}
	client := &http.Client{Transport: transport}
	resp, err := client.Get("http://not-approved.invalid/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestDestinationRules(t *testing.T) {
	if isPublicIP(net.ParseIP("127.0.0.1")) || isPublicIP(net.ParseIP("10.0.0.2")) || isPublicIP(net.ParseIP("169.254.169.254")) {
		t.Fatal("private or metadata address considered public")
	}
	if !isPublicIP(net.ParseIP("8.8.8.8")) {
		t.Fatal("public address rejected")
	}
	req, _ := http.NewRequest(http.MethodConnect, "https://example.com:80", nil)
	req.Host = "example.com:80"
	if _, _, err := proxyDestination(req); err == nil {
		t.Fatal("CONNECT to port 80 should be rejected")
	}
}
