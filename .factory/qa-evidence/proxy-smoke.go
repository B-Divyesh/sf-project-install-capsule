package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/B-Divyesh/sf-project-install-capsule/internal/capsule"
)

func main() {
	dir, err := os.MkdirTemp("", "capsule-proxy-smoke-")
	if err != nil { panic(err) }
	defer os.RemoveAll(dir)
	socket := filepath.Join(dir, "proxy.sock")
	p, err := capsule.StartAllowProxy(socket, []string{"example.com"})
	if err != nil { panic(err) }
	defer p.Close()
	transport := &http.Transport{
		Proxy: func(*http.Request) (*url.URL, error) { return url.Parse("http://unix.invalid") },
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socket)
		},
	}
	client := &http.Client{Transport: transport}
	for _, target := range []string{"http://example.com/", "http://example.org/", "http://127.0.0.1/"} {
		resp, err := client.Get(target)
		if err != nil { fmt.Printf("target=%s error=%q\n", target, err); continue }
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		fmt.Printf("target=%s status=%d\n", target, resp.StatusCode)
	}
	counts := map[int]int{}
	retryAfter := 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := client.Get("http://example.org/")
			if err != nil { return }
			resp.Body.Close()
			mu.Lock(); counts[resp.StatusCode]++; if resp.Header.Get("Retry-After") != "" { retryAfter++ }; mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("denied_burst=%v retry_after_responses=%d\n", counts, retryAfter)
}
