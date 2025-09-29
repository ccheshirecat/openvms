package conformance

import (
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestExampleManifestBasicShape(t *testing.T) {
    p := filepath.Join("..", "..", "spec", "examples", "ubuntu-manifest.json")
    b, err := os.ReadFile(p)
    if err != nil {
        t.Fatalf("read example manifest: %v", err)
    }
    var m map[string]any
    if err := json.Unmarshal(b, &m); err != nil {
        t.Fatalf("unmarshal: %v", err)
    }
    if v, ok := m["schemaVersion"].(float64); !ok || int(v) != 1 {
        t.Fatalf("schemaVersion must be 1, got %v", m["schemaVersion"])
    }
    if mt, ok := m["mediaType"].(string); !ok || mt != "application/vnd.ovms.manifest.v1+json" {
        t.Fatalf("mediaType mismatch: %v", m["mediaType"])
    }
    // basic presence checks
    for _, k := range []string{"name", "version", "kernel", "diskLayers"} {
        if _, ok := m[k]; !ok {
            t.Fatalf("missing required field: %s", k)
        }
    }
}

func TestOpenAPIPresent(t *testing.T) {
    p := filepath.Join("..", "..", "spec", "api", "ovms-runtime.openapi.yaml")
    b, err := os.ReadFile(p)
    if err != nil {
        t.Fatalf("read openapi: %v", err)
    }
    s := string(b)
    if !strings.Contains(s, "openapi: 3.0.3") {
        t.Fatalf("openapi file must declare 3.0.3")
    }
    for _, path := range []string{"/start", "/stop", "/snapshot", "/status/{instance_id}", "/logs/{instance_id}"} {
        if !strings.Contains(s, path+":") {
            t.Fatalf("missing path in openapi: %s", path)
        }
    }
}
