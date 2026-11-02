package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestFirstRegularUserAndRoleChanges(t *testing.T) {
	var role bool
	var username string
	calls := 0
	b := &bridge{peer: "192.0.2.1", syncUser: func(_ context.Context, uid int, name string, admin bool) (string, error) {
		if uid != 1024 || name != "listener" {
			t.Fatal("identity not verified")
		}
		role, username = admin, name
		calls++
		return "u1", nil
	}}
	b.proxy = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Remote-User") != username || r.Header.Get("X-ND-Authorization") != "" || r.Header.Get("Cookie") != "" || r.URL.Query().Get("jwt") != "" {
			t.Fatal("native token overrides DSM identity")
		}
		w.WriteHeader(204)
	})
	for _, admin := range []bool{false, true, false} {
		r := httptest.NewRequest("GET", "http://music.example/synopkgland/SynoPkgNavidrome/app/?jwt=forged", nil)
		r.RemoteAddr = "192.0.2.1:4000"
		r.Header.Set("Remote-User", "listener")
		r.Header.Set("Remote-Uid", "1024")
		group := "user"
		if admin {
			group = "admin"
		}
		r.Header.Set("Remote-Groups", group)
		r.Header.Set("X-ND-Authorization", "Bearer stale-admin")
		r.Header.Set("Cookie", "jwt=stale-admin")
		w := httptest.NewRecorder()
		b.ServeHTTP(w, r)
		if w.Code != 204 || role != admin {
			t.Fatalf("role/status: %v %d", role, w.Code)
		}
	}
	if calls != 3 {
		t.Fatal("role must be resolved on every protected request")
	}
}

func TestBypassNeverCarriesIdentity(t *testing.T) {
	b := &bridge{peer: "192.0.2.1", syncUser: func(context.Context, int, string, bool) (string, error) {
		t.Fatal("bypass provisioned a user")
		return "", nil
	}}
	b.proxy = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, h := range []string{"Remote-User", "Remote-Uid", "Remote-Groups", "Remote-Name", "Remote-Email"} {
			if r.Header.Get(h) != "" {
				t.Fatalf("leaked %s", h)
			}
		}
		if r.URL.Query().Get("t") != "native-token" {
			t.Fatal("lost native credential")
		}
		w.WriteHeader(204)
	})
	for _, path := range []string{"/rest/ping.view?t=native-token", "/share/public?t=native-token"} {
		r := httptest.NewRequest("GET", "http://music.example"+basePath+path, nil)
		r.RemoteAddr = "192.0.2.1:1"
		for _, h := range []string{"Remote-User", "Remote-Uid", "Remote-Groups", "Remote-Name", "Remote-Email"} {
			r.Header.Set(h, "forged")
		}
		w := httptest.NewRecorder()
		b.ServeHTTP(w, r)
		if w.Code != 204 {
			t.Fatal(w.Code)
		}
	}
}

func TestForgedPeerAndMissingIdentityFailClosed(t *testing.T) {
	b := &bridge{peer: "192.0.2.1", proxy: http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("reached upstream") })}
	for _, tc := range []struct{ peer, path string }{{"192.0.2.2:1", "/app/"}, {"192.0.2.2:1", "/rest/ping"}, {"192.0.2.1:1", "/app/"}, {"192.0.2.1:1", "/rest-evil"}, {"192.0.2.1:1", "/rest/../api/user"}, {"192.0.2.1:1", "/rest%2fping"}} {
		r := httptest.NewRequest("GET", "http://music.example"+basePath+tc.path, nil)
		r.RemoteAddr = tc.peer
		w := httptest.NewRecorder()
		b.ServeHTTP(w, r)
		if w.Code < 400 {
			t.Fatal(tc)
		}
	}
}

func TestProvisionBeforeForwardWithUpstreamAPI(t *testing.T) {
	users := []user{{"id": "bootstrap", "userName": bootstrapName, "isAdmin": true}}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Remote-User") != bootstrapName {
			t.Fatal("no private bootstrap identity")
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "GET":
			json.NewEncoder(w).Encode(users)
		case r.Method == "POST":
			var u user
			json.NewDecoder(r.Body).Decode(&u)
			if u["isAdmin"] != false || !strings.HasPrefix(u["password"].(string), "__NAVIDROME_AUTOGEN__") {
				t.Fatal("first regular user unsafe")
			}

			delete(u, "password")
			users = append(users, u)
			json.NewEncoder(w).Encode(u)
		case r.Method == "PUT":
			var u user
			json.NewDecoder(r.Body).Decode(&u)
			users[1] = u
			json.NewEncoder(w).Encode(u)
		default:
			t.Fatal("unexpected request")
		}
	}))
	defer s.Close()
	a := newAccounts(s.URL, t.TempDir())
	id, err := a.sync(context.Background(), 1024, "listener", false)
	if err != nil || id == "" || users[1]["isAdmin"] != false {
		t.Fatalf("%s %v", id, err)
	}
	if _, err = a.sync(context.Background(), 1024, "listener", true); err != nil || users[1]["isAdmin"] != true {
		t.Fatal(err)
	}
	if _, err = a.sync(context.Background(), 1024, "listener", false); err != nil || users[1]["isAdmin"] != false {
		t.Fatal(err)
	}
	if _, err = a.sync(context.Background(), 1025, "LISTENER", false); err == nil {
		t.Fatal("case insensitive identity collision")
	}
}

