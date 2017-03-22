package thuder

import (
	"os"
	"path/filepath"
)

//Node is a node to be modified in the file system, such as files, folders, and
//deletes
type Node struct {
	fc   *FileContext //allow sharing for node with same context
	info os.FileInfo  //basic data read from the file system
}

//fileContext contains additional node information
type FileContext struct {
	from    string      //source directory
	perm    os.FileMode //save as mode perm
	isDelet bool        //if true, this file should be removed in a push
}

//NewFileContext Creat a new root node, the fullname must be an absolute path.
func NewRootNode(fullname string) (*Node, error) {
	if !filepath.IsAbs(fullname) {
		return nil, ErrBadPath
	}
	dir, _ := filepath.Split(fullname)
	fc := &FileContext{
		from: dir,
		perm: os.FileMode(0755),
	}
	info, err := os.Stat(fullname)
	if err != nil {
		return nil, err
	}
	return &Node{
		fc:   fc,
		info: info,
	}, nil
}

//Open calls os.Open on the file refrenced by this node
func (n *Node) Open() (*os.File, error) {
	return os.Open(n.FullName())
}

func (n *Node) FullName() string {
	return filepath.Join(n.fc.from, n.info.Name())
}

//NewFileContext Creat a new child file context to be used by files with the same dir and perm
func NewFileContext(parent *Node) *FileContext {
	return &FileContext{
		from: parent.FullName(),
		perm: parent.fc.perm,
	}
}
