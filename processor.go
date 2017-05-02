package thuder

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

//PullAndPush does push and pulls based on given configurations, it uses Processors
func PullAndPush(hc *HostConfig, mc *MediaConfig, debug io.Writer) error {
	if debug == nil {
		debug = ioutil.Discard
	}

	actions := make(chan action, 8)
	apply := func(p *Processor) {
		go p.Do()
		for {
			a := <-actions
			if len(a.from) == 0 {
				return
			}
			err := applyAction(a)
			if err != nil {
				fmt.Fprintln(debug, err)
			}
		}
	}

	p, err := NewProcessor(mc.Pulls, hc.PullTarget(), actions)
	if err != nil {
		return err
	}
	apply(p)

	/*
		p, err = NewProcessor(mc.Pushes, "/", actions)
		if err != nil {
			return err
		}
		apply(p)
	*/
	return nil
}

//Processor does the recursive, depth first, processing of directories
type Processor struct {
	stack   []layer
	actions chan<- action // a buffered channal of queued actions to take
}

//NewProcessor create a new Processor
func NewProcessor(dirs []string, to string, actions chan<- action) (*Processor, error) {
	var sources []Node
	for _, fullname := range dirs[:2] {
		rootNode, err := NewRootNode(fullname, false)
		if err != nil {
			return nil, err
		}
		sources = append(sources, *rootNode)
	}
	p := Processor{
		stack: []layer{
			layer{from: sources, to: to},
		},
		actions: actions,
	}
	return &p, nil
}

//String returns string debugging
func (p *Processor) String() string {
	b := bytes.NewBufferString("{stack:[")
	for _, l := range p.stack {
		b.WriteString("\n\t" + l.String())
	}
	b.WriteString("]")
	return b.String()
}

//layer is a layer in a Processor's stack
type layer struct {
	from []Node
	to   string
}

//String returns string debugging
func (l layer) String() string {
	b := bytes.NewBufferString("{from:[")
	for _, n := range l.from {
		b.WriteString(n.String() + " ")
	}
	b.WriteString("] to:" + l.to + "}")
	return b.String()
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
	}
	return errs
}

func applyNode(n Node, to string) error {
	target := filepath.Join(to, n.info.Name())
	if n.IsDelete() {
		//fmt.Println("remove", n.info.Name())
		return fs.RemoveAll(target)
	}
	if n.IsDir() {
		//fmt.Println("mkdir", n.info.Name())
		return fs.Mkdir(target, n.FileMode())
	}
	//fmt.Println("copy", n.info.Name())
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
	//fmt.Println(p)
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
