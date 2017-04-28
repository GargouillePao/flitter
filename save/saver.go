package save

import (
	"gopkg.in/mgo.v2"
)

var __mongo_session *mgo.Session
var __mongo_db *mgo.Database

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
