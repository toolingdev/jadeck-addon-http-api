package main

import (
	"testing"

	"github.com/toolingdev/jadeck/sdk/addon/addontest"
)

func TestOverThePipe(t *testing.T) {
	h := addontest.Start(t, spec())
	if err := h.TryRequest("open", map[string]any{"conn": "x", "type": "http", "config": map[string]any{"baseUrl": "nope"}}, nil); err == nil {
		t.Fatal("bad base URL accepted")
	}
}
