package allu

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func hydrateRaw(t *testing.T, raw string) map[string]any {
	t.Helper()
	root, err := payload(strings.NewReader(`<script type="application/json" id="__NUXT_DATA__">` + raw + `</script>`))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestHydrate(t *testing.T) {
	root := hydrateRaw(t, `[{"data":1},{"k":2,"r":3,"n":-1,"z":-6},"v",["Reactive",4],{"x":2}]`)
	data := root["data"].(map[string]any)
	if data["k"] != "v" {
		t.Errorf("data.k = %v, want v", data["k"])
	}
	if r, _ := data["r"].(map[string]any); r["x"] != "v" {
		t.Errorf("data.r.x = %v, want v (shared index)", data["r"])
	}
	if v, ok := data["n"]; !ok || v != nil {
		t.Errorf("data.n = %v (present %v), want nil", v, ok)
	}
	if data["z"] != 0.0 {
		t.Errorf("data.z = %v, want 0 (negative zero)", data["z"])
	}

	root = hydrateRaw(t, `[{"s":1,"d":4,"m":5,"e":7},["Set",2,3],"a","b",["Date","2026-10-13T00:00:00Z"],["Map",2,6],1.5,["EmptyRef"]]`)
	if !reflect.DeepEqual(root["s"], []any{"a", "b"}) {
		t.Errorf("Set = %#v", root["s"])
	}
	if root["d"] != "2026-10-13T00:00:00Z" {
		t.Errorf("Date = %#v", root["d"])
	}
	if !reflect.DeepEqual(root["m"], map[string]any{"a": 1.5}) {
		t.Errorf("Map = %#v", root["m"])
	}
	if root["e"] != nil {
		t.Errorf("EmptyRef = %#v, want nil", root["e"])
	}

	// A null-prototype object keeps its keys as literal strings; the values
	// are indices. Read as indices, both keys would resolve to nil and collide.
	root = hydrateRaw(t, `[{"o":2},"x",["null","a",3,"b",1],2.5]`)
	if !reflect.DeepEqual(root["o"], map[string]any{"a": 2.5, "b": "x"}) {
		t.Errorf("null-prototype object = %#v", root["o"])
	}

	// A self-referencing node resolves its back edge to nil instead of looping.
	root = hydrateRaw(t, `[{"self":1},{"me":1,"list":2},[1,3],"x"]`)
	self := root["self"].(map[string]any)
	if self["me"] != nil {
		t.Errorf("cycle back edge = %#v, want nil", self["me"])
	}
	if l, _ := self["list"].([]any); len(l) != 2 || l[0] != nil || l[1] != "x" {
		t.Errorf("list = %#v", self["list"])
	}
}

func TestPayloadMissing(t *testing.T) {
	if _, err := payload(strings.NewReader(`<html><body><p>no state</p></body></html>`)); err == nil ||
		!strings.Contains(err.Error(), "no __NUXT_DATA__") {
		t.Fatalf("err = %v, want missing __NUXT_DATA__", err)
	}
}

func TestNode(t *testing.T) {
	root := hydrateRaw(t, `[{"data":1},{"market:asset:1":2,"market:asset:2":2},{"color":3},"red"]`)
	if _, _, err := node[apiAsset](root, "market:asset:"); err == nil || !strings.Contains(err.Error(), "two") {
		t.Errorf("two matches: err = %v", err)
	}
	if _, _, err := node[apiAsset](root, "market:searchItems:"); err == nil {
		t.Error("no match: want error")
	}
}
