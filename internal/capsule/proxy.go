package capsule

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type AllowProxy struct {
	allowed map[string]struct{}
	ln      net.Listener
	wg      sync.WaitGroup
}

func StartAllowProxy(socketPath string, hosts []string) (*AllowProxy, error) {
	_ = os.Remove(socketPath)
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("listen on proxy socket: %w", err)
	}
	if err := os.Chmod(socketPath, 0o666); err != nil {
		ln.Close()
		return nil, fmt.Errorf("set proxy socket permissions: %w", err)
	}
	p := &AllowProxy{allowed: make(map[string]struct{}, len(hosts)), ln: ln}
	for _, host := range hosts {
		p.allowed[strings.ToLower(host)] = struct{}{}
	}
	p.wg.Add(1)
	go p.serve()
	return p, nil
}

func (p *AllowProxy) Close() error {
	err := p.ln.Close()
	p.wg.Wait()
	if errors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}

func (p *AllowProxy) serve() {
	defer p.wg.Done()
	for {
		conn, err := p.ln.Accept()
		if err != nil {
			return
		}
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			defer conn.Close()
			p.handle(conn)
		}()
	}
}

func (p *AllowProxy) handle(client net.Conn) {
	_ = client.SetDeadline(time.Now().Add(2 * time.Minute))
	reader := bufio.NewReader(io.LimitReader(client, 1<<20))
	req, err := http.ReadRequest(reader)
	if err != nil {
		writeProxyError(client, http.StatusBadRequest, "malformed proxy request")
		return
	}
	defer req.Body.Close()

	host, port, err := proxyDestination(req)
	if err != nil {
		writeProxyError(client, http.StatusBadRequest, err.Error())
		return
	}
	if _, ok := p.allowed[host]; !ok {
		writeProxyError(client, http.StatusForbidden, "destination is not in allow_hosts")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	remote, err := dialPublic(ctx, host, port)
	if err != nil {
		writeProxyError(client, http.StatusBadGateway, "approved destination could not be reached")
		return
	}
	defer remote.Close()

	if req.Method == http.MethodConnect {
		_, _ = io.WriteString(client, "HTTP/1.1 200 Connection Established\r\n\r\n")
		copyBoth(client, remote)
		return
	}
	stripProxyHeaders(req.Header)
	req.RequestURI = ""
	req.URL.Scheme = ""
	req.URL.Host = ""
	req.Close = true
	if err := req.Write(remote); err != nil {
		return
	}
	_, _ = io.Copy(client, remote)
}

func proxyDestination(req *http.Request) (string, string, error) {
	if req.Method == http.MethodConnect {
		host, port, err := net.SplitHostPort(req.Host)
		if err != nil || port != "443" {
			return "", "", errors.New("CONNECT is restricted to explicit hostnames on port 443")
		}
		return normalizeProxyHost(host)
	}
	if req.URL == nil || req.URL.Scheme != "http" {
		return "", "", errors.New("plain proxy requests must use http:// URLs")
	}
	host := req.URL.Hostname()
	port := req.URL.Port()
	if port == "" {
		port = "80"
	}
	if port != "80" {
		return "", "", errors.New("plain HTTP is restricted to port 80")
	}
	normal, _, err := normalizeProxyHost(host)
	return normal, port, err
}

func normalizeProxyHost(host string) (string, string, error) {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	if net.ParseIP(host) != nil || !hostnamePattern.MatchString(host) {
		return "", "", errors.New("destination must be an allowed hostname, not an IP address")
	}
	return host, "443", nil
}

func dialPublic(ctx context.Context, host, port string) (net.Conn, error) {
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	dialer := net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	for _, addr := range addrs {
		if !isPublicIP(addr.IP) {
			continue
		}
		conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(addr.IP.String(), port))
		if err == nil {
			return conn, nil
		}
	}
	return nil, errors.New("hostname has no reachable public address")
}

func isPublicIP(ip net.IP) bool {
	return ip != nil && !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() &&
		!ip.IsLinkLocalMulticast() && !ip.IsMulticast() && !ip.IsUnspecified()
}

func stripProxyHeaders(h http.Header) {
	for _, name := range []string{"Proxy-Authorization", "Proxy-Connection", "Connection", "Keep-Alive", "TE", "Trailer", "Transfer-Encoding", "Upgrade"} {
		h.Del(name)
	}
}

func writeProxyError(w io.Writer, status int, message string) {
	body := message + "\n"
	fmt.Fprintf(w, "HTTP/1.1 %d %s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s", status, http.StatusText(status), len(body), body)
}

func copyBoth(a, b net.Conn) {
	done := make(chan struct{}, 2)
	go func() { _, _ = io.Copy(a, b); done <- struct{}{} }()
	go func() { _, _ = io.Copy(b, a); done <- struct{}{} }()
	<-done
}

func parseProxyURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "http" || u.Host == "" {
		return nil, errors.New("invalid HTTP proxy URL")
	}
	return u, nil
}
