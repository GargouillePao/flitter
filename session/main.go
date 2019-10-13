package session

import (
	"errors"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"

	"gopkg.in/mgo.v2"
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
	l.id = uint32(idField.Uint())
}

func (l *LoginInfo) setIdToRole() {
	idField := reflect.ValueOf(l.Role).FieldByName("Id")
	idField.SetUint(uint64(l.id))
}

type Session struct {
	info     LoginInfo
	expireAt time.Time
	tokenLen int
	login    bool
	m        *Manager
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

func (s *Session) Login() error {
	if s.info.Visit {
		s.login = true
		return nil
	} else {
		err := s.m.validateAccount(s.info)
		if err != nil {
			return err
		}
		s.login = true
		return nil
	}
}

func (s *Session) Logout() {
	s.m.quitSession(s.info.id)
}

func (s *Session) Role(role interface{}) error {
	return nil
}

func (s *Session) Roles(roles interface{}) error {
	return nil
}

type MongoSessParam struct {
	C   string
	URL string
}

type MongoSessInfo struct {
	s *mgo.Session
	c *mgo.Collection
}

type ManagerParam struct {
	TokenLength  int
	MongoRole    MongoSessParam
	MongoAccount MongoSessParam
}

var DefaultParam = func(db string) ManagerParam {
	return ManagerParam{
		TokenLength: 16,
		MongoRole: MongoSessParam{
			C:   "role",
			URL: "mongodb://127.0.0.1:27017/" + db,
		},
		MongoAccount: MongoSessParam{
			C:   "account",
			URL: "mongodb://127.0.0.1:27017/" + db,
		},
	}
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
	m := &Manager{
		param: param,
	}
	return m
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
		m:        m,
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
	m.quitSession(id)
	m.dbRole.c.RemoveId(id)
}

func (m *Manager) quitSession(id uint32) {
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
	if info.Passport == _EmptyString || info.Password == _EmptyString {
		return ErrAccountInvalid
	}
	acountInfo := AccountInfo{}
	err := m.dbAccount.c.FindId(info.id).One(&acountInfo)
	if err != nil {
		return ErrAccountNotFound
	}
	if acountInfo.Passport == info.Passport &&
		acountInfo.Password == info.Password {
		return nil
	}
	return ErrAccountInvalid
}

func (m *Manager) Get(info LoginInfo) (*Session, error) {
	(&info).setIdFromRole()
	s := m.getSession(info.id)
	if s != nil && !s.hasExpired(info) {
		return s, nil
	}
	return m.createSession(info)
}
