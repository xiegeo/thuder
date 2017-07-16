package thuder

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func acceptAll(n *Node) bool {
	return true
}

func TestGetAppliedTo(t *testing.T) {
	cw, _ := filepath.Abs(".")
	rootN, err := NewRootNode(cw, false)
	if err != nil {
		t.Fatal(err)
	}
	c := NewCollection(acceptAll)
	err = c.Add(rootN)
	if err != nil {
		t.Fatal(err)
	}
	t.Run(fmt.Sprintf("test apply to self should be no-op, except for child dirs"), func(t *testing.T) {
		deletes, changedfiles, _, err := c.GetAppliedTo(cw)
		if err != nil {
			t.Fatal(err)
		}
		if len(deletes)+len(changedfiles) != 0 {
			t.Fatal(deletes, changedfiles, "should all be empty for no-op")
		}
	})

	for _, n := range c.nodes {
		n[0].fc.isDelete = true
		break // fc is shared, so it only need to set once
	}
	t.Run(fmt.Sprintf("test apply of deletes"), func(t *testing.T) {
		deletes, changedfiles, dirs, err := c.GetAppliedTo(cw)
		if err != nil {
			t.Fatal(err)
		}
		if len(changedfiles) != 0 {
			t.Fatal(changedfiles, "should be empty for no-op")
		}
		if len(deletes)+len(dirs) != len(c.nodes) {
			t.Fatal("all should be deletes or dirs", deletes, dirs, c)
		}
	})

	c = NewCollection(acceptAll)
	c.AddList(&FileContext{from: "/overwritten"}, []os.FileInfo{&testFileInfo{name: "AddFile"}})
	c.AddList(&FileContext{from: "/overwritten"}, []os.FileInfo{&testFileInfo{name: "AddDir"}})
	c.AddList(&FileContext{from: "/other"}, []os.FileInfo{&testFileInfo{name: "AddFile"}})
	c.AddList(&FileContext{from: "/other"}, []os.FileInfo{&testFileInfo{name: "AddDir", dir: true}})
	c.AddList(&FileContext{isDelete: true}, []os.FileInfo{&testFileInfo{name: "DeleteMe"}})
	c.AddList(&FileContext{isDelete: true}, []os.FileInfo{&testFileInfo{name: "deleteme"}}) //double delete, to be ignored
	c.AddList(&FileContext{}, []os.FileInfo{&testFileInfo{name: "AddDir", dir: true}})
	c.AddList(&FileContext{isDelete: true}, []os.FileInfo{&testFileInfo{name: "AddDir2", dir: true}})
	c.AddList(&FileContext{isDelete: true}, []os.FileInfo{&testFileInfo{name: "addFILE"}}) //delete of an added, to be ignored
	c.AddList(&FileContext{}, []os.FileInfo{&testFileInfo{name: "AddFile"}})
	c.AddList(&FileContext{}, []os.FileInfo{&testFileInfo{name: "addfile"}})
	c.AddList(&FileContext{isDelete: true}, []os.FileInfo{&testFileInfo{name: "addFILE"}}) //delete of an added, is ignored
	c.PrintTo(t.Logf)
	t.Run(fmt.Sprintf("test apply of a sample collection"), func(t *testing.T) {
		deletes, changedfiles, dirs, err := c.GetAppliedTo(cw)
		if err != nil {
			t.Fatal(err)
		}

		if len(deletes) != 0 {
			t.Fatal(deletes)
		}
		if len(changedfiles) != 2 {
			t.Fatal(changedfiles)
		}
		if len(dirs) != 2 {
			t.Fatal(dirs)
		}

	})
}
