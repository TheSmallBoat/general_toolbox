package metrics

import (
	"fmt"
	"sort"
	"sync"
)

// http://www.cs.ucsb.edu/research/tech_reports/reports/2005-23.pdf

type Element struct {
	Value string
	Count int
}

type Samples []*Element

func (sm Samples) Len() int {
	return len(sm)
}

func (sm Samples) Less(i, j int) bool {
	return sm[i].Count < sm[j].Count
}

func (sm Samples) Swap(i, j int) {
	sm[i], sm[j] = sm[j], sm[i]
}

type Stream struct {
	mu  sync.Mutex
	k   int
	mon map[string]*Element

	// the minimum Element
	min *Element

	// last new elements
	lastNewElements []*Element
}

func NewStream(k int) *Stream {
	s := new(Stream)
	s.k = k
	s.mon = make(map[string]*Element)
	s.min = &Element{}
	s.lastNewElements = []*Element{}

	// Track k+1 so that less frequent items contended for that spot, resulting in k being more accurate.
	return s
}

func (s *Stream) Reset() {
	defer s.mu.Unlock()
	s.mu.Lock()

	s.mon = nil
	s.mon = make(map[string]*Element)
	s.min = &Element{}
	s.lastNewElements = []*Element{}
}

func (s *Stream) Insert(x string) {
	s.insert(&Element{x, 1})
}

func (s *Stream) Merge(sm Samples) {
	for _, e := range sm {
		s.insert(e)
	}
}

func (s *Stream) SetNewElementBeforeMerge(sm Samples) {
	s.lastNewElements = []*Element{}
	for _, e := range sm {
		if s.mon[e.Value] == nil {
			s.lastNewElements = append(s.lastNewElements, e)
		}
	}
}

func (s *Stream) insert(in *Element) {
	defer s.mu.Unlock()
	s.mu.Lock()

	e := s.mon[in.Value]
	if e != nil {
		e.Count += in.Count
	} else {
		if len(s.mon) < s.k+1 {
			e = &Element{in.Value, in.Count}
			s.mon[in.Value] = e
		} else {
			e = s.min
			delete(s.mon, e.Value)
			e.Value = in.Value
			e.Count += in.Count
			s.min = e
		}
	}
	if e.Count < s.min.Count {
		s.min = e
	}
}

func (s *Stream) Query() Samples {
	var sm Samples
	for _, e := range s.mon {
		sm = append(sm, e)
	}
	sort.Sort(sort.Reverse(sm))

	if len(sm) < s.k {
		return sm
	}

	return sm[:s.k]
}

func (s *Stream) ReportTopKInfoByTxt(title string) string {
	report := ""
	if len(s.lastNewElements) > 0 {
		report += fmt.Sprintf("%s ...... New hot elements :[%d] \n", title, len(s.lastNewElements))
		for i, e := range s.lastNewElements {
			report += fmt.Sprintf("[NE%d -> '%s':%d] ", i+1, e.Value, e.Count)
			if (i+1)%4 == 0 {
				report += "\n"
			}
		}
		report += "\n"
	}

	sm := s.Query()
	if sm.Len() > 0 {
		report += fmt.Sprintf("%s ...... hot elements :[%d] \n", title, sm.Len())
		for i, e := range sm {
			report += fmt.Sprintf("[Top%d => '%s':%d] ", i+1, e.Value, e.Count)
			if (i+1)%4 == 0 {
				report += "\n"
			}
		}
		report += "\n"
	}
	return report
}

func (s *Stream) ReportTopKInfoByJson() string {
	report := fmt.Sprintf("{")
	if len(s.lastNewElements) > 0 {
		report += fmt.Sprintf("\"New Hot (%d)\":{", len(s.lastNewElements))
		for i, e := range s.lastNewElements {
			if i > 0 {
				report += ","
			}
			report += fmt.Sprintf("\"TOP%d\":{\"value\":\"%s\",\"count\":%d}", i, e.Value, e.Count)
		}
		report += "},"
	}

	sm := s.Query()
	if sm.Len() > 0 {
		report += fmt.Sprintf("\"Hot (%d)\":{", sm.Len())
		for i, e := range sm {
			if i > 0 {
				report += ","
			}
			report += fmt.Sprintf("\"TOP%d\":{\"value\":\"%s\",\"count\":%d}", i, e.Value, e.Count)
		}
		report += "}"
	}
	report += "}"
	return report
}
