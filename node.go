package thuder

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

//Node is a node to be modified in the file system, such as files, folders, and
//deletes
type Node struct {
	fc   *FileContext //allow sharing for node with same context
	info os.FileInfo  //basic data read from the file system
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
func (n Node) Open() (*os.File, error) {
	return os.Open(n.FullName())
}

//FullName returns the absolute path to find the node
func (n Node) FullName() string {
	return filepath.Join(n.fc.from, n.info.Name())
}

//String returns printable node for debugging
func (n Node) String() string {
	if n.info.IsDir() {
		return fmt.Sprintf("Dir %s %s %v", n.fc, n.info.Name(), n.info.ModTime())
	}
	return fmt.Sprintf("File %s %s %.2fkb %v", n.fc, n.info.Name(), float64(n.info.Size())/1024, n.info.ModTime())
}

//IsDelete returns if the current node should be deleted at the target
func (n Node) IsDelete() bool {
	return n.fc.isDelete
}

//FileMode returns the file mode (ie 0755) used for saving
func (n Node) FileMode() os.FileMode {
	return n.fc.perm
}

//IsDelete returns if it is a directory
func (n Node) IsDir() bool {
	return n.info.IsDir()
}

func (n Node) SameDir(n2 Node) bool {
	return n.fc.from == n2.fc.from
}

//fileContext contains additional node information
type FileContext struct {
	from     string      //source directory
	perm     os.FileMode //save as mode perm
	isDelete bool        //if true, this file should be removed in a push
}

//NewFileContext Creat a new child file context to be used by files with the same dir and perm
func NewFileContext(parent *Node) *FileContext {
	return &FileContext{
		from:     parent.FullName(),
		perm:     parent.fc.perm,
		isDelete: parent.IsDelete(),
	}
}

//String prints out FileContext for debugging
func (c *FileContext) String() string {
	if c.isDelete {
		return fmt.Sprintf("Delete (%s)", c.from)
	}
	return fmt.Sprintf("0%s (%s)", strconv.FormatUint(uint64(c.perm), 8), c.from)
}

//addNode with ordering
//	ordering: (later is more important, and only the last one is ordered)
//	- files before dirs
//	- dirs by insertion order (later added dirs can overwrite earlier once)
//	- files in same dir ordered by case (so there is one consitant winner)
//	- new file from different a dir remove existing files
func addNode(ns []Node, n Node) []Node {
	return ns //todo
}
