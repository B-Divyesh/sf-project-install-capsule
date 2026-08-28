package capsule

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Bridge runs inside the network-denied container. It exposes the host's Unix
// allowlist proxy on loopback and exposes requested container ports as Unix
// sockets in the ephemeral shared bridge directory.
func Bridge(ctx context.Context, outboundSocket string, forwards []string, command []string) error {
	if outboundSocket == "" {
		return errors.New("bridge socket is required")
	}
	if len(command) == 0 {
		return errors.New("bridge command is required")
	}

	var listeners []net.Listener
	proxy, err := net.Listen("tcp", "127.0.0.1:3128")
	if err != nil {
		return fmt.Errorf("start container proxy bridge: %w", err)
	}
	listeners = append(listeners, proxy)
	go acceptForward(proxy, func() (net.Conn, error) { return net.Dial("unix", outboundSocket) })

	for _, spec := range forwards {
		port, socket, err := parseForward(spec)
		if err != nil {
			closeListeners(listeners)
			return err
		}
		_ = os.Remove(socket)
		ln, err := net.Listen("unix", socket)
		if err != nil {
			closeListeners(listeners)
			return fmt.Errorf("listen for port %d: %w", port, err)
		}
		if err := os.Chmod(socket, 0o666); err != nil {
			closeListeners(append(listeners, ln))
			return err
		}
		listeners = append(listeners, ln)
		address := fmt.Sprintf("127.0.0.1:%d", port)
		go acceptForward(ln, func() (net.Conn, error) { return net.DialTimeout("tcp", address, 5*time.Second) })
	}

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = append(withoutProxyEnvironment(os.Environ()),
		"HTTP_PROXY=http://127.0.0.1:3128", "HTTPS_PROXY=http://127.0.0.1:3128",
		"http_proxy=http://127.0.0.1:3128", "https_proxy=http://127.0.0.1:3128",
		"ALL_PROXY=", "all_proxy=", "NO_PROXY=127.0.0.1,localhost", "no_proxy=127.0.0.1,localhost",
		"HOME=/workspace", "XDG_CACHE_HOME=/workspace/.cache",
	)
	err = cmd.Run()
	closeListeners(listeners)
	return err
}

func withoutProxyEnvironment(environment []string) []string {
	blocked := map[string]struct{}{
		"HTTP_PROXY": {}, "HTTPS_PROXY": {}, "ALL_PROXY": {}, "NO_PROXY": {},
		"http_proxy": {}, "https_proxy": {}, "all_proxy": {}, "no_proxy": {},
		"HOME": {}, "XDG_CACHE_HOME": {},
	}
	filtered := make([]string, 0, len(environment))
	for _, item := range environment {
		key, _, found := strings.Cut(item, "=")
		if _, remove := blocked[key]; found && remove {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func parseForward(value string) (int, string, error) {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 || parts[1] == "" {
		return 0, "", fmt.Errorf("invalid bridge forward %q", value)
	}
	port, err := strconv.Atoi(parts[0])
	if err != nil || port < 1 || port > 65535 {
		return 0, "", fmt.Errorf("invalid bridge port %q", parts[0])
	}
	return port, parts[1], nil
}

func acceptForward(ln net.Listener, dial func() (net.Conn, error)) {
	for {
		incoming, err := ln.Accept()
		if err != nil {
			return
		}
		go func() {
			defer incoming.Close()
			outgoing, err := dial()
			if err != nil {
				return
			}
			defer outgoing.Close()
			copyBoth(incoming, outgoing)
		}()
	}
}

func closeListeners(listeners []net.Listener) {
	for _, ln := range listeners {
		_ = ln.Close()
	}
}

type PortPublisher struct {
	listeners []net.Listener
	wg        sync.WaitGroup
}

func StartPortPublishers(runtimeDir string, ports []int) (*PortPublisher, error) {
	p := &PortPublisher{}
	for _, port := range ports {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			p.Close()
			return nil, fmt.Errorf("approved port %d is unavailable on loopback: %w", port, err)
		}
		p.listeners = append(p.listeners, ln)
		socket := PortSocket(runtimeDir, port)
		p.wg.Add(1)
		go func(listener net.Listener, path string) {
			defer p.wg.Done()
			acceptForward(listener, func() (net.Conn, error) {
				deadline := time.Now().Add(10 * time.Second)
				for {
					conn, err := net.DialTimeout("unix", path, 500*time.Millisecond)
					if err == nil || time.Now().After(deadline) {
						return conn, err
					}
					time.Sleep(75 * time.Millisecond)
				}
			})
		}(ln, socket)
	}
	return p, nil
}

func (p *PortPublisher) Close() {
	for _, ln := range p.listeners {
		_ = ln.Close()
	}
	p.wg.Wait()
}

func PortSocket(runtimeDir string, port int) string {
	return fmt.Sprintf("%s/port-%d.sock", strings.TrimSuffix(runtimeDir, "/"), port)
}

func waitGroupCopy(dst io.Writer, src io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	_, _ = io.Copy(dst, src)
}
