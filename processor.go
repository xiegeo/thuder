package thuder

import (
	"fmt"
	"os"
	"path/filepath"
)

//Processor does the recursive, depth first, processing of directories
type Processor struct {
	stack   []layer
	actions chan<- action // a buffered channal of queued actions to take
}

//layer is a layer in a Processor's stack
type layer struct {
	from []Node
	to   string
}

///action is an action to be done to the file system
type action struct {
	from []Node
	to   string
}

func applyAction(a action) []error {
	var errs []error
	for i := len(a.from) - 1; i >= 0; i-- {
		n := a.from[i]
		err := applyNode(n, a.to)
		if err != nil {
			errs = append(errs, err)
		}
		if n.IsDir() {
			break //only create dir once
		}
	}
	return errs
}

func applyNode(n Node, to string) error {
	target := filepath.Join(to, n.info.Name())
	if n.IsDelete() {
		return fs.RemoveAll(target)
	}
	if n.IsDir() {
		return fs.Mkdir(target, n.FileMode())
	}
	return atomicCopy(n, to)
}

//Do make the Processor process the stack until done
func (p *Processor) Do() {
	for p.doOnce() {
	}
	p.actions <- action{} //empty action means done
}

// returns true until there is nothing left to do
func (p *Processor) doOnce() bool {
	top := len(p.stack) - 1
	if top < 0 {
		return false
	}
	var l layer
	p.stack, l = p.stack[:top], p.stack[top] //pop from stack

	c := NewCollection()
	for _, node := range l.from {
		err := c.Add(&node)
		if err != nil {
			p.logError(node.FullName(), err)
		}
	}
	deletes, changedfiles, dirs, err := c.GetAppliedTo(l.to)
	if err != nil {
		p.logError(l.to, err)
		// continue even on error
	}
	a := action{to: l.to}
	if len(deletes) > 0 {
		a.from = deletes
		p.actions <- a
	}
	if len(changedfiles) > 0 {
		a.from = changedfiles
		p.actions <- a
	}
	if len(dirs) > 0 {
		for _, d := range dirs {
			a.from = d
			p.actions <- a //the create dir action

			last := d[len(d)-1]

			newLayer := layer{
				from: d,
				to:   filepath.Join(l.to, last.info.Name()),
			}
			p.stack = append(p.stack, newLayer)
		}
	}
	return true
}

func (p *Processor) logError(dir string, err error) {
	//todo: change this to a file on removalbe media
	fmt.Fprintln(os.Stderr, dir, err)
}
