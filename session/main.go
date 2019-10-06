package session

import (
	"errors"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const _EmptyString string = ""

var (
	ErrPlayerNotFound  error = errors.New("Player Not Found")
	ErrAccountNotFound       = errors.New("Account Not Found")
	ErrAccountInvalid        = errors.New("Account Invalid")
)

type LoginInfo struct {
	Visit    bool
	Token    string
	Passport string
	Password string
	id       uint32
	Role     interface{}
}

func (l *LoginInfo) setIdFromRole() {
	idField := reflect.ValueOf(l.Role).FieldByName("Id")
	id := uint32(idField.Uint())
	l.id = id
}

func (l *LoginInfo) setIdToRole() {
	idField := reflect.ValueOf(l.Role).FieldByName("Id")
	idField.SetUint(uint64(l.id))
}

type Session struct {
	info     LoginInfo
	expireAt time.Time
	tokenLen int
}

func (s *Session) UpdateToken() {
	s.expireAt = time.Now()
	b := strings.Builder{}
	for i := 0; i < s.tokenLen; i++ {
		b.WriteByte(byte(10 + rand.Intn(88)))
	}
	s.info.Token = b.String()
}

func (s *Session) hasExpired(info LoginInfo) bool {
	return s.info.Token != info.Token ||
		s.expireAt.Before(time.Now())
}

type MongoSessParam struct {
	C   string
	URL string
}

type ManagerParam struct {
	TokenLength  int
	MongoRole    MongoSessParam
	MongoAccount MongoSessParam
}

type MongoSessInfo struct {
	s *mgo.Session
	c *mgo.Collection
}

type Manager struct {
	sess       sync.Map
	mutex      sync.Mutex
	expireTime time.Duration
	dbRole     MongoSessInfo
	dbAccount  MongoSessInfo
	param      ManagerParam
}

func NewManager(param ManagerParam) *Manager {
	return &Manager{
		param: param,
	}
}

func (m *Manager) dialMongo(url string, cname string) (s *mgo.Session, c *mgo.Collection, err error) {
	info, err := mgo.ParseURL(url)
	if err != nil {
		return
	}
	s, err = mgo.DialWithInfo(info)
	if err != nil {
		return
	}
	c = s.DB(info.Database).C(cname)
	return
}

func (m *Manager) Init() (err error) {
	m.dbAccount.s, m.dbAccount.c, err = m.dialMongo(m.param.MongoAccount.URL, m.param.MongoAccount.C)
	if err != nil {
		return
	}
	m.dbRole.s, m.dbRole.c, err = m.dialMongo(m.param.MongoRole.URL, m.param.MongoRole.C)
	if err != nil {
		return
	}
	return
}

func (m *Manager) Close() {
	if m.dbAccount.s != nil {
		m.dbAccount.s.Close()
	}
	if m.dbRole.s != nil {
		m.dbRole.s.Close()
	}
}

func (m *Manager) createSession(info LoginInfo) (*Session, error) {
	s := &Session{
		info:     info,
		expireAt: time.Now().Add(m.expireTime),
		tokenLen: m.param.TokenLength,
	}
	s.UpdateToken()
	err := m.dbAccount.c.FindId(info.id).One(&info.Role)
	if err != nil {
		return s, nil
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for i := 0; i < 6; i++ {
		info.id = rand.Uint32()
		(&info).setIdToRole()
		err := m.dbRole.c.Insert(info.Role)
		if err == nil {
			m.sess.Store(info.id, s)
			return s, nil
		}
	}
	return nil, ErrPlayerNotFound
}
func (m *Manager) deleteSession(id uint32) {
	m.dbRole.c.RemoveId(id)
	m.sess.Delete(id)
}
func (m *Manager) getSession(id uint32) *Session {
	v, ok := m.sess.Load(id)
	if ok {
		s, ok := v.(*Session)
		if ok {
			return s
		}
	}
	return nil
}

type AccountInfo struct {
	Passport string `bson:"passport"`
	Password string `bson:"password"`
}

func (m *Manager) validateAccount(info LoginInfo) error {
	passport := _EmptyString
	password := _EmptyString
	(&info).setIdFromRole()
	s := m.getSession(info.id)
	if s == nil {
		acountInfo := AccountInfo{}
		err := m.dbAccount.c.FindId(info.id).One(&acountInfo)
		if err != nil {
			return ErrAccountNotFound
		}
		passport = acountInfo.Passport
		password = acountInfo.Password
	} else {
		passport = s.info.Passport
		password = s.info.Password
	}
	if passport == info.Passport &&
		password == info.Password {
		return nil
	}
	return nil
}

func (m *Manager) saveAccount(info LoginInfo) error {
	return m.dbAccount.c.UpdateId(info.Passport, bson.M{
		"$set": bson.M{
			"roles." + strconv.Itoa(int(info.id)): info.Role,
		},
	})
}

func (m *Manager) Login(info LoginInfo) (*Session, error) {
	(&info).setIdFromRole()
	if info.Visit {
		s := m.getSession(info.id)
		if s != nil && !s.hasExpired(info) {
			return s, nil
		}
		return m.createSession(info)
	} else {
		err := m.validateAccount(info)
		if err != nil {
			return nil, err
		}
		s, err := m.createSession(info)
		if err != nil {
			return nil, err
		}
		err = m.saveAccount(s.info)
		if err != nil {
			m.deleteSession(s.info.id)
			return nil, err
		}
		return s, nil
	}
}
