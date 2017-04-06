package thuder

import (
	"fmt"
	"os"
	"path/filepath"

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
		err = afero.WriteFile(mfs, filepath.Join(dir, name), []byte(name), fm)
		if err != nil {
			t.Fatal(err)
		}
	}
	afero.Walk(mfs, root, func(path string, info os.FileInfo, err error) error {
		t.Log(path, info.Name())
		return nil
	})

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
	}

}
