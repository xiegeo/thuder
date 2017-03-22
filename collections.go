package thuder

import (
	"errors"
	"os"
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

//DirReader can list os.FileInfo, as implemented by os.File
type DirReader interface {
	Readdir(n int) ([]os.FileInfo, error)
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
