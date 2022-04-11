package oval

import (
	"fmt"
	"sync"
)

func (o *Variables) init() {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		o.dpkginfoMemo = make(map[string]int, len(o.ConstantVariables))
		for i, v := range o.ConstantVariables {
			o.dpkginfoMemo[v.ID] = i
		}
	}()

	wg.Wait()
}

// Lookup returns the kind of object and index into that kind-specific slice, if
// found.
func (o *Variables) Lookup(ref string) (kind string, index int, err error) {
	o.once.Do(o.init)
	if i, ok := o.dpkginfoMemo[ref]; ok {
		return o.ConstantVariables[i].XMLName.Local, i, nil
	}

	// We didn't find it, maybe we can say why.
	id, err := ParseID(ref)
	if err != nil {
		return "", -1, err
	}
	if id.Type != OvalVariable {
		return "", -1, fmt.Errorf("oval: wrong identifier type %q", id.Type)
	}
	return "", -1, ErrNotFound(ref)
}
