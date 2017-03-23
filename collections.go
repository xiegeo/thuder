package thuder

import (
	"errors"
	"os"
	"strings"
)

var (
	ErrBadPath = errors.New("the file path is not of required formate")
	ErrNeedDir = errors.New("a directory is required for this operation")
)

//Collection is a document tree that collects meta data of changes in a directory
//to be made
type Collection struct {
	nodes map[string][]Node
}

func NewCollection() *Collection {
	return &Collection{
		nodes: make(map[string][]Node),
	}
}

func (c *Collection) PrintTo(f func(format string, args ...interface{})) {
	for k, v := range c.nodes {
		f("%s: %s\n", k, &v)
	}
}

//get returns all nodes seen with this name ignoring case
func (c *Collection) Get(name string) []Node {
	return c.nodes[strings.ToUpper(name)]
}

//Add adds all nodes (direct child of given parent) by filename to the collection.
//Existing files with the same name are overwitten.
//Existing folders are added.
func (c *Collection) Add(parent *Node) error {
	if !parent.info.IsDir() {
		return ErrNeedDir
	}
	f, err := parent.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	list, err := f.Readdir(-1)
	if err != nil {
		return err
	}
	fc := NewFileContext(parent)
	c.AddList(fc, list)
	return nil
}

//AddList is same as Add, but with give FileContext and FileInfo slice
func (c *Collection) AddList(fc *FileContext, list []os.FileInfo) {
	for _, fi := range list {
		name := strings.ToUpper(fi.Name())
		node := Node{
			fc:   fc,
			info: fi,
		}
		old := c.nodes[name]
		if len(old) == 0 {
			c.nodes[name] = []Node{node}
		} else if node.IsDir() {
			if old[0].IsDir() {
				//add dir to dir list
				c.nodes[name] = append(old, node)
			} else {
				//replace file with new dir list
				c.nodes[name] = []Node{node}
			}
		} else if old[0].IsDir() {
			//keep dir list, ignore new file
		} else if old[0].fc.from != node.fc.from { //files (no folders) only starting here
			//different path, replace
			c.nodes[name] = []Node{node}
		} else if node.IsDelete() {
			//delete duplicate in same path, ignore
		} else if old[0].IsDelete() {
			//replace delete
			c.nodes[name] = []Node{node}
		} else if strings.Compare(old[0].info.Name(), node.info.Name()) > 0 {
			//consistently choose one file
			c.nodes[name] = []Node{node}
		}

		//assertion
		if len(c.nodes[name]) > 1 {
			for _, n := range c.nodes[name] {
				if !n.IsDir() {
					panic("assertion failed: []node longer than 1 must be directories only")
				}
			}
		}
	}
}

//GetAppliedTo returns list of nodes as actions to be taken on the target
//path such that the operation is consistant.
//Such as: case-sensitive act as case-perserving.
//
//The target dir must have been created
func (c *Collection) GetAppliedTo(target string) ([]Node, error) {
	t, err := NewRootNode(target)
	if err != nil {
		return nil, err
	}
	if !t.IsDir() {
		return nil, ErrNeedDir
	}

	exist := NewCollection() //Collect nodes from target
	exist.Add(t)

	return nil, nil //todo
}