func TestDeleteControlRequiresPrivateSignature(t *testing.T) {
	secret := []byte(strings.Repeat("a", 64))
	deleted := 0
	b := &bridge{peer: "192.0.2.1", secret: secret, deleteUser: func(_ context.Context, uid int) error {
		if uid != 1024 {
			t.Fatal(uid)
		}
		deleted++
		return nil
	}}
	for _, valid := range []bool{false, true} {
		r := httptest.NewRequest("DELETE", "http://music.example/_bridge/delete/1024", nil)
		r.RemoteAddr = "192.0.2.1:1"
		stamp := strconv.FormatInt(time.Now().Unix(), 10)
		r.Header.Set("X-Bridge-Time", stamp)
		if valid {
			r.Header.Set("X-Bridge-Signature", signature(secret, "1024", stamp))
		}
		w := httptest.NewRecorder()
		b.ServeHTTP(w, r)
		if valid && w.Code != 204 || !valid && w.Code != 403 {
			t.Fatal(w.Code)
		}
	}
	if deleted != 1 {
		t.Fatal(deleted)
	}
}

func TestLogoutUsesGatewayPost(t *testing.T) {
	b := &bridge{peer: "192.0.2.1", syncUser: func(context.Context, int, string, bool) (string, error) { return "u1", nil }}
	r := httptest.NewRequest("GET", "https://music.example/synopkgland/SynoPkgNavidrome/_dsm/logout", nil)
	r.RemoteAddr = "192.0.2.1:1"
	r.Header.Set("Remote-User", "listener")
	r.Header.Set("Remote-Uid", "1024")
	r.Header.Set("Remote-Groups", "user")
	w := httptest.NewRecorder()
	b.ServeHTTP(w, r)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `fetch("/synopkgland/SynoPkgNavidrome/_sso/logout"`) || !strings.Contains(w.Body.String(), `redirect:"manual"`) || !strings.Contains(w.Header().Get("Content-Security-Policy"), "connect-src 'self'") {
		t.Fatal("logout does not use SDK's same-origin POST")
	}
}

func TestShareLinkRetainsVerifiedHTTPSOrigin(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != basePath+"/rest/createShare" {
			t.Errorf("native URI prefix lost: %s", r.URL.Path)
		}
		if r.Header.Get("X-Forwarded-Proto") != "https" || r.Header.Get("X-Forwarded-Host") != "music.example:5001" || r.Header.Get("X-Forwarded-Port") != "5001" {
			t.Errorf("public origin lost: proto=%q host=%q", r.Header.Get("X-Forwarded-Proto"), r.Header.Get("X-Forwarded-Host"))
		}
	}))
	defer upstream.Close()
	target, _ := url.Parse(upstream.URL)
	req := httptest.NewRequest("GET", "http://music.example:5001/synopkgland/SynoPkgNavidrome/rest/createShare", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Port", "5001")
	upstreamProxy(target).ServeHTTP(httptest.NewRecorder(), req)
}

func TestPrefixedAccountPolicy(t *testing.T) {
	b := &bridge{peer: "192.0.2.1", syncUser: func(context.Context, int, string, bool) (string, error) { return "u1", nil }, proxy: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })}
	for _, tc := range []struct {
		method, path string
		want         int
	}{
		{"POST", "/auth/createAdmin", 403}, {"POST", "/auth/login", 403},
		{"POST", "/api/user", 403}, {"DELETE", "/api/user/u2", 403},
		{"PUT", "/api/user/u2", 403}, {"PUT", "/api/user/u1", 204}, {"GET", "/api/user", 204},
	} {
		r := httptest.NewRequest(tc.method, "https://music.example"+basePath+tc.path, nil)
		r.RemoteAddr = "192.0.2.1:1"
		r.Header.Set("Remote-User", "listener")
		r.Header.Set("Remote-Uid", "1024")
		r.Header.Set("Remote-Groups", "user")
		w := httptest.NewRecorder()
		b.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Errorf("%s %s: %d", tc.method, tc.path, w.Code)
		}
	}
}

func TestPrefixBoundaryAndHealth(t *testing.T) {
	b := &bridge{peer: "192.0.2.1", proxy: http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("unexpected upstream access") })}
	for _, path := range []string{"/rest/ping.view", "/share/public", basePath + "-other/rest/ping.view", basePath + "/rest/../api/user", basePath + "/rest%2fping"} {
		r := httptest.NewRequest("GET", "https://music.example"+path, nil)
		r.RemoteAddr = "192.0.2.1:1"
		w := httptest.NewRecorder()
		b.ServeHTTP(w, r)
		if w.Code < 400 {
			t.Errorf("accepted %s", path)
		}
	}
	r := httptest.NewRequest("GET", "https://music.example/ping", nil)
	r.RemoteAddr = "192.0.2.1:1"
	w := httptest.NewRecorder()
	b.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
}

func TestBootstrapUsesNativeBasePath(t *testing.T) {
	seen := []string{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.URL.Path)
		if r.URL.Path == basePath+"/app/" {
			w.WriteHeader(200)
			return
		}
		if r.URL.Path == basePath+"/api/user" {
			json.NewEncoder(w).Encode([]user{{"id": "bootstrap", "userName": bootstrapName, "isAdmin": true}})
			return
		}
		t.Errorf("unexpected bootstrap path %s", r.URL.Path)
		w.WriteHeader(404)
	}))
	defer s.Close()
	if err := newAccounts(s.URL+basePath, t.TempDir()).bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 2 {
		t.Fatal(seen)
	}
}

func TestLogoutConfigUsesPackageNamespace(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<script>window.__APP_CONFIG__={"extAuthLogoutURL":"https://navidrome.invalid/_sso/logout"}</script>`))
	}))
	defer s.Close()
	up, _ := url.Parse(s.URL)
	w := httptest.NewRecorder()
	upstreamProxy(up).ServeHTTP(w, httptest.NewRequest("GET", "https://music.example"+basePath+"/app/", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), basePath+"/_dsm/logout") || strings.Contains(w.Body.String(), "navidrome.invalid") {
		t.Fatal("logout escaped package namespace")
	}
}
