package data

import (
	"fmt"
	core "github.com/gargous/flitter/core"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var __mongo_session *mgo.Session
var __mongo_db *mgo.Database

type SaveAction uint8

const (
	_ SaveAction = iota
	SA_Add
	SA_Remove
	SA_Clean
)

func OpenMongo(addr string, db string) (err error) {
	__mongo_session, err = mgo.Dial(addr)
	if err != nil {
		return
	}
	__mongo_session.SetMode(mgo.Monotonic, true)
	__mongo_db = __mongo_session.DB(db)
	return
}

func CloseMongo() {
	__mongo_session.Close()
}

type NodeTreeSaver interface {
	Load() error
	Save(action SaveAction, nodeAddr string) error
}

type nodeTreeSaver struct {
	id         string
	collection *mgo.Collection
	tree       core.NodeTree
}

type nodeSaveItem struct {
	Name     string
	Action   SaveAction
	NodeAddr string
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
			n.tree.Add(actions[i].NodeAddr)
		case SA_Remove:
			n.tree.Remove(actions[i].NodeAddr)
		}
	}
	return
}
func (n *nodeTreeSaver) Save(action SaveAction, nodeAddr string) error {
	if action == SA_Clean {
		return n.collection.Remove(bson.M{"name": n.id})
	}
	return n.collection.Insert(nodeSaveItem{Name: n.id, Action: action, NodeAddr: nodeAddr})
}
