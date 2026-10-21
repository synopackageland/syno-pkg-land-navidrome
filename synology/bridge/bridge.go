package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const bootstrapName = "__dsm_bridge_internal__"

var identityHeaders = []string{"Remote-User", "Remote-Uid", "Remote-Groups", "Remote-Name", "Remote-Email"}

type bridge struct {
	peer       string
	secret     []byte
	proxy      http.Handler
	syncUser   func(context.Context, int, string, bool) (string, error)
	deleteUser func(context.Context, int) error
}

func bypass(p string) bool {
	return p == "/rest" || strings.HasPrefix(p, "/rest/") || p == "/share" || strings.HasPrefix(p, "/share/")
}
func clean(r *http.Request) bool {
	p := r.URL.Path
	if strings.ContainsAny(p, "\\\x00") || strings.Contains(p, "//") {
		return false
	}
	for _, s := range strings.Split(p, "/") {
		if s == "." || s == ".." {
			return false
		}
	}
	// Reject encoded separators and nested path escapes; upstream and gateway must agree.
	escaped := strings.ToLower(r.URL.EscapedPath())
	return !strings.Contains(escaped, "%2f") && !strings.Contains(escaped, "%5c") && !strings.Contains(escaped, "%25")
}
func (b *bridge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	peer, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || peer != b.peer {
		http.Error(w, "gateway required", 403)
		return
	}
	if !clean(r) {
		http.Error(w, "invalid path", 400)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/_bridge/") {
		b.control(w, r)
		return
	}
	if bypass(r.URL.Path) {
		for _, h := range identityHeaders {
			r.Header.Del(h)
		}
		b.proxy.ServeHTTP(w, r)
		return
	}
	if r.URL.Path == "/ping" && r.Method == "GET" && r.Header.Get("Remote-User") == "" {
		w.WriteHeader(200)
		return
	}
	name := r.Header.Get("Remote-User")
	uid, err := strconv.Atoi(r.Header.Get("Remote-Uid"))
	group := r.Header.Get("Remote-Groups")
	if err != nil || uid <= 0 || name == "" || strings.EqualFold(name, bootstrapName) || len(name) > 256 || strings.ContainsAny(name, "\r\n\x00/") || (group != "admin" && group != "user") {
		http.Error(w, "DSM identity required", 403)
		return
	}
	id, err := b.syncUser(r.Context(), uid, name, group == "admin")
	if err != nil {
		http.Error(w, "account synchronization unavailable", 503)
		return
	}
	if r.URL.Path == "/_dsm/logout" && r.Method == http.MethodGet {
		nonce := randomID()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'nonce-"+nonce+"'; connect-src 'self'; form-action 'self'; base-uri 'none'; frame-ancestors 'none'")
		fmt.Fprintf(w, `<!doctype html><meta charset="utf-8"><title>Sign out</title><form method="post" action="/_sso/logout"><button type="submit">Sign out</button></form><script nonce="%s">const form=document.forms[0];async function signOut(){try{const r=await fetch("/_sso/logout",{method:"POST",credentials:"same-origin",redirect:"manual"});if(r.type!=="opaqueredirect"&&!r.ok)throw new Error();location.replace("/")}catch{document.getElementById("status").textContent="Sign out failed. Try again."}}form.addEventListener("submit",e=>{e.preventDefault();signOut()});signOut()</script><p id="status" role="status"></p>`, nonce)
		return
	}
	// Native browser JWTs must never choose a different user than the DSM gateway.
	r.Header.Del("X-ND-Authorization")
	r.Header.Del("Cookie")
	q := r.URL.Query()
	q.Del("jwt")
	r.URL.RawQuery = q.Encode()
	// Account roles belong to DSM. Keep password/preferences editing for the owner.
	if r.URL.Path == "/auth/createAdmin" || r.URL.Path == "/auth/login" ||
		strings.HasPrefix(r.URL.Path, "/api/user") && r.Method != "GET" && (r.Method != "PUT" || strings.TrimSuffix(r.URL.Path, "/") != "/api/user/"+url.PathEscape(id)) {
		http.Error(w, "accounts are managed by DSM", 403)
		return
	}
	b.proxy.ServeHTTP(w, r)
}
func signature(secret []byte, uid, stamp string) string {
	mac := hmac.New(sha256.New, secret)
	fmt.Fprintf(mac, "delete\n%s\n%s", uid, stamp)
	return hex.EncodeToString(mac.Sum(nil))
}
func (b *bridge) control(w http.ResponseWriter, r *http.Request) {
	uidText := strings.TrimPrefix(r.URL.Path, "/_bridge/delete/")
	uid, err := strconv.Atoi(uidText)
	stamp := r.Header.Get("X-Bridge-Time")
	ts, e := strconv.ParseInt(stamp, 10, 64)
	signed, _ := hex.DecodeString(r.Header.Get("X-Bridge-Signature"))
	want, _ := hex.DecodeString(signature(b.secret, uidText, stamp))
	delta := time.Now().Unix() - ts
	if r.Method != "DELETE" || err != nil || uid <= 0 || e != nil || delta < -30 || delta > 30 || len(b.secret) < 32 || !hmac.Equal(signed, want) || r.Header.Get("Remote-User") != "" || r.Header.Get("Remote-Uid") != "" {
		http.Error(w, "invalid control request", 403)
		return
	}
	if err := b.deleteUser(r.Context(), uid); err != nil {
		http.Error(w, "account deletion pending", 503)
		return
	}
	w.WriteHeader(204)
}
