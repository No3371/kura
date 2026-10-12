package listing

import (
	"sync"
)

// Grade is one condition grade as a source defines it. Kura shows the source's
// own label and definition rather than translating grades into a house scale:
// a grade is the seller's judgement and should be presented as such.
type Grade struct {
	Code   string `json:"code"`    // stable key stored with listings (the source's label)
	Label  string `json:"label"`   // as shown by the source, e.g. ランクA
	DescJA string `json:"desc_ja"` // the source's own definition, verbatim
	DescZH string `json:"desc_zh"` // Traditional Chinese translation of DescJA
	// Rank places the grade on Kura's internal scale. It is only a bridge for
	// cross-grade price estimates and "at least this good" filters; it is never
	// shown in place of the source's grade.
	Rank Rank `json:"rank"`
}

// Scheme is a source's condition grading system.
type Scheme struct {
	Source string  `json:"source"`
	Name   string  `json:"name"`   // e.g. "BRAND OFF 商品ランク"
	Graded string  `json:"graded"` // who assigns grades, e.g. "店家鑑定" or "賣家自評"
	Ref    string  `json:"ref"`    // where the definitions come from
	Grades []Grade `json:"grades"` // best first
}

// Grade looks up a grade by code.
func (s Scheme) Grade(code string) (Grade, bool) {
	for _, g := range s.Grades {
		if g.Code == code {
			return g, true
		}
	}
	return Grade{}, false
}

var (
	schemesMu sync.RWMutex
	schemes   = map[string]Scheme{}
)

// RegisterScheme is called by source adapters (typically from init).
func RegisterScheme(s Scheme) {
	schemesMu.Lock()
	defer schemesMu.Unlock()
	schemes[s.Source] = s
}

// SchemeFor returns the grading scheme of a source.
func SchemeFor(source string) (Scheme, bool) {
	schemesMu.RLock()
	defer schemesMu.RUnlock()
	s, ok := schemes[source]
	return s, ok
}

// GradeFor resolves a stored grade code for a source. Unknown codes come back
// as a bare grade carrying just the code, so nothing is silently relabelled.
func GradeFor(source, code string) Grade {
	if s, ok := SchemeFor(source); ok {
		if g, ok := s.Grade(code); ok {
			return g
		}
	}
	return Grade{Code: code, Label: code}
}
