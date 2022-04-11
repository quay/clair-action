package oval

import (
	"fmt"
	"sync"
)

// Init sets up the memoization maps.
func (t *Tests) init() {
	var wg sync.WaitGroup
	wg.Add(7)

	go func() {
		defer wg.Done()
		t.lineMemo = make(map[string]int, len(t.LineTests))
		for i, v := range t.LineTests {
			t.lineMemo[v.ID] = i
		}
	}()

	go func() {
		defer wg.Done()
		t.version55Memo = make(map[string]int, len(t.Version55Tests))
		for i, v := range t.Version55Tests {
			t.version55Memo[v.ID] = i
		}
	}()

	go func() {
		defer wg.Done()
		t.rpminfoMemo = make(map[string]int, len(t.RPMInfoTests))
		for i, v := range t.RPMInfoTests {
			t.rpminfoMemo[v.ID] = i
		}
	}()

	go func() {
		defer wg.Done()
		t.dpkginfoMemo = make(map[string]int, len(t.DpkgInfoTests))
		for i, v := range t.DpkgInfoTests {
			t.dpkginfoMemo[v.ID] = i
		}
	}()

	go func() {
		defer wg.Done()
		t.rpmverifyfileMemo = make(map[string]int, len(t.RPMVerifyFileTests))
		for i, v := range t.RPMVerifyFileTests {
			t.rpmverifyfileMemo[v.ID] = i
		}
	}()

	go func() {
		defer wg.Done()
		t.unameMemo = make(map[string]int, len(t.UnameTests))
		for i, v := range t.UnameTests {
			t.unameMemo[v.ID] = i
		}
	}()

	go func() {
		defer wg.Done()
		t.textfilecontent54Memo = make(map[string]int, len(t.TextfileContent54Tests))
		for i, v := range t.TextfileContent54Tests {
			t.textfilecontent54Memo[v.ID] = i
		}
	}()

	wg.Wait()
}

// Lookup returns the kind of test and index into that kind-specific slice, if
// found.
func (t *Tests) Lookup(ref string) (kind string, index int, err error) {
	// Pay to construct an index, once.
	t.once.Do(t.init)

	if i, ok := t.lineMemo[ref]; ok {
		return t.LineTests[i].XMLName.Local, i, nil
	}
	if i, ok := t.version55Memo[ref]; ok {
		return t.Version55Tests[i].XMLName.Local, i, nil
	}
	if i, ok := t.rpminfoMemo[ref]; ok {
		return t.RPMInfoTests[i].XMLName.Local, i, nil
	}
	if i, ok := t.dpkginfoMemo[ref]; ok {
		return t.DpkgInfoTests[i].XMLName.Local, i, nil
	}
	if i, ok := t.rpmverifyfileMemo[ref]; ok {
		return t.RPMVerifyFileTests[i].XMLName.Local, i, nil
	}
	if i, ok := t.unameMemo[ref]; ok {
		return t.UnameTests[i].XMLName.Local, i, nil
	}
	if i, ok := t.textfilecontent54Memo[ref]; ok {
		return t.TextfileContent54Tests[i].XMLName.Local, i, nil
	}

	// We didn't find it, maybe we can say why.
	id, err := ParseID(ref)
	if err != nil {
		return "", -1, err
	}
	if id.Type != OvalTest {
		return "", -1, fmt.Errorf("oval: wrong identifier type %q", id.Type)
	}
	return "", -1, ErrNotFound(ref)
}

// Test is an interface for OVAL objects that reports Object and State refs.
type Test interface {
	ObjectRef() []ObjectRef
	StateRef() []StateRef
}

// TestRef is an embeddable struct that implements the Test interface.
type testRef struct {
	ObjectRefs []ObjectRef `xml:"object"`
	StateRefs  []StateRef  `xml:"state"`
}

func (t *testRef) ObjectRef() []ObjectRef { return t.ObjectRefs }
func (t *testRef) StateRef() []StateRef   { return t.StateRefs }
