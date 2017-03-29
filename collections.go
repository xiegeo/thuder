package thuder

import (
	"errors"
	"os"
	"strings"
)

var (
	//ErrBadPath is returned when path is not of required formate, such as a absolute
	//path when required. This signals a misconfiguration for the current operating system.
	ErrBadPath = errors.New("the file path is not of required formate")
	//ErrNeedDir is returned when an expected directory does not exist or is not a directory.
	//There for, the operation can not be applied.
	ErrNeedDir = errors.New("a directory is required for this operation")
)

//Collection is a document tree that collects meta data of changes in a directory
//to be made.
type Collection struct {
	nodes map[string][]Node
}

//NewCollection initializes a new empty Collection.
func NewCollection() *Collection {
	return &Collection{
		nodes: make(map[string][]Node),
	}
}

//PrintTo prints all nodes in a collection, useful for debuging.
func (c *Collection) PrintTo(f func(format string, args ...interface{})) {
	for k, v := range c.nodes {
		f("%s: %s\n", k, &v)
	}
}

//Get returns all nodes seen with this name ignoring case
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

		/*
			conflict resolution matrix
			old only checks last element in nodes list

			old: | dif folder| same folder      | empty
							 | file | del | dir |
			new:
			file | R | A | R | A | R
			del  | R | X | X | X | R
			dir  | A | A | R | A | R

			R : replace with new list
			A : use addNode function
			X : no-op
		*/

		if len(old) == 0 { //empty case
			c.nodes[name] = []Node{node}
			continue
		}
		index := len(old) - 1
		last := old[index]

		if node.IsDelete() && !node.IsDir() { //del file case
			if node.SameDir(last) {
				continue // no-op
			}
			c.nodes[name] = []Node{node}
			continue
		}

		if node.SameDir(last) { //same folder case
			if last.IsDelete() && !last.IsDir() {
				c.nodes[name] = []Node{node}
				continue
			}
			c.nodes[name] = addNode(c.nodes[name], node)
			continue
		}

		//start dif folder case
		if node.IsDir() {
			c.nodes[name] = addNode(c.nodes[name], node)
			continue
		}
		c.nodes[name] = []Node{node}
	}
}

//GetAppliedTo returns list of nodes as actions to be taken on the target
//path such that the operation is consistent.
//Such as: case-sensitive act as case-preserving.
//
//The target dir must have been created
func (c *Collection) GetAppliedTo(target string) (deletes []Node, changedfiles []Node, dirs []Node, err error) {
	t, err := NewRootNode(target)
	if err != nil {
		return
	}

	exist := NewCollection() //Collect nodes from target
	err = exist.Add(t)
	if err != nil {
		return
	}

	for name, nodes := range c.nodes {
		en := exist.nodes[name]
		/*
			en :| file | dir | none
			c. (last)
			file| U | R | U
			del | D | D | X
			dir | R | A | A

			* del only means delete a file,
			  delete a dir means files inside dir should be deleted

			U : update -> changedfiles (and deletes other capitalizations)
			D : delete -> deletes
			R : replace -> D + U (case 2) or D + A (case 7)
			A : add -> dirs
			X : no-op
		*/
		last := nodes[len(nodes)-1]
		updated := false
		for _, e := range en {
			if last.IsDir() && e.IsDir() {
				continue //case 8
			}

			if !last.IsDir() && !e.IsDir() && !last.IsDelete() &&
				last.info.Name() == e.info.Name() {
				//case 1 with same capitalization
				if last.SameData(e) {
					updated = true //avoid updating an updated file
				}
				continue
			}
			deletes = append(deletes, e) // all others delete exisiting
		}

		if last.IsDir() {
			dirs = append(dirs, last) //finish cases 7, 8, and 9
		} else if !last.IsDelete() && !updated {
			changedfiles = append(changedfiles, last) //finish cases 1, 2, and 3
		}
	}
	return
}
