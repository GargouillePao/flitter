package session

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"

	"gopkg.in/mgo.v2"
)

const _EmptyString string = ""

var (
	ErrPlayerNotFound      error = errors.New("Player Not Found")
	ErrDBNotFound                = errors.New("DB Not Found")
	ErrAccountNotFound           = errors.New("Account Not Found")
	ErrAccountInvalid            = errors.New("Account Invalid")
	ErrAuthInvalidURL            = errors.New("Auth Invalid URL")
	ErrAuthGetNothing            = errors.New("Auth Get Nothing")
	ErrAccountGenrateLimit       = errors.New("Account Genrate Limited")
)

type Session struct {
	Info     AccountInfo
	expireAt time.Time
	tokenLen int
	m        *Manager
	OnKick   func()
	using    uint64
}

func (s *Session) UseRole(id uint64) (role RoleInfo, err error) {
	role, err = s.m.getRole(id)
	if err != nil {
		return
	}
	s.using = id
	return
}

func (s *Session) CreateRole(role proto.Message) uint64 {
	id := s.m.createRole(&s.Info, role)
	if id > 0 {
		s.using = id
	}
	return id
}

func (s *Session) DeleteRole(id uint64) {
	s.m.deleteRole(&s.Info, id)
	if s.using == id {
		s.using = 0
	}
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
	AuthURLs     map[string]string
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

type AuthType byte

const (
	AT_VISIT AuthType = iota
	AT_PASSWORD
	AT_AUTH2
	AT_OIDC
)

type AccountInfo struct {
	Id          string     `bson:"_id"`
	Platform    string     `bson:"platform"`
	Type        AuthType   `bson:"type"`
	UserId      string     `bson:"user_id"`
	Password    string     `bson:"password"`
	AccessToken string     `bson:"access_token"`
	IdToken     string     `bson:"id_token"`
	Roles       []RoleInfo `bson:"roles"`
}

func (a AccountInfo) genId() string {
	b := strings.Builder{}
	b.WriteString(a.Platform)
	b.WriteString("(")
	b.WriteString(a.UserId)
	b.WriteString(")")
	return b.String()
}

func (a AccountInfo) randUserID() AccountInfo {
	b := strings.Builder{}
	for i := 0; i < 18; i++ {
		switch rand.Intn(3) {
		case 0:
			b.WriteByte(byte(48 + rand.Intn(10)))
		case 1:
			b.WriteByte(byte(65 + rand.Intn(26)))
		case 2:
			b.WriteByte(byte(97 + rand.Intn(26)))
		}
	}
	a.UserId = b.String()
	a.Id = a.genId()
	return a
}

type Manager struct {
	sess       sync.Map
	mutex      sync.Mutex
	expireTime time.Duration
	dbRole     MongoSessInfo
	dbAccount  MongoSessInfo
	param      ManagerParam
	httpC      *http.Client
}

func NewManager(param ManagerParam) *Manager {
	m := &Manager{
		param: param,
		httpC: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
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

func (m *Manager) deleteSession(id uint32) {
	m.quitSession(id)
	m.dbRole.c.RemoveId(id)
}

func (m *Manager) quitSession(id uint32) {
	m.sess.Delete(id)
}
func (m *Manager) getSession(id uint32) (s *Session) {
	v, ok := m.sess.Load(id)
	if ok {
		s, ok := v.(*Session)
		if ok {
			return s
		}
	}
	return nil
}

func (m *Manager) authGet(platform string, param map[string]string) (jData map[string]interface{}, err error) {
	authURL, ok := m.param.AuthURLs[platform]
	if !ok || authURL == _EmptyString {
		err = ErrAuthInvalidURL
		return
	}
	values := url.Values{}
	for k, v := range param {
		values.Add(k, v)
	}
	s := strings.Builder{}
	s.WriteString(authURL)
	s.WriteString("?")
	s.WriteString(values.Encode())
	resp, err := m.httpC.Get(s.String())
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = ErrAuthGetNothing
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &jData)
	return
}

func (m *Manager) newSession(a AccountInfo) (s *Session) {
	s = &Session{
		Info:   a,
		m:      m,
		OnKick: func() {},
	}
	return
}

func (m *Manager) get(a AccountInfo) (s *Session, err error) {
	if v, ok := m.sess.Load(a.Id); ok {
		s = v.(*Session)
		return
	}
	c := m.dbAccount.c
	if c != nil {
		s = m.newSession(a)
		err = c.FindId(a.Id).One(&s.Info)
		if err != nil {
			return
		}
		m.sess.Store(a.Id, s)
	} else {
		err = ErrDBNotFound
	}
	return
}

func (m *Manager) create(a AccountInfo) (s *Session, err error) {
	c := m.dbAccount.c
	if c != nil {
		for i := 0; i < 20; i++ {
			a = a.randUserID()
			err = c.Insert(a)
			if err == nil {
				s = m.newSession(a)
				return
			}
		}
		err = ErrAccountGenrateLimit
		return
	} else {
		err = ErrDBNotFound
	}
	return
}

type RoleInfo struct {
	Id        uint64        `bson:"_id"`
	AccountId string        `bson:"account_id"`
	Brief     proto.Message `bson:"brief"`
	Detail    proto.Message `bson:"detail"`
}

func (m *Manager) createRole(a *AccountInfo, brief proto.Message) uint64 {
	rc := m.dbRole.c
	ac := m.dbAccount.c
	role := RoleInfo{
		AccountId: a.Id,
		Brief:     brief,
	}
	if rc != nil && ac != nil {
		for i := 0; i < 20; i++ {
			role.Id = rand.Uint64()
			if role.Id > 0 {
				err := rc.Insert(role)
				if err == nil {
					a.Roles = append(a.Roles, role)
					ac.UpdateId(a.Id, a)
					return role.Id
				}
			}
		}
	}
	return 0
}

func (m *Manager) deleteRole(a *AccountInfo, id uint64) {
	rc := m.dbRole.c
	ac := m.dbAccount.c
	if rc != nil && ac != nil {
		rc.RemoveId(id)
		for i := 0; i < len(a.Roles); i++ {
			if a.Roles[i].Id == id {
				a.Roles = append(a.Roles[:i-1], a.Roles[i+1:]...)
				break
			}
		}
		ac.UpdateId(a.Id, a)
	}
}

func (m *Manager) getRole(id uint64) (role RoleInfo, err error) {
	rc := m.dbRole.c
	if rc != nil {
		err = rc.FindId(id).One(&role)
	} else {
		err = ErrDBNotFound
	}
	return
}

func (m *Manager) loginRet(s *Session, err error) {
	if err != nil || s == nil {
		m.sess.Delete(s.Info.Id)
	} else {
		s.OnKick()
	}
}

func (m *Manager) Login(pass AccountInfo) (s *Session, err error) {
	pass.Id = pass.genId()
	switch pass.Type {
	case AT_VISIT:
		if pass.UserId == _EmptyString {
			s, err = m.create(pass)
		} else {
			s, err = m.get(pass)
		}
	case AT_PASSWORD:
		s, err = m.get(pass)
		if err != nil {
			return
		}
		if s.Info.Password != pass.Password {
			err = ErrAccountInvalid
			return
		}
	case AT_AUTH2:
		authURL, ok := m.param.AuthURLs[pass.Platform]
		if !ok || authURL == _EmptyString {
			err = ErrAuthInvalidURL
			return
		}
		data, _err := m.authGet(pass.Platform, map[string]string{
			"openid":       pass.UserId,
			"access_token": pass.AccessToken,
		})
		if _err != nil {
			err = _err
			return
		}
		if data["errcode"] != 0 {
			err = ErrAccountInvalid
			return
		}
	case AT_OIDC:
		authURL, ok := m.param.AuthURLs[pass.Platform]
		if !ok || authURL == _EmptyString {
			err = ErrAuthInvalidURL
			return
		}
		data, _err := m.authGet(pass.Platform, map[string]string{
			"idtoken": pass.IdToken,
		})
		if _err != nil {
			err = _err
			return
		}
		if data["userid"] != pass.UserId {
			err = ErrAccountInvalid
			return
		}
	}
	m.loginRet(s, err)
	return
}

func (m *Manager) Logout(s *Session) {
	m.sess.Delete(s.Info.Id)
}
