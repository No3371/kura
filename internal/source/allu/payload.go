package allu

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// payload reads the page's __NUXT_DATA__ script and returns the hydrated
// root object. Nuxt serialises state with devalue: a flat JSON array where
// containers hold indices into the array; arrays whose first element is a
// string are typed values (["Reactive", i], ["Set", …], ["Date", s], …).
func payload(r io.Reader) (map[string]any, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	s := doc.Find("script#__NUXT_DATA__").First()
	if s.Length() == 0 {
		return nil, errors.New("allu: no __NUXT_DATA__ on page")
	}
	var arr []any
	if err := json.Unmarshal([]byte(s.Text()), &arr); err != nil {
		return nil, fmt.Errorf("allu: page payload: %w", err)
	}
	if len(arr) == 0 {
		return nil, errors.New("allu: empty page payload")
	}
	h := hydrator{arr: arr, memo: map[int]any{}, busy: map[int]bool{}}
	root, ok := h.at(0).(map[string]any)
	if !ok {
		return nil, errors.New("allu: page payload root is not an object")
	}
	return root, nil
}

// hydrator resolves devalue indices. Shared indices resolve to the same value
// (memo); an index met again while it is still being resolved is a cycle and
// resolves to nil. Kura reads no cyclic node, so dropping the back edge loses
// nothing it uses; rejected: following it, which never terminates.
type hydrator struct {
	arr  []any
	memo map[int]any
	busy map[int]bool
}

// ref resolves a slot that holds an index. encoding/json decodes numbers as
// float64; devalue's negative indices are specials: -1 undefined, -2 hole,
// -3 NaN, -4/-5 ±Infinity, -6 negative zero.
func (h *hydrator) ref(v any) any {
	f, ok := v.(float64)
	if !ok {
		return nil
	}
	i := int(f)
	if i < 0 {
		if i == -6 {
			return 0.0
		}
		return nil
	}
	return h.at(i)
}

func (h *hydrator) at(i int) any {
	if i >= len(h.arr) {
		return nil
	}
	if v, ok := h.memo[i]; ok {
		return v
	}
	if h.busy[i] {
		return nil
	}
	h.busy[i] = true
	v := h.value(h.arr[i])
	delete(h.busy, i)
	h.memo[i] = v
	return v
}

func (h *hydrator) value(v any) any {
	switch v := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, x := range v {
			out[k] = h.ref(x)
		}
		return out
	case []any:
		if len(v) > 0 {
			if tag, ok := v[0].(string); ok {
				return h.typed(tag, v[1:])
			}
		}
		out := make([]any, len(v))
		for k, x := range v {
			out[k] = h.ref(x)
		}
		return out
	default:
		return v // string, bool, float64, nil
	}
}

func (h *hydrator) typed(tag string, rest []any) any {
	switch tag {
	case "Reactive", "ShallowReactive", "Ref", "ShallowRef":
		if len(rest) == 0 {
			return nil
		}
		return h.ref(rest[0])
	case "Date":
		if len(rest) == 0 {
			return nil
		}
		s, _ := rest[0].(string)
		return s
	case "Set":
		out := make([]any, len(rest))
		for k, x := range rest {
			out[k] = h.ref(x)
		}
		return out
	case "Map":
		out := make(map[string]any, len(rest)/2)
		for k := 0; k+1 < len(rest); k += 2 {
			out[fmt.Sprint(h.ref(rest[k]))] = h.ref(rest[k+1])
		}
		return out
	case "null":
		// A null-prototype object: ["null", "key", i, …]. Unlike Map, its keys
		// are literal strings, not indices; resolving them through ref would
		// turn every key into <nil> and collapse the pairs.
		out := make(map[string]any, len(rest)/2)
		for k := 0; k+1 < len(rest); k += 2 {
			key, _ := rest[k].(string)
			out[key] = h.ref(rest[k+1])
		}
		return out
	}
	return nil // EmptyRef, EmptyShallowRef and tags Kura does not read
}

// node finds the one entry of root["data"] whose key starts with keyPrefix
// and decodes it into T. Nuxt keys async data by request (for example
// "market:asset:2118097"), so only the prefix is stable.
func node[T any](root map[string]any, keyPrefix string) (T, string, error) {
	var zero T
	data, _ := root["data"].(map[string]any)
	var key string
	for k := range data {
		if !strings.HasPrefix(k, keyPrefix) {
			continue
		}
		if key != "" {
			return zero, "", fmt.Errorf("allu: two %q entries in page payload: %q and %q", keyPrefix, key, k)
		}
		key = k
	}
	if key == "" {
		return zero, "", fmt.Errorf("allu: no %q in page payload", keyPrefix)
	}
	b, err := json.Marshal(data[key])
	if err != nil {
		return zero, "", err
	}
	var out T
	if err := json.Unmarshal(b, &out); err != nil {
		return zero, "", fmt.Errorf("allu: %s: %w", key, err)
	}
	return out, key, nil
}
