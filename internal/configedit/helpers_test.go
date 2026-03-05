package configedit_test

import (
	"encoding/json"
	"testing"

	"github.com/hermetiq/buildbarn-config-mcp/internal/configedit"
)

const sampleJsonnet = `{
  grpcServers: [
    {
      listenAddresses: ["0.0.0.0:8980"],
    },
  ],
  schedulers: {
    primary: {
      endpoint: "scheduler:8982",
    },
  },
}`

func mustEval(t *testing.T, src string) interface{} {
	t.Helper()
	data, err := configedit.EvalJsonnet(src)
	if err != nil {
		t.Fatalf("EvalJsonnet: %v", err)
	}
	return data
}

func mustParsePath(t *testing.T, raw string) []string {
	t.Helper()
	p, err := configedit.ParsePath(raw)
	if err != nil {
		t.Fatalf("ParsePath: %v", err)
	}
	return p
}

// ── EvalJsonnet ───────────────────────────────────────────────────────────────

func TestEvalJsonnet_Simple(t *testing.T) {
	data := mustEval(t, sampleJsonnet)
	m, ok := data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", data)
	}
	if _, ok := m["grpcServers"]; !ok {
		t.Fatal("missing grpcServers key")
	}
}

func TestEvalJsonnet_Invalid(t *testing.T) {
	_, err := configedit.EvalJsonnet("{bad syntax")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEvalJsonnet_ImportBlocked(t *testing.T) {
	_, err := configedit.EvalJsonnet(`import "some/file.libsonnet"`)
	if err == nil {
		t.Fatal("expected import to be blocked")
	}
}

// ── ParsePath ─────────────────────────────────────────────────────────────────

func TestParsePath_StringKeys(t *testing.T) {
	p := mustParsePath(t, `["grpcServers","0","listenAddresses","0"]`)
	if len(p) != 4 || p[0] != "grpcServers" || p[2] != "listenAddresses" {
		t.Fatalf("unexpected path: %v", p)
	}
}

func TestParsePath_NumericIndex(t *testing.T) {
	p := mustParsePath(t, `[0, 1, 2]`)
	if p[0] != "0" || p[1] != "1" || p[2] != "2" {
		t.Fatalf("unexpected: %v", p)
	}
}

func TestParsePath_Empty(t *testing.T) {
	_, err := configedit.ParsePath("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

// ── GetAt ─────────────────────────────────────────────────────────────────────

func TestGetAt_DeepValue(t *testing.T) {
	data := mustEval(t, sampleJsonnet)
	path := mustParsePath(t, `["grpcServers","0","listenAddresses","0"]`)
	val, err := configedit.GetAt(data, path)
	if err != nil {
		t.Fatalf("GetAt: %v", err)
	}
	if val != "0.0.0.0:8980" {
		t.Fatalf("expected '0.0.0.0:8980', got %v", val)
	}
}

func TestGetAt_MissingKey(t *testing.T) {
	data := mustEval(t, sampleJsonnet)
	path := mustParsePath(t, `["nonexistent"]`)
	_, err := configedit.GetAt(data, path)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

// ── SetAt ─────────────────────────────────────────────────────────────────────

func TestSetAt_UpdateScalar(t *testing.T) {
	data := mustEval(t, sampleJsonnet)
	path := mustParsePath(t, `["grpcServers","0","listenAddresses","0"]`)
	updated, err := configedit.SetAt(data, path, "0.0.0.0:9090")
	if err != nil {
		t.Fatalf("SetAt: %v", err)
	}
	val, _ := configedit.GetAt(updated, path)
	if val != "0.0.0.0:9090" {
		t.Fatalf("expected updated value, got %v", val)
	}
	// Original should be unchanged
	orig, _ := configedit.GetAt(data, path)
	if orig != "0.0.0.0:8980" {
		t.Fatal("original data was mutated")
	}
}

func TestSetAt_CreateNewPath(t *testing.T) {
	data := mustEval(t, `{}`)
	path := mustParsePath(t, `["newKey"]`)
	updated, err := configedit.SetAt(data, path, "hello")
	if err != nil {
		t.Fatalf("SetAt: %v", err)
	}
	val, _ := configedit.GetAt(updated, path)
	if val != "hello" {
		t.Fatalf("expected 'hello', got %v", val)
	}
}

func TestSetAt_ReplaceRoot(t *testing.T) {
	data := mustEval(t, `{a: 1}`)
	updated, err := configedit.SetAt(data, []string{}, map[string]interface{}{"b": float64(2)})
	if err != nil {
		t.Fatalf("SetAt: %v", err)
	}
	m := updated.(map[string]interface{})
	if m["b"] != float64(2) {
		t.Fatalf("unexpected root: %v", updated)
	}
}

// ── DeleteAt ──────────────────────────────────────────────────────────────────

func TestDeleteAt_ObjectKey(t *testing.T) {
	data := mustEval(t, sampleJsonnet)
	path := mustParsePath(t, `["schedulers","primary","endpoint"]`)
	updated, err := configedit.DeleteAt(data, path)
	if err != nil {
		t.Fatalf("DeleteAt: %v", err)
	}
	m := updated.(map[string]interface{})["schedulers"].(map[string]interface{})["primary"].(map[string]interface{})
	if _, ok := m["endpoint"]; ok {
		t.Fatal("key should have been deleted")
	}
}

func TestDeleteAt_ArrayElement(t *testing.T) {
	data := mustEval(t, `{items: ["a","b","c"]}`)
	path := mustParsePath(t, `["items","1"]`)
	updated, err := configedit.DeleteAt(data, path)
	if err != nil {
		t.Fatalf("DeleteAt: %v", err)
	}
	arr := updated.(map[string]interface{})["items"].([]interface{})
	if len(arr) != 2 || arr[0] != "a" || arr[1] != "c" {
		t.Fatalf("unexpected array after delete: %v", arr)
	}
}

func TestDeleteAt_Root(t *testing.T) {
	data := mustEval(t, `{a: 1}`)
	_, err := configedit.DeleteAt(data, []string{})
	if err == nil {
		t.Fatal("expected error deleting root")
	}
}

// ── DataToJsonnet ─────────────────────────────────────────────────────────────

func TestDataToJsonnet_RoundTrip(t *testing.T) {
	data := mustEval(t, sampleJsonnet)
	out, err := configedit.DataToJsonnet(data)
	if err != nil {
		t.Fatalf("DataToJsonnet: %v", err)
	}
	// Re-parse as JSON (DataToJsonnet outputs standard JSON)
	var reparsed interface{}
	if err := json.Unmarshal([]byte(out), &reparsed); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	// Check a field survived the round-trip
	path := mustParsePath(t, `["schedulers","primary","endpoint"]`)
	val, err := configedit.GetAt(reparsed, path)
	if err != nil {
		t.Fatalf("GetAt after round-trip: %v", err)
	}
	if val != "scheduler:8982" {
		t.Fatalf("unexpected value: %v", val)
	}
}
