package thuder

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"testing"

	"github.com/spf13/afero"
)

//waf wraps afero.Fs into the locally defined Fs
type waf struct{ afero.Fs }

func wrapAfero(f afero.Fs) Fs {
	return waf{f}
}

func (w waf) IsAbs(path string) bool {
	return true //we are only using absolute for these tests
}

func (w waf) Create(name string) (File, error) {
	return w.Fs.Create(name)
}

func (w waf) Open(name string) (File, error) {
	return w.Fs.Open(name)
}

func (w waf) OpenFile(name string, flag int, mode os.FileMode) (File, error) {
	return w.Fs.OpenFile(name, flag, mode)
}

func testCopied(a, b string) error {
	fa, err := fs.Open(a)
	if err != nil {
		return err
	}
	ia, _ := fa.Stat()
	fb, err := fs.Open(b)
	if err != nil {
		return err
	}
	ib, _ := fb.Stat()
	if ia.IsDir() != ib.IsDir() {
		return errors.New("not same type")
	}
	if ia.Size() != ib.Size() {
		return errors.New("not same size")
	}
	if ia.Mode() != ib.Mode() {
		return errors.New("not same file mode")
	}
	if ia.ModTime() != ib.ModTime() {
		return errors.New("not same modification time")
	}
	da, _ := ioutil.ReadAll(fa)
	db, _ := ioutil.ReadAll(fb)
	if !bytes.Equal(da, db) {
		return errors.New("not same data")
	}
	return nil
}

func TestProcessor(t *testing.T) {
	fs2 := fs
	defer func() { fs = fs2 }()

	mfs := afero.NewMemMapFs().(*afero.MemMapFs)
	fs = wrapAfero(mfs)
	fm := os.FileMode(0755)
	dirs := []string{
		"a", "b",
		"a/a", "b/b",
		"a/c", "b/c",
		"a/D", "b/d",
		"t",
	}
	root := "/"
	if os.PathSeparator == '\\' {
		root = "\\" //should be c:\, but bug in afero
	}
	for i, dir := range dirs {
		dir = root + dir
		dirs[i] = dir
		err := fs.Mkdir(dir, fm)
		if err != nil {
			t.Fatal(err)
		}
		name := fmt.Sprintf("n%v", i)
		fullName := filepath.Join(dir, name)
		err = afero.WriteFile(mfs, fullName, []byte(name), fm)
		if err != nil {
			t.Fatal(err)
		}
		mt := time.Unix(int64(i*10), 0)
		fs.Chtimes(fullName, mt, mt)

	}

	var sources []Node
	for _, fullname := range dirs[:2] {
		rootNode, err := NewRootNode(fullname)
		if err != nil {
			t.Fatal(fullname, err)
		}
		sources = append(sources, *rootNode)
	}

	actions := make(chan action)
	p := Processor{
		stack: []layer{
			layer{from: sources, to: root + "t"},
		},
		actions: actions,
	}
	go p.Do()
	for {
		a := <-actions
		if len(a.from) == 0 {
			break
		}
		t.Log(a.from, a.to)
		err := applyAction(a)
		if err != nil {
			t.Error(err)
		}
	}
	err := testCopied(root+"b/d/n7", root+"t/D/n7")
	if err != nil {
		t.Error("copy expected: ", err)
	}
	// reset stack and redo
	p.stack = []layer{layer{from: sources, to: root + "t"}}
	go p.Do()
	for {
		a := <-actions
		if len(a.from) == 0 {
			break
		}
		if !a.from[0].IsDir() {
			t.Error("file operation: ", a.from, " should not be redone.")
		}
	}
	afero.Walk(mfs, root, func(path string, info os.FileInfo, err error) error {
		t.Log(path, info.Name(), info.Size(), info.ModTime())
		return nil
	})
}
