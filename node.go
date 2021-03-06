package thuder

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//Node is a node to be modified in the file system, such as files, folders, and
//deletes
type Node struct {
	fc   *FileContext //allow sharing for node with same context
	info os.FileInfo  //basic data read from the file system
}

//NewRootNode Creat a new root node, the fullname must be an absolute path.
//isDelete sets the isDelete property.
func NewRootNode(fullname string, isDelete bool) (*Node, error) {
	if !fs.IsAbs(fullname) {
		return nil, ErrBadPath
	}
	fullname = filepath.Clean(fullname)
	dir, _ := filepath.Split(fullname)
	fc := &FileContext{
		from:     dir,
		perm:     os.FileMode(0755),
		isDelete: isDelete,
	}
	info, err := fs.Stat(fullname)
	if err != nil {
		return nil, err
	}
	return &Node{
		fc:   fc,
		info: info,
	}, nil
}

//Open calls os.Open on the file or dir for reading as referenced by this node
func (n Node) Open() (File, error) {
	return fs.Open(n.FullName())
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

//IsDir returns if it is a directory
func (n Node) IsDir() bool {
	return n.info.IsDir()
}

//SameDir returns if two nodes have the same parent dir path
func (n Node) SameDir(n2 Node) bool {
	return n.fc.from == n2.fc.from
}

//ModTime returns the last modified time of the file represented by this node
func (n Node) ModTime() time.Time {
	return n.info.ModTime()
}

//SameData returns if two files have the same data based on size and modtime,
// panics if either is a dir
//Todo: File mode changes are tracked to propergate mode changes too?
func (n Node) SameData(n2 Node) bool {
	if n.IsDir() || n2.IsDir() {
		panic(fmt.Sprintf("SameData can not be used for %v, %v", n, n2))
	}

	return n.info.Size() == n2.info.Size() &&
		EqualFileModTime(n.ModTime(), n2.ModTime())
}

//EqualFileModTime test if two file last mod times from different file system
//are equal.
//
//supports:
//  FAT: 2 sec, (ceiling for windows, floor for linux)
//	others: 1 sec
//	and any subsec: if diff less than 0.5 sec
func EqualFileModTime(t1, t2 time.Time) bool {
	if t1.Equal(t2) {
		return true
	}
	if t1.After(t2) {
		t1, t2 = t2, t1 //make t1 the smaller time
	}
	if !t1.Add(2 * time.Second).After(t2) { //t1 + 2s <= t2
		// 2 sec or more diff, they are different
		return false
	}
	if t1.Nanosecond() == 0 || t2.Nanosecond() == 0 {
		// one time have no sub sec precision
		if t1.Second()%2 == 0 || t2.Second()%2 == 0 {
			return true
		}
		return t1.Add(time.Second).After(t2) // t1 + 1s > t2
	}
	return t1.Add(time.Second / 2).After(t2) // t1 + .5s > t2
}

//FileContext contains additional node information
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
//	1) files before dirs
//	2) dirs by insertion order (later added files from different dirs can overwrite earlier once)
//	3) files in same dir ordered by case (so there is one consitant winner)
func addNode(ns []Node, n Node) []Node {
	if len(ns) == 0 {
		return append(ns, n) //base case, only used in tests
	}
	if n.IsDir() {
		return append(ns, n) // 1 and 2
	}

	index := len(ns) - 1
	last := ns[index]

	// 3
	c := strings.Compare(last.info.Name(), n.info.Name())
	if c >= 0 {
		return append(ns, n)
	}
	ns[index] = n
	return append(ns, last)

}
