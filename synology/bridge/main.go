package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Navidrome integration failed:", err)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) == 4 && os.Args[1] == "delete" {
		return deleteAccount(os.Args[2], os.Args[3])
	}
	if os.Geteuid() == 0 {
		return errors.New("nonroot runtime required")
	}
	peer := os.Getenv("BRIDGE_PEER")
	secret := []byte(os.Getenv("BRIDGE_CONTROL"))
	os.Unsetenv("BRIDGE_CONTROL")
	if peer == "" || len(secret) < 32 {
		return errors.New("gateway peer and private control key required")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	// Override rather than inherit any address that could expose the backend.
	cmd := exec.Command("/app/navidrome", "--address", "127.0.0.1", "--port", "4534")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGTERM}
	if err := cmd.Start(); err != nil {
		return errors.New("upstream start failed")
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	defer func() {
		select {
		case <-exited:
			return
		default:
		}
		syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		select {
		case <-exited:
		case <-time.After(10 * time.Second):
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			<-exited
		}
	}()
	base := "http://127.0.0.1:4534"
	a := newAccounts(base, os.Getenv("ND_DATAFOLDER"))
	ready := false
	for i := 0; i < 120; i++ {
		c := http.Client{Timeout: time.Second}
		res, err := c.Get(base + "/ping")
		if err == nil {
			res.Body.Close()
			if res.StatusCode == 200 {
				ready = true
				break
			}
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
		}
	}
	if !ready {
		return errors.New("upstream readiness timeout")
	}
	if err := a.bootstrap(ctx); err != nil {
		return err
	}
	up, _ := url.Parse(base)
	proxy := upstreamProxy(up)
	b := &bridge{peer: peer, secret: secret, proxy: proxy, syncUser: a.sync, deleteUser: a.delete}
	server := &http.Server{Addr: os.Getenv("BRIDGE_ADDRESS"), Handler: b, ReadHeaderTimeout: 10 * time.Second, MaxHeaderBytes: 64 << 10}
	serving := make(chan error, 1)
	go func() { serving <- server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		stop, c := context.WithTimeout(context.Background(), 10*time.Second)
		defer c()
		server.Shutdown(stop)
		return nil
	case err := <-serving:
		return err
	case <-exited: // Avoid consuming the process result needed by deferred cleanup.
		exited <- errors.New("upstream stopped")
		server.Close()
		return errors.New("upstream stopped")
	}
}
func deleteAccount(dest, uid string) error {
	if n, err := strconv.Atoi(uid); err != nil || n <= 0 {
		return errors.New("invalid account UID")
	}
	var cfg struct {
		Upstream string `json:"upstream"`
	}
	raw, err := os.ReadFile(filepath.Join(dest, "sso/gateway.json"))
	if err != nil || json.Unmarshal(raw, &cfg) != nil {
		return errors.New("gateway configuration unavailable")
	}
	secret, err := os.ReadFile(filepath.Join(dest, "compose/secrets/runtime-bridge-control.secret"))
	if err != nil {
		return errors.New("private control key unavailable")
	}
	stamp := strconv.FormatInt(time.Now().Unix(), 10)
	r, err := http.NewRequest("DELETE", strings.TrimRight(cfg.Upstream, "/")+"/_bridge/delete/"+uid, nil)
	if err != nil {
		return err
	}
	r.Header.Set("X-Bridge-Time", stamp)
	r.Header.Set("X-Bridge-Signature", signature(secret, uid, stamp))
	c := http.Client{Timeout: 20 * time.Second}
	res, err := c.Do(r)
	if err != nil {
		return errors.New("account deletion transport failed")
	}
	defer res.Body.Close()
	io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
	if res.StatusCode != 204 {
		return errors.New("account deletion pending")
	}
	return nil
}

// Navidrome requires an absolute URL at startup. Its generated browser config
// uses the same-origin gateway route, independent of the administrator's NAS name.
func gatewayLogout(r *http.Response) error {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/html") {
		return nil
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, (4<<20)+1))
	r.Body.Close()
	if err != nil || len(body) > 4<<20 {
		return errors.New("upstream HTML exceeds limit")
	}
	body = bytes.ReplaceAll(body, []byte("https://navidrome.invalid/_sso/logout"), []byte("/_dsm/logout"))
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.ContentLength = int64(len(body))
	r.Header.Set("Content-Length", strconv.Itoa(len(body)))
	return nil
}

func upstreamProxy(up *url.URL) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{Rewrite: func(pr *httputil.ProxyRequest) {
		pr.SetURL(up)
		pr.Out.Host = pr.In.Host
		pr.Out.Header.Set("X-Forwarded-Host", pr.In.Host)
		// ReverseProxy removes forwarded headers before Rewrite. Only the
		// verified SDK peer can reach this handler; retain its public scheme.
		for _, name := range []string{"X-Forwarded-Proto", "X-Forwarded-Port", "X-Forwarded-For"} {
			pr.Out.Header.Set(name, pr.In.Header.Get(name))
		}
		pr.Out.Header.Set("Accept-Encoding", "identity")
	}, ModifyResponse: gatewayLogout, ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) { http.Error(w, "upstream unavailable", 502) }}
}
