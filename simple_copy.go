package thuder

import (
	"fmt"
	"time"
)

//Copy exposes a simple copy command to be used programtically.
func Copy(source, dest string, filters []Filter) error {
	sourceNode, err := NewRootNode(source, false)
	if err != nil {
		return fmt.Errorf("source string error: %v", err)
	}
	err = PrepareFilters(filters)
	if err != nil {
		return fmt.Errorf("filter format error: %v", err)
	}

	actions := make(chan action, 8)
	now := time.Now()

	p := Processor{
		stack:   []layer{{from: []Node{*sourceNode}, to: dest}},
		actions: actions,
		accept: func(n *Node) bool {
			_, a := MatchFilters(filters, n, false, now)
			return a
		},
	}

	go p.Do()
	for {
		a := <-actions
		if len(a.from) == 0 {
			return p.LoggedErrors()
		}
		LogP("Applying %v actions to %v.\n", len(a.from), a.to)
		errs := applyAction(a)
		if len(errs) != 0 {
			p.logErrors(a.to, errs)
		}
	}
}
