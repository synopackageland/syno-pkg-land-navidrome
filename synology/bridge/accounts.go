package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type user map[string]any
type binding struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type accounts struct {
	mu        sync.Mutex
	base, dir string
	client    *http.Client
	bindings  map[string]binding
}

func newAccounts(base, dir string) *accounts {
	return &accounts{base: base, dir: dir, client: &http.Client{Timeout: 10 * time.Second}, bindings: map[string]binding{}}
}
func randomID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("random source unavailable")
	}
	return hex.EncodeToString(b)
}
func (a *accounts) request(ctx context.Context, method, path string, body any, out any) error {
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	r, err := http.NewRequestWithContext(ctx, method, a.base+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	r.Header.Set("Remote-User", bootstrapName)
	r.Header.Set("Content-Type", "application/json")
	res, err := a.client.Do(r)
	if err != nil {
		return errors.New("upstream transport failed")
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("upstream account operation failed (%d)", res.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(io.LimitReader(res.Body, 4<<20)).Decode(out)
	}
	_, err = io.Copy(io.Discard, io.LimitReader(res.Body, 4<<20))
	return err
}
func (a *accounts) load() error {
	p := filepath.Join(a.dir, "bridge-accounts.json")
	st, err := os.Lstat(p)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !st.Mode().IsRegular() || st.Mode().Perm()&0077 != 0 || st.Size() > 4<<20 {
		return errors.New("invalid account state")
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &a.bindings)
}
func (a *accounts) save() error {
	data, err := json.Marshal(a.bindings)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(a.dir, ".bridge-accounts-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(f.Name(), filepath.Join(a.dir, "bridge-accounts.json"))
}
func (a *accounts) bootstrap(ctx context.Context) error {
	// Only this container can reach Navidrome's loopback listener. The upstream
	// header flow creates its first administrator before we expose any ingress.
	req, err := http.NewRequestWithContext(ctx, "GET", a.base+"/app/", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Remote-User", bootstrapName)
	res, err := a.client.Do(req)
	if err != nil {
		return errors.New("bootstrap transport failed")
	}
	io.Copy(io.Discard, io.LimitReader(res.Body, 4<<20))
	res.Body.Close()
	var users []user
	if err = a.request(ctx, "GET", "/api/user", nil, &users); err != nil {
		return err
	}
	for _, u := range users {
		if u["userName"] == bootstrapName && u["isAdmin"] == true {
			return a.load()
		}
	}
	return errors.New("bootstrap administrator unavailable")
}
func (a *accounts) sync(ctx context.Context, uid int, name string, admin bool) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	key := strconv.Itoa(uid)
	var users []user
	if err := a.request(ctx, "GET", "/api/user", nil, &users); err != nil {
		return "", err
	}
	bind, known := a.bindings[key]
	for other, b := range a.bindings {
		if other != key && strings.EqualFold(b.Name, name) {
			return "", errors.New("DSM identity name collision")
		}
	}
	var found user
	for _, u := range users {
		if known && u["id"] == bind.ID {
			found = u
		}
		if strings.EqualFold(fmt.Sprint(u["userName"]), name) && (!known || u["id"] != bind.ID) {
			return "", errors.New("unbound native account collision")
		}
	}
	if !known {
		bind = binding{ID: randomID(), Name: name}
		a.bindings[key] = bind
		if err := a.save(); err != nil {
			return "", err
		}
	}
	if found == nil {
		found = user{"id": bind.ID, "userName": name, "name": name, "email": "", "isAdmin": admin, "password": "__NAVIDROME_AUTOGEN__" + randomID()}
		if err := a.request(ctx, "POST", "/api/user", found, nil); err != nil {
			return "", err
		}
	} else if found["isAdmin"] != admin || found["userName"] != name {
		found["isAdmin"] = admin
		found["userName"] = name
		if err := a.request(ctx, "PUT", "/api/user/"+url.PathEscape(bind.ID), found, nil); err != nil {
			return "", err
		}
	}
	if bind.Name != name {
		bind.Name = name
		a.bindings[key] = bind
		if err := a.save(); err != nil {
			return "", err
		}
	}
	return bind.ID, nil
}
func (a *accounts) delete(ctx context.Context, uid int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	key := strconv.Itoa(uid)
	b, ok := a.bindings[key]
	if !ok {
		return nil
	}
	// Read first so an SDK retry after a successful upstream delete is idempotent.
	var users []user
	if err := a.request(ctx, "GET", "/api/user", nil, &users); err != nil {
		return err
	}
	for _, u := range users {
		if u["id"] == b.ID {
			if err := a.request(ctx, "DELETE", "/api/user/"+url.PathEscape(b.ID), nil, nil); err != nil {
				return err
			}
		}
	}
	delete(a.bindings, key)
	return a.save()
}
