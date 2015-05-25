package auth

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"math/rand"
	"time"

	"github.com/ptrost/mosaic2go/config"
	"github.com/ptrost/mosaic2go/model"

	"github.com/bmizerany/mc"
)

// The Auth is responsible for authentication, creating and updating users and mantaining a user session.
type Auth struct {
	memcached *mc.Conn
	ttl       int
}

// New creates new Auth using the given memcached client for session storage
// and ttl as time after a session expires.
func New(memcached *mc.Conn, ttl int) *Auth {
	// TODO: move memcached to storage provider (in that case also implement a memory provider)
	return &Auth{memcached, ttl}
}

// GenerateToken creates a new unique auth token to authenticate the user on each API request.
func (a *Auth) GenerateToken() string {
	password := string(time.Now().UnixNano()) + a.randString(42)
	hasher := sha512.New()
	hasher.Write([]byte(password))
	token := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return token
}

// CreateSession creates a new session for a given user.
func (a *Auth) CreateSession(user *model.User) error {
	if user.Token == "" {
		panic("Auth.CreateSession() missing user token")
	}
	value, err := a.encodeUser(user)
	if err != nil {
		return err
	}
	err2 := a.memcached.Set(user.Token, value, 0, 0, a.ttl)
	if err2 != nil {
		return err2
	}
	return nil
}

// GetSession checks if a user session for a given auth token exists, and returns
// the user if a session exists.
func (a *Auth) GetSession(token string) (*model.User, error) {
	var err error
	value, _, _, err := a.memcached.Get(token)
	if err != nil {
		if err.Error() == "mc: not found" {
			return nil, nil // return nil instead of on error to indicate, that the session was not found
		}
		return nil, err
	}
	user, err := a.decodeUser(value)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateSession updates the session for the given user.
func (a *Auth) UpdateSession(user *model.User) error {
	// TODO: implement Set for memcached client
	err := a.CreateSession(user)
	return err
}

// ExtendSession extends the session for the given user.
func (a *Auth) ExtendSession(user *model.User) error {
	// TODO: implement Touch for memcached client
	err := a.CreateSession(user)
	return err
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// randString generates a random string of a given length, used for generating auth tokens.
func (a *Auth) randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// encodeUser encodes a given user object to be stored as session.
func (a *Auth) encodeUser(user *model.User) (string, error) {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(user)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

// decodeUser decodes a user object from given string
func (a *Auth) decodeUser(data string) (*model.User, error) {
	var user model.User
	buffer := bytes.NewBuffer([]byte(data))
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// NewMemcachedConn creates a memcache connection for the Auth service
func NewMemcachedConn(cfg *config.Config) (*mc.Conn, error) {
	conn, err := mc.Dial("tcp", cfg.Get("memcached_server"))
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to Memcached Server: %s.", cfg.Get("memcached_server"))
	}
	if (cfg.Get("memcached_username") != "") && (cfg.Get("memcached_password") != "") {
		err = conn.Auth(cfg.Get("memcached_username"), cfg.Get("memcached_password"))
		if err != nil {
			return nil, fmt.Errorf("Unable to authenticate to Memcached Server: %s.", cfg.Get("memcached_server"))
		}
	}
	return conn, nil
}
