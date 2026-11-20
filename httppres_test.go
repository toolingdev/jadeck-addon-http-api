package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestActions(t *testing.T) {
	var got []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.Method+" "+r.URL.Path+" "+r.Header.Get("X-Key"))
		if r.URL.Path == "/bad" {
			w.WriteHeader(500)
		}
	}))
	defer ts.Close()
	d, err := New(Config{BaseURL: ts.URL + "/", Headers: map[string]string{"X-Key": "k"},
		Actions: map[string]Action{"next": {Path: "/v1/trigger/next"}, "boom": {Method: "POST", Path: "/bad", Body: "{}"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Do(context.Background(), "next"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Do(context.Background(), "boom"); err == nil {
		t.Fatal("HTTP 500 must fail")
	}
	if got[0] != "GET /v1/trigger/next k" || got[1] != "POST /bad k" {
		t.Fatal(got)
	}
	if _, err := New(Config{BaseURL: "ftp://x"}); err == nil {
		t.Fatal("bad base")
	}
}
