package save

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
		t.Fatal(err)
	}
	info, ok := tree.SearchForName("root")
	if !ok {
		t.Fatal("No searched")
	}
	t.Log(info)
	tree.Add("root@127.0.0.1:1000")
	saver.SaveLastItem(SA_Add)
	tree.Add("login@127.0.0.1:2000")
	saver.SaveLastItem(SA_Add)
	tree.Add("scenes@127.0.0.1:3000")
	saver.SaveLastItem(SA_Add)
	tree.Add("scenes1@127.0.0.1:5000")
	saver.SaveLastItem(SA_Add)
	t.Log(err, tree)
}
