package thuder

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var copyBufs = make(chan []byte, 4)

func init() {
	copyBufs <- make([]byte, 1024*1024)
	copyBufs <- make([]byte, 1024*1024)
	copyBufs <- make([]byte, 1024*1024)
	copyBufs <- make([]byte, 1024*1024)
}

func atomicCopy(n Node, to string) error {
	LEDOn()
	defer LEDOff()
	a, err := newAtomicFile(to, n.info.Name(), n.FileMode())
	if err != nil {
		return err
	}
	reader, err := n.Open()
	if err != nil {
		a.Cancle()
		return err
	}
	defer reader.Close()
	buf := <-copyBufs
	_, err = io.CopyBuffer(a, reader, buf)
	copyBufs <- buf
	if err != nil {
		a.Cancle()
		return err
	}
	err = a.CommitRename()
	if err != nil {
		return err
	}
	mt := n.ModTime()
	return fs.Chtimes(filepath.Join(to, n.info.Name()), mt, mt)
}

type afw struct {
	Folder    string
	FinalName string
	File
}

func newAtomicFile(folder, finalName string, perm os.FileMode) (*afw, error) {
	tempPrefix := "thuder_temp_"
	temp, err := tempFile(folder, tempPrefix, perm)
	if err != nil {
		return nil, err
	}
	return &afw{Folder: folder, FinalName: finalName, File: temp}, nil
}

func (a *afw) TempPath() (string, error) {
	tempStat, err := a.Stat()
	if err != nil {
		return "", err
	}
	return filepath.Join(a.Folder, tempStat.Name()), nil
}

func (a *afw) Cancle() {
	path, err := a.TempPath()
	a.Close()
	if err != nil {
		return
	}
	fs.Remove(path)
}

func (a *afw) CommitRename() error {
	tempPath, err := a.TempPath()
	if err != nil {
		return err
	}
	finalPath := filepath.Join(a.Folder, a.FinalName)
	err = a.Close()
	if err != nil {
		return err
	}
	err = fs.Rename(tempPath, finalPath)
	if err != nil {
		fs.Remove(tempPath)
	}
	return err
}

// the following is modified from ioutil

// Random number state.
// We generate random temporary file names so that there's a good
// chance the file doesn't exist yet - keeps the number of tries in
// TempFile to a minimum.
var rand uint32
var randmu sync.Mutex

func reseed() uint32 {
	return uint32(time.Now().UnixNano() + int64(os.Getpid()))
}

func nextSuffix() string {
	randmu.Lock()
	r := rand
	if r == 0 {
		r = reseed()
	}
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rand = r
	randmu.Unlock()
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}

// TempFile creates a new temporary file in the directory dir
// with a name beginning with prefix, opens the file for reading
// and writing, and returns the resulting *os.File.
// If dir is the empty string, TempFile uses the default directory
// for temporary files (see os.TempDir).
// Multiple programs calling TempFile simultaneously
// will not choose the same file. The caller can use f.Name()
// to find the pathname of the file. It is the caller's responsibility
// to remove the file when no longer needed.
func tempFile(dir, prefix string, perm os.FileMode) (f File, err error) {
	if dir == "" {
		//dir = fs.TempDir()
		err = fmt.Errorf("dir is requared for atomic_file_writer")
		return
	}

	nconflict := 0
	for i := 0; i < 10000; i++ {
		name := filepath.Join(dir, prefix+nextSuffix())
		f, err = fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, perm)
		if os.IsExist(err) {
			if nconflict++; nconflict > 10 {
				randmu.Lock()
				rand = reseed()
				randmu.Unlock()
			}
			continue
		}
		break
	}
	return
}
