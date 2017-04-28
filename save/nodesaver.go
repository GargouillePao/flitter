package save

import (
	"fmt"
	core "github.com/gargous/flitter/core"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type NodeTreeSaver interface {
	Load() error
	Save(action SaveAction, npath core.NodePath) error
	SaveLastItem(action SaveAction) error
}

type nodeTreeSaver struct {
	id         string
	recentItem nodeSaveItem
	collection *mgo.Collection
	tree       core.NodeTree
}

type nodeSaveItem struct {
	Name     string
	Action   SaveAction
	NodePath core.NodePath
}

func NewNodeTreeSaver(tree core.NodeTree) NodeTreeSaver {
	saver := &nodeTreeSaver{
		id:   "NodeTree",
		tree: tree,
	}
	if __mongo_db != nil {
		saver.collection = __mongo_db.C("node")
	}
	return saver
}
func (n *nodeTreeSaver) Load() (err error) {
	//action := bson.M{"action": SA_Add, "node": ""}
	var actions []nodeSaveItem
	err = n.collection.Find(bson.M{"name": n.id}).All(&actions)
	if err != nil {
		fmt.Println("Load err", err)
		return
	}

	for i := 0; i < len(actions); i++ {
		switch actions[i].Action {
		case SA_Add:
			n.tree.Add(actions[i].NodePath)
		case SA_Remove:
			n.tree.Remove(actions[i].NodePath)
		}
	}
	return
}
func (n *nodeTreeSaver) Save(action SaveAction, npath core.NodePath) (err error) {
	if action == SA_Clean {
		err = n.collection.Remove(bson.M{"name": n.id})
		return
	}
	err = n.collection.Insert(nodeSaveItem{Name: n.id, Action: action, NodePath: npath})
	return
}
func (n *nodeTreeSaver) SaveLastItem(action SaveAction) (err error) {
	switch action {
	case SA_Add:
		err = n.Save(action, n.tree.GetLastAdd())
	case SA_Remove:
		err = n.Save(action, n.tree.GetLastRemove())
	}
	return
}
