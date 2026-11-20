// Package httppres is a generic HTTP presentation driver: each action is
// one HTTP request. Covers slide software with a REST API (ProPresenter's
// network API, Companion buttons, custom bridges).
//
// addon.json:
//
//	{"driver": "http", "label": "ProPresenter", "baseUrl": "http://10.1.10.20:1025",
//	 "headers": {"Authorization": "..."},
//	 "actions": {
//	   "next": {"path": "/v1/trigger/next"},
//	   "prev": {"path": "/v1/trigger/previous"},
//	   "clear": {"method": "GET", "path": "/v1/clear/layer/slide"}}}
//
// Paths shown are illustrative; check them against your software's API.
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/toolingdev/ptz-controller/sdk/addon"
)

type Action struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Body   string `json:"body"`
}

type Config struct {
	BaseURL string            `json:"baseUrl"`
	Headers map[string]string `json:"headers"`
	Actions map[string]Action `json:"actions"`
}

type Driver struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last map[string]any
}

func New(c Config) (*Driver, error) {
	if !strings.HasPrefix(c.BaseURL, "http://") && !strings.HasPrefix(c.BaseURL, "https://") {
		return nil, fmt.Errorf("http presentation needs baseUrl (http://...)")
	}
	if len(c.Actions) == 0 {
		return nil, fmt.Errorf("http presentation needs at least one action")
	}
	c.BaseURL = strings.TrimRight(c.BaseURL, "/")
	return &Driver{cfg: c, http: &http.Client{Timeout: 3 * time.Second}}, nil
}

func (d *Driver) Info() addon.PresentationInfo {
	d.mu.Lock()
	defer d.mu.Unlock()
	return addon.PresentationInfo{Available: true, Target: d.cfg.BaseURL,
		Details: map[string]any{"lastSent": d.last}}
}

func (d *Driver) Actions() []string {
	out := make([]string, 0, len(d.cfg.Actions))
	for a := range d.cfg.Actions {
		out = append(out, a)
	}
	sort.Strings(out)
	return out
}

func (d *Driver) Do(ctx context.Context, action string) (map[string]any, error) {
	a, ok := d.cfg.Actions[action]
	if !ok {
		return nil, fmt.Errorf("%w %q", addon.ErrUnknownAction, action)
	}
	m := a.Method
	if m == "" {
		m = http.MethodGet
	}
	var body io.Reader
	if a.Body != "" {
		body = strings.NewReader(a.Body)
	}
	req, err := http.NewRequestWithContext(ctx, m, d.cfg.BaseURL+a.Path, body)
	if err != nil {
		return nil, err
	}
	for k, v := range d.cfg.Headers {
		req.Header.Set(k, v)
	}
	if a.Body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := d.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("presentation not reachable: %w", err)
	}
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("presentation answered HTTP %d for %s", resp.StatusCode, action)
	}
	det := map[string]any{"action": action, "status": resp.StatusCode, "port": d.cfg.BaseURL}
	d.mu.Lock()
	d.last = det
	d.mu.Unlock()
	return det, nil
}
