package oval

import (
	"fmt"
	"sync"
)

func (s *States) init() {
	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		s.lineMemo = make(map[string]int, len(s.LineStates))
		for i, v := range s.LineStates {
			s.lineMemo[v.ID] = i
		}
	}()

	go func() {
		defer wg.Done()
		s.version55Memo = make(map[string]int, len(s.Version55States))
		for i, v := range s.Version55States {
			s.version55Memo[v.ID] = i
		}
	}()

	go func() {
		defer wg.Done()
		s.rpminfoMemo = make(map[string]int, len(s.RPMInfoStates))
		for i, v := range s.RPMInfoStates {
			s.rpminfoMemo[v.ID] = i
		}
	}()

	go func() {
		defer wg.Done()
		s.dpkginfoMemo = make(map[string]int, len(s.DpkgInfoStates))
		for i, v := range s.DpkgInfoStates {
			s.dpkginfoMemo[v.ID] = i
		}
	}()

	wg.Wait()
}

// Lookup returns the kind of object and index into that kind-specific slice, if
// found.
func (s *States) Lookup(ref string) (kind string, index int, err error) {
	s.once.Do(s.init)

	if i, ok := s.lineMemo[ref]; ok {
		return s.LineStates[i].XMLName.Local, i, nil
	}
	if i, ok := s.version55Memo[ref]; ok {
		return s.Version55States[i].XMLName.Local, i, nil
	}
	if i, ok := s.rpminfoMemo[ref]; ok {
		return s.RPMInfoStates[i].XMLName.Local, i, nil
	}
	if i, ok := s.dpkginfoMemo[ref]; ok {
		return s.DpkgInfoStates[i].XMLName.Local, i, nil
	}

	// We didn't find it, maybe we can say why.
	id, err := ParseID(ref)
	if err != nil {
		return "", -1, err
	}
	if id.Type != OvalState {
		return "", -1, fmt.Errorf("oval: wrong identifier type %q", id.Type)
	}
	return "", -1, ErrNotFound(ref)
}
