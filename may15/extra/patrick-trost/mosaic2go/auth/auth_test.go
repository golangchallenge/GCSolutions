package auth

import (
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
	"time"

	"github.com/ptrost/mosaic2go/config"
	"github.com/ptrost/mosaic2go/model"
	"github.com/ptrost/mosaic2go/test"
)

func TestAuthGenerateToken(t *testing.T) {
	auth, _ := buildAuth()
	token := auth.GenerateToken()
	if len(token) != 88 {
		t.Errorf("Auth.GenerateToken() returned wrong token, expected 88 character string, got: %s", token)
	}
}

func TestAuthCreateSession(t *testing.T) {
	auth, _ := buildAuth()
	token := auth.GenerateToken()
	user := model.NewUser(token)
	err := auth.CreateSession(user)

	test.AssertNotErr("Auth.CreateSession", err, t)
}

func TestAuthUpdateSession(t *testing.T) {
	auth, _ := buildAuth()
	token := auth.GenerateToken()
	user := model.NewUser(token)
	err := auth.CreateSession(user)
	test.AssertNotErr("Auth.CreateSession", err, t)

	user.ID = "updated"
	errUpd := auth.UpdateSession(user)
	usercheck, errGet := auth.GetSession(token)

	test.AssertNotErr("Auth.UpdateSession", errUpd, t)
	test.AssertNotErr("formatURL", errGet, t)
	test.Assert("Auth.UpdateSession", user.ID, usercheck.ID, t)
}

func TestAuthGetSession(t *testing.T) {
	auth, _ := buildAuth()
	token := auth.GenerateToken()
	user := model.NewUser(token)
	err := auth.CreateSession(user)
	test.AssertNotErr("Auth.CreateSession", err, t)
	user, errGet := auth.GetSession(token)

	test.AssertNotErr("Auth.GetSession", errGet, t)
	test.AssertNotNil("Auth.GetSession", user, t)
}

func TestAuthGetSessionUnknownToken(t *testing.T) {
	auth, _ := buildAuth()
	token := auth.GenerateToken()
	user, err := auth.GetSession(token)

	test.AssertNotErr("Auth.GetSession", err, t)
	test.AssertNil("Auth.GetSession", user, t)
}

func TestAuthGetSessionExpireSession(t *testing.T) {
	auth, _ := buildAuth()
	token := auth.GenerateToken()
	user := model.NewUser(token)
	err := auth.CreateSession(user)
	test.AssertNotErr("Auth.CreateSession", err, t)

	time.Sleep(time.Second)
	user, errGet := auth.GetSession(token)

	test.AssertNotErr("Auth.GetSession", errGet, t)
	test.AssertNil("Auth.GetSession", user, t)
}

func TestAuthRandString(t *testing.T) {
	auth, _ := buildAuth()
	rand := auth.randString(10)
	if len(rand) != 10 {
		t.Errorf("Auth.randString() returned a string of length %s, expected 10", len(rand))
	}
	pattern := "[a-z0-9]+"
	match, err := regexp.MatchString(pattern, rand)

	test.AssertNotErr("Auth.randString", err, t)
	test.Assert("Auth.randString", true, match, t)
}

func TestAuthEncodeDecodeUser(t *testing.T) {
	auth, _ := buildAuth()
	token := auth.GenerateToken()
	user := model.NewUser(token)
	encoded, err := auth.encodeUser(user)
	test.AssertNotErr("Auth.encodeUser", err, t)

	_, err2 := auth.decodeUser(encoded)
	test.AssertNotErr("Auth.edecodeUser", err2, t)
}

func buildAuth() (*Auth, *config.Config) {
	cfg := config.New(filepath.Join(getRootDir(), "config.json"))
	conn, _ := NewMemcachedConn(cfg)
	auth := New(conn, 1)
	return auth, cfg
}

func getRootDir() string {
	_, currentfile, _, _ := runtime.Caller(1)
	return path.Dir(filepath.Join(currentfile, "../"))
}
