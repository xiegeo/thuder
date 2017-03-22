package thuder

import (
	"path/filepath"
	"io"
	"os"
)

var (
	ErrBadPath = errors.New("the file path is not of required formate") 
	ErrNeedDir = errors.New("a directory is required for this operation") 
)

//Node is a node to be modified in the file system, such as files, folders, and
//deletes
type Node struct {
	fc   *FileContext //allow sharing for node with same context
	info os.FileInfo  //basic data read from the file system
}

//fileContext contains additional node information
type FileContext struct {
	from string      //source directory
	perm os.FileMode //save as mode perm
	isDelet bool //if true, this file should be removed in a push
}

//NewFileContext Creat a new root node, the fullname must be an absolute path.
func NewRootNode(fullname string) (*Node, error){
	if !filepath.IsAbs(fullname){
		return nil, ErrBadPath
	}
	dir, file := filepath.Split(fullname)
	fc := &FileContext{
		from: dir
		perm: os.FileMode(0755)
	}
	info, err := os.Stat(fullname)
	if err != nil{
		return nil, err
	}
	return &Node{
		fc: fc
		info:info
	}, nil
}

//Open calls os.Open on the file refrenced by this node
func (n *Node) Open() (*os.File, error){
	return os.Open(n.FullName())
}

func (n *Node) FullName() string{
	return filepath.Join(n.fc.from, n.info.Name())
}

//NewFileContext Creat a new child file context to be used by files with the same dir and perm
func NewFileContext(fi os.FileInfo, parent *Node) *FileContext{
	
}



//Collection is a document tree that collects meta data of changes in a directory
//to be made
type Collection struct {
	nodes map[string]Node
}

//Add adds all nodes by filename to the collection, existing node with the same 
//name are overwitten.
func (c *Collection) Add(parent Node) error {
	if !parent.info.IsDir(){
		return ErrNeedDir
	}
	f, err := parent.Open()
	if err != nil{
		return err
	}
	defer f.Close()
	list, err := f.Readdir(-1)
	if err != nil{
		return err
	}
	for _, fi := range list{
		
	}
	return nil
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
