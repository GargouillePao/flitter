package data

import (
	"github.com/gargous/flitter/core"
	"testing"
)

func Test_NodeSaver(t *testing.T) {
	err := OpenMongo("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer CloseMongo()
	var tree core.NodeTree
	tree = core.NewNodeTree()
	saver := NewNodeTreeSaver(tree)
	err = saver.Load()
	t.Log(tree)
	if err != nil {
		t.Log(err)
	}
	tree.Search("127.0.0.1:1000")
	tree.Add("127.0.0.1:1000")
	saver.Save(SA_Add, "127.0.0.1:1000")
	tree.Add("127.0.0.1:1000")
	saver.Save(SA_Add, "127.0.0.1:1000")
	tree.Add("127.0.0.1:1000")
	saver.Save(SA_Add, "127.0.0.1:1000")
	tree.Add("127.0.0.1:1000")
	saver.Save(SA_Add, "127.0.0.1:1000")
	t.Log(err, tree)
}
