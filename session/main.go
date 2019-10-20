package session

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	info     AccountInfo
	expireAt time.Time
	tokenLen int
	login    bool
	m        *Manager
	OnKick   func()
	using    uint64
}

func (s *Session) UseRole(role interface{}) (err error) {
	id := reflect.ValueOf(role).FieldByName("Id").Uint()
	err = s.m.getRole(id, role)
	if err != nil {
		return
	}
	s.using = id
	return
}

func (s *Session) GetRoles(roles interface{}) error {
	return s.m.getRoles(s.info.Roles, roles)
}

func (s *Session) CreateRole(role interface{}) uint64 {
	id := s.m.createRole(s.info, role)
	if id > 0 {
		s.using = id
	}
	return id
}

func (s *Session) DeleteRole(id uint64) {
	s.m.deleteRole(s.info, id)
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
	id          string   `bson:"_id"`
	Platform    string   `bson:"platform"`
	Type        AuthType `bson:"type"`
	UserID      string   `bson:"user_id"`
	Password    string   `bson:"password"`
	AccessToken string   `bson:"access_token"`
	IDToken     string   `bson:"id_token"`
	Roles       []uint64 `bson:"roles"`
}

func (a *AccountInfo) ID() string {
	b := strings.Builder{}
	b.WriteString(a.Platform)
	b.WriteString("(")
	b.WriteString(a.UserID)
	b.WriteString(")")
	a.id = b.String()
	return a.id
}

func (a *AccountInfo) RandUserID() string {
	b := strings.Builder{}
	for i := 0; i < 18; i++ {
		b.WriteByte(byte(10 + rand.Intn(88)))
	}
	a.UserID = b.String()
	return a.UserID
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

func (m *Manager) get(a AccountInfo) (s *Session, err error) {
	id := (&a).ID()
	if v, ok := m.sess.Load(id); ok {
		s = v.(*Session)
		return
	}
	c := m.dbAccount.c
	if c != nil {
		err = c.FindId(id).One(&a)
		if err != nil {
			return
		}
		s.info = a
		m.sess.Store(id, s)
	} else {
		err = ErrDBNotFound
	}
	return
}

func (m *Manager) create() (s *Session, err error) {
	a := AccountInfo{}
	c := m.dbAccount.c
	if c != nil {
		for i := 0; i < 20; i++ {
			(&a).RandUserID()
			err = c.Insert(a)
			if err == nil {
				s = &Session{
					info:  a,
					login: true,
				}
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

func (m *Manager) createRole(a AccountInfo, role interface{}) uint64 {
	rc := m.dbRole.c
	ac := m.dbAccount.c
	idf := reflect.ValueOf(role).FieldByName("Id")
	uidf := reflect.ValueOf(role).FieldByName("accountId")
	accountID := a.ID()
	uidf.SetString(accountID)
	if rc != nil && ac != nil {
		for i := 0; i < 20; i++ {
			id := rand.Uint64()
			if id > 0 {
				idf.SetUint(id)
				err := rc.Insert(role)
				if err == nil {
					a.Roles = append(a.Roles, id)
					ac.UpdateId(accountID, a)
					return id
				}
			}
		}
	}
	return 0
}

func (m *Manager) deleteRole(a AccountInfo, id uint64) {
	rc := m.dbRole.c
	ac := m.dbAccount.c
	if rc != nil && ac != nil {
		rc.RemoveId(id)
		for i := 0; i < len(a.Roles); i++ {
			if a.Roles[i] == id {
				a.Roles = append(a.Roles[:i-1], a.Roles[i+1:]...)
				ac.UpdateId(a.ID(), a)
				return
			}
		}
	}
}

func (m *Manager) getRole(id uint64, role interface{}) (err error) {
	rc := m.dbRole.c
	if rc != nil {
		err = rc.FindId(id).One(role)
	}
	return
}

func (m *Manager) getRoles(ids []uint64, roles interface{}) (err error) {
	rc := m.dbRole.c
	if rc != nil {
		err = rc.Find(bson.M{
			"_id": bson.M{
				"$in": ids,
			},
		}).All(roles)
	}
	return
}

func (m *Manager) Login(pass AccountInfo) (s *Session, err error) {
	switch pass.Type {
	case AT_VISIT:
		if pass.UserID == _EmptyString {
			s, err = m.create()
		} else {
			s, err = m.get(pass)
		}
	case AT_PASSWORD:
		s, err = m.get(pass)
		if err != nil {
			return
		}
		if s.info.Password != pass.Password {
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
			"openid":       pass.UserID,
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
			"idtoken": pass.IDToken,
		})
		if _err != nil {
			err = _err
			return
		}
		if data["userid"] != pass.UserID {
			err = ErrAccountInvalid
			return
		}
	}
	s.login = true
	return
}

func (m *Manager) Logout(id uint32) {

}
