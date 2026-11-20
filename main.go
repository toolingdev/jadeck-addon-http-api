// HTTP API: a generic jadeck add-on. Each slide action is one
// web request you set up (ProPresenter and others).
package main

import (
	"encoding/json"

	"github.com/toolingdev/jadeck/sdk/addon"
)

// Version is set by the release build.
var Version = "1.0.0"

func main() { addon.Serve(spec()) }

func spec() addon.Spec {
	return addon.Spec{
		Manifest: addon.Manifest{
			ID: "http-api", Name: "HTTP API", Version: Version, Publisher: "jadeck", APIVersion: addon.APIVersion, Repo: "toolingdev/jadeck-addon-http-api", Generic: true,
			Description: "Each slide action is one web request you set up (ProPresenter and others).",
			Reaches:     "The web addresses you enter for each action",
			ConnectionTypes: []addon.ConnectionType{{
				ID: "http", Label: "HTTP API", Capabilities: []string{addon.CapSlides},
				Description: "Each action is one HTTP request to your slide software's network API.",
				Fields: []addon.Field{
					{Key: "baseUrl", Label: "Base URL", Type: "text", Placeholder: "http://192.168.1.20:1025", Required: true},
					{Key: "actions", Label: "Actions", Type: "json", Required: true,
						Default: map[string]any{"next": map[string]string{"path": "/v1/trigger/next"}, "prev": map[string]string{"path": "/v1/trigger/previous"}},
						Help:    `Action → {"method": "GET", "path": "/…", "body": ""}. Check the paths against your software's API docs.`},
					{Key: "headers", Label: "Headers", Type: "json", Advanced: true, Help: `e.g. {"Authorization": "…"}`},
				},
				Sample:     map[string]any{"baseUrl": "http://192.168.1.20:1025"},
				SampleTest: "Ready — sends to http://192.168.1.20:1025",
			}},
			Widgets: []addon.WidgetDef{{Kind: "core", Type: "slides.prevnext", Label: "Prev / Next", Sample: map[string]any{
				"/api/presentation/status": map[string]any{"available": true, "port": "http://192.168.1.20:1025", "actions": []string{"prev", "next"}}}}},
		},
		Open: map[string]addon.OpenFunc{"http": func(raw json.RawMessage) (any, error) {
			var c Config
			if err := json.Unmarshal(raw, &c); err != nil {
				return nil, err
			}
			return New(c)
		}},
	}
}
