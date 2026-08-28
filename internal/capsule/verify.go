package capsule

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type VerifyResult struct {
	RootlessEngine       bool `json:"rootless_engine"`
	HostSecretUnreadable bool `json:"host_secret_unreadable"`
	DirectNetworkDenied  bool `json:"direct_network_denied"`
	UndeclaredHostDenied bool `json:"undeclared_host_denied"`
}

func (v VerifyResult) Passed() bool {
	return v.RootlessEngine && v.HostSecretUnreadable && v.DirectNetworkDenied && v.UndeclaredHostDenied
}

// VerifyIsolation runs active checks inside the configured image. It creates a
// short-lived sentinel in the caller's home and verifies that the capsule
// cannot see it or establish direct/undeclared network connections.
func VerifyIsolation(ctx context.Context, c Config, configPath string) (VerifyResult, error) {
	engine, err := FindEngine(ctx, true)
	if err != nil {
		return VerifyResult{}, err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return VerifyResult{}, fmt.Errorf("locate home for sentinel: %w", err)
	}
	var nonce [8]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return VerifyResult{}, err
	}
	sentinel := filepath.Join(home, ".capsule-verify-"+hex.EncodeToString(nonce[:]))
	if err := os.WriteFile(sentinel, []byte("must not cross capsule boundary\n"), 0o600); err != nil {
		return VerifyResult{}, fmt.Errorf("write temporary verification sentinel: %w", err)
	}
	defer os.Remove(sentinel)

	executable, err := os.Executable()
	if err != nil {
		return VerifyResult{}, err
	}
	executable, _ = filepath.EvalSymlinks(executable)
	name := "capsule-verify-" + ConfigID(configPath)
	runtimeDir, err := os.MkdirTemp("", "capsule-verify-")
	if err != nil {
		return VerifyResult{}, err
	}
	defer os.RemoveAll(runtimeDir)
	if err := os.Chmod(runtimeDir, 0o711); err != nil {
		return VerifyResult{}, err
	}
	proxy, err := StartAllowProxy(filepath.Join(runtimeDir, "outbound.sock"), c.AllowHosts)
	if err != nil {
		return VerifyResult{}, err
	}
	defer proxy.Close()

	denyHost := hex.EncodeToString(nonce[:]) + ".capsule.invalid"
	args := baseRuntimeArgs(c.Image, name, runtimeDir, executable)
	args = append(args, "/capsule", "bridge", "--socket", "/run/capsule/outbound.sock", "--", "/capsule", "self-test", "--host-secret", sentinel, "--deny-host", denyHost)
	cmd := exec.CommandContext(ctx, engine.Path, args...)
	out, err := cmd.CombinedOutput()
	_, _ = exec.CommandContext(context.Background(), engine.Path, "rm", "-f", name).CombinedOutput()
	if err != nil {
		return VerifyResult{}, fmt.Errorf("isolation probe failed: %s", compactOutput(out, err))
	}
	var result VerifyResult
	if err := json.Unmarshal(out, &result); err != nil {
		return VerifyResult{}, fmt.Errorf("decode isolation probe: %w (%s)", err, string(out))
	}
	result.RootlessEngine = true
	if !result.Passed() {
		return result, errors.New("one or more live isolation checks failed")
	}
	return result, nil
}

func SelfTest(hostSecret, denyHost string) (VerifyResult, error) {
	result := VerifyResult{}
	_, err := os.Stat(hostSecret)
	result.HostSecretUnreadable = errors.Is(err, os.ErrNotExist)
	direct, err := net.DialTimeout("tcp", "1.1.1.1:443", 1500*time.Millisecond)
	if err == nil {
		direct.Close()
	} else {
		result.DirectNetworkDenied = true
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://" + denyHost + "/")
	if err == nil {
		result.UndeclaredHostDenied = resp.StatusCode == http.StatusForbidden
		resp.Body.Close()
	}
	if !result.HostSecretUnreadable || !result.DirectNetworkDenied || !result.UndeclaredHostDenied {
		return result, errors.New("isolation boundary did not hold")
	}
	return result, nil
}
