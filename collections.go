package thuder

import (
	"path/filepath"
	"io"
	"os"
)

var (
	ErrBadPath = errors.New("the file path is not of required formate") 
)

//Node is a node to be modified in the file system, such as files, folders, and
//deletes
type Node struct {
	fc   *FileContext //allow sharing for node with same context
	info os.FileInfo
}

//fileContext contains addition node information
type FileContext struct {
	from string      //source directory
	perm os.FileMode //save as mode perm
}

//NewFileContext Creat a new root node, the fullname must be an absolute path,
//but the file does not need to exist
func NewRootNode(fullname string) (*FileContext, error){
	if !filepath.IsAbs(fullname){
		return nil, ErrBadPath
	}
	dir, file := filepath.Split(fullname)
	fc := &FileContext{
		from dir
	}
}

//NewFileContext Creat a new child file context to be used by files with the same dir and perm
func NewFileContext(fi os.FileInfo, parent *Node) *FileContext{
	
}



//Collection is a document tree that collects meta data of changes in a directory
//to be made
type Collection struct {
	nodes map[string]Node
}

func (c *Collection) Collect(d DirReader) error {

}

//DirReader can list os.FileInfo, as implemented by os.File
type DirReader interface {
	Readdir(n int) ([]FileInfo, error)
}



type PullJob struct {
	Source string //source path
	Target string //target path
}

func (p *PullJob) Do() error {
	
	
	
	fi, err := os.Stat(dir)
	if err != nil {
		return
	}
	
	os.Open(name string)
	
	c := Collection{}
	c.Collect()
}

func NewNode(dir string, fi os.FileInfo, parent *Node) (*Node, error) {
	fi, err := os.Stat(filepath.Join(dir, fi.Name()))
	if err != nil{
		return error
	}
	Node{}
}
