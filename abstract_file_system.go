package thuder

import (
	"io"
	"os"
	"time"
)

// Allows file system calls to be replaced during testing

// File represents a file in the filesystem. (from spf13/afero/afero.go, take
// just what we need to allow testing holks)
type File interface {
	io.ReadWriteCloser
	//io.ReaderAt
	//io.Seeker
	//io.WriterAt

	//Name() string
	Readdir(count int) ([]os.FileInfo, error)
	//Readdirnames(n int) ([]string, error)
	Stat() (os.FileInfo, error)
	//Sync() error
	//Truncate(size int64) error
	WriteString(s string) (ret int, err error)
}

var fs = newOsFs() //for overide in tests

// Fs is the filesystem interface.
//
// Any simulated or real filesystem should implement this interface.
type Fs interface {
	// Create creates a file in the filesystem, returning the file and an
	// error, if any happens.
	Create(name string) (File, error)

	// Mkdir creates a directory in the filesystem, return an error if any
	// happens.
	Mkdir(name string, perm os.FileMode) error

	// MkdirAll creates a directory path and all parents that does not exist
	// yet.
	MkdirAll(path string, perm os.FileMode) error

	// Open opens a file, returning it or an error, if any happens.
	Open(name string) (File, error)

	// OpenFile opens a file using the given flags and the given mode.
	OpenFile(name string, flag int, perm os.FileMode) (File, error)

	// Remove removes a file identified by name, returning an error, if any
	// happens.
	Remove(name string) error

	// RemoveAll removes a directory path and any children it contains. It
	// does not fail if the path does not exist (return nil).
	RemoveAll(path string) error

	// Rename renames a file.
	Rename(oldname, newname string) error

	// Stat returns a FileInfo describing the named file, or an error, if any
	// happens.
	Stat(name string) (os.FileInfo, error)

	// The name of this FileSystem
	Name() string

	//Chmod changes the mode of the named file to mode.
	Chmod(name string, mode os.FileMode) error

	//Chtimes changes the access and modification times of the named file
	Chtimes(name string, atime time.Time, mtime time.Time) error
}

// osFs is a Fs implementation that uses functions provided by the os package.
//
// For details in any method, check the documentation of the os package
// (http://golang.org/pkg/os/).
type osFs struct{}

func newOsFs() Fs {
	return &osFs{}
}

func (osFs) Name() string { return "OsFs" }

func (osFs) Create(name string) (File, error) {
	f, e := os.Create(name)
	if f == nil {
		// while this looks strange, we need to return a bare nil (of type nil) not
		// a nil value of type *os.File or nil won't be nil
		return nil, e
	}
	return f, e
}

func (osFs) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (osFs) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (osFs) Open(name string) (File, error) {
	f, e := os.Open(name)
	if f == nil {
		// while this looks strange, we need to return a bare nil (of type nil) not
		// a nil value of type *os.File or nil won't be nil
		return nil, e
	}
	return f, e
}

func (osFs) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	f, e := os.OpenFile(name, flag, perm)
	if f == nil {
		// while this looks strange, we need to return a bare nil (of type nil) not
		// a nil value of type *os.File or nil won't be nil
		return nil, e
	}
	return f, e
}

func (osFs) Remove(name string) error {
	return os.Remove(name)
}

func (osFs) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (osFs) Rename(oldname, newname string) error {
	return os.Rename(oldname, newname)
}

func (osFs) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (osFs) Chmod(name string, mode os.FileMode) error {
	return os.Chmod(name, mode)
}

func (osFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.Chtimes(name, atime, mtime)
}
