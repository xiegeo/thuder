package thuder

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

//PullAndPush does push and pulls based on given configurations, it uses Processors
//
//TODO: push filtering
func PullAndPush(hc *HostConfig, mc *MediaConfig) error {

	err := PrepareFilters(hc.Filters)
	if err != nil {
		return fmt.Errorf("HostConfig PrepareFilters error: %v", err)
	}
	err = PrepareFilters(mc.Filters)
	if err != nil {
		return fmt.Errorf("MediaConfig PrepareFilters error: %v", err)
	}
	isPush := false
	now := time.Now()
	accept := func(n *Node) bool {
		_, a := MatchFilters(hc.Filters, n, isPush, now)
		if !a {
			return false
		}
		_, a = MatchFilters(mc.Filters, n, isPush, now)
		return a
	}

	actions := make(chan action, 8)
	apply := func(p *Processor) {
		go p.Do()
		for {
			a := <-actions
			if len(a.from) == 0 {
				return
			}
			LogP("Appling %v actions to %v.\n", len(a.from), a.to)
			errs := applyAction(a)
			if len(errs) != 0 {
				p.logErrors(a.to, errs)
			}
		}
	}

	p, err := NewPullingProcessor(mc.Pulls, hc.PullTarget(), actions, accept)
	if err != nil {
		return err
	}
	apply(p)

	isPush = true
	p, err = NewPushingProcessor(hc, actions, accept)
	if err != nil {
		return err
	}
	apply(p)

	syncWriteCache()

	return nil
}

//Processor does the recursive, depth first, processing of directories
type Processor struct {
	stack   []layer
	actions chan<- action // a buffered channal of queued actions to take
	accept  func(n *Node) bool
}

//joinSub is filePath.Join with additional special charcter handling
func joinSub(parent, sub string) string {
	if filepath.Separator == '\\' && len(sub) > 1 && sub[1] == ':' {
		if len(sub) > 2 {
			sub = sub[0:1] + sub[2:]
		} else {
			sub = sub[0:1]
		}
	}
	return filepath.Join(parent, sub)
}

//NewPullingProcessor create a new Processor for pulling dirs from host to media.
func NewPullingProcessor(dirs []string, pullTo string, actions chan<- action, accept func(n *Node) bool) (*Processor, error) {
	var stack []layer
	for _, fullname := range dirs {
		rootNode, err := NewRootNode(fullname, false)
		if err != nil {
			return nil, err
		}
		to := joinSub(pullTo, fullname)
		LogP("Pulls dir %v to %v.\n", fullname, to)
		err = fs.MkdirAll(to, 0755)
		if err != nil {
			return nil, err
		}
		stack = append(stack, layer{from: []Node{*rootNode}, to: to})
	}
	p := Processor{
		stack:   stack,
		actions: actions,
		accept:  accept,
	}
	return &p, nil
}

//NewPushingProcessor create a new Processor for pushing dirs from media to host.
func NewPushingProcessor(hc *HostConfig, actions chan<- action, accept func(n *Node) bool) (*Processor, error) {
	var stack []layer
	sources, isDeletes := hc.PushSources()
	for _, root := range hc.PushRoots() {
		var nodes []Node
		for i := 0; i < len(sources); i++ {
			from := joinSub(sources[i], root)
			node, err := NewRootNode(from, isDeletes[i])
			if err != nil {
				//LogP("Pushes skipped from %v because error %v.\n", from, err)
				continue
			} else {
				LogP("pushes from %v to %v.\n", from, root)
			}
			nodes = append(nodes, *node)
		}
		stack = append(stack, layer{from: nodes, to: root})
	}
	p := Processor{
		stack:   stack,
		actions: actions,
		accept:  accept,
	}
	return &p, nil
}

//NewProcessor create a new Processor
func newProcessor(dirs []string, to string, actions chan<- action, accept func(n *Node) bool) (*Processor, error) {
	var sources []Node
	for _, fullname := range dirs {
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
		accept:  accept,
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

//String returns string for debugging
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
	FlashLED()
	var l layer
	p.stack, l = p.stack[:top], p.stack[top] //pop from stack

	c := NewCollection(p.accept)
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

//LogErrorOut used to redirect error logs
var LogErrorOut io.Writer = os.Stderr

func (p *Processor) logError(dir string, err error) {
	fmt.Fprintln(LogErrorOut, "Processor Error: ", dir, err)
}
func (p *Processor) logErrors(dir string, errs []error) {
	fmt.Fprintln(LogErrorOut, "Processor Errors: ", dir, errs)
}

//LogVerbosOut used to redirect verbos logs
var LogVerbosOut io.Writer = os.Stdout

//LogP is the handler for logging live progress, in the form of fmt.Printf
var LogP = func(format string, a ...interface{}) {
	fmt.Fprintf(LogVerbosOut, format, a...)
}
