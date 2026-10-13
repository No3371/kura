package listing

import (
	"slices"
	"testing"
)

// Schemes is iterated from a map, so the sort is the only thing keeping the
// methodology page's ladders in a stable order. Registering in reverse order
// with several entries makes a missing sort fail on almost every run instead
// of passing by luck of map iteration.
func TestSchemesSortedBySource(t *testing.T) {
	want := []string{"sch-a", "sch-b", "sch-c", "sch-d", "sch-e", "sch-f"}
	for i := len(want) - 1; i >= 0; i-- {
		RegisterScheme(Scheme{Source: want[i], Name: want[i]})
	}
	var got []string
	for _, s := range Schemes() {
		if slices.Contains(want, s.Source) {
			got = append(got, s.Source)
		}
	}
	if !slices.Equal(got, want) {
		t.Errorf("Schemes() order = %v, want %v", got, want)
	}
}
