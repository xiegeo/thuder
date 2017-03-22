package thuder

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrBadPath = errors.New("the file path is not of required formate")
	ErrNeedDir = errors.New("a directory is required for this operation")
)

//Collection is a document tree that collects meta data of changes in a directory
//to be made
type Collection struct {
	nodes map[string]Node
}

func NewCollection() *Collection {
	return &Collection{
		nodes: make(map[string]Node),
	}
}

func (c *Collection) PrintTo(f func(format string, args ...interface{})) {
	for k, v := range c.nodes {
		f("%s: %s\n", k, &v)
	}
}

//Add adds all nodes by filename to the collection, existing node with the same
//name are overwitten.
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
	for _, fi := range list {
		c.nodes[fi.Name()] = Node{
			fc:   fc,
			info: fi,
		}
	}
	return nil
}

//GetAppliedTo returns list of nodes as actions to be taken on the target
//path such that the operation is consistant.
//Such as: case-sensitive act as case-perserving.
//
//The target dir must have been created
func (c *Collection) GetAppliedTo(target string) ([]Node, error) {
	if !filepath.IsAbs(target) {
		return nil, ErrBadPath
	}
	fi, err := os.Stat(target)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, ErrNeedDir
	}

	return nil, nil //todo
}

//pathCompare returns the more "specific" Node by path, to see which should be
//choosen for conflictes. Only called if the two nodes overwrite each other.
func pathCompare(a, b *Node) *Node {
	adl := len(a.fc.from)
	bdl := len(b.fc.from)
	if adl > bdl {
		return a
	}
	if adl < bdl {
		return b
	}
	c := strings.Compare(a.FullName(), b.FullName())
	if c < 0 {
		return a
	}
	return b
}

type PullJob struct {
	Source string //source path
	Target string //target path
}

/*
func (p *PullJob) Do() error {



	fi, err := os.Stat(dir)
	if err != nil {
		return err
	}

	os.Open(name string)

	c := Collection{}
	c.Collect()

	return nil
}

func ChildNodes(dir string, fi os.FileInfo, parent *Node) ([]Node, error) {
	fi, err := os.Stat(filepath.Join(dir, fi.Name()))
	if err != nil{
		return nil, err
	}
	retu &Node{
	fc   *FileContext //allow sharing for node with same context
	info fi
	}
}
*/
