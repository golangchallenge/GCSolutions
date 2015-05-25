package common

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// SessionManager, does lots of stuff ;-)
//Based on http://astaxie.gitbooks.io/build-web-application-with-golang/content/en/06.3.html
//TODO: Need to add a session expiry check that would invalidate older sessions.

type SessionManager struct {
	cookieName  string     //private cookiename
	lock        sync.Mutex // protects session
	maxlifetime int64
	sessions    map[string]*Session //map of session id to cookiestorage
}

//Session is the 'db' where the session info is stored.
// Think of the value as the table and sessionID as the table name.
type Session struct {
	sessionID    string                      // unique session id
	timeAccessed time.Time                   // last access time
	value        map[interface{}]interface{} // session value stored inside
}

func (manager *SessionManager) sessionInit(sessionID string) (Session, error) {
	value := make(map[interface{}]interface{}, 0)
	newSession := &Session{sessionID: sessionID, timeAccessed: time.Now(), value: value}
	if manager.sessions == nil {
		manager.sessions = make(map[string]*Session, 0)
	}
	manager.sessions[sessionID] = newSession
	return *newSession, nil
}

func NewSessionManager(cookieName string, maxlifetime int64) (*SessionManager, error) {
	return &SessionManager{cookieName: cookieName, maxlifetime: maxlifetime}, nil
}

func (manager *SessionManager) SessionStart(w http.ResponseWriter, r *http.Request) (session Session) {
	log.Println("Locking Session store now.")
	manager.lock.Lock()
	defer manager.lock.Unlock()
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		log.Println("Setting cookie")
		sessionID := manager.getSessionID()
		log.Println("Session ID is:", sessionID)
		session, _ = manager.sessionInit(sessionID)
		log.Println("Got the manager object now")
		cookie := http.Cookie{Name: manager.cookieName, Value: url.QueryEscape(sessionID), Path: "/", HttpOnly: true, MaxAge: int(manager.maxlifetime)}
		http.SetCookie(w, &cookie)
		log.Println("Done setting cookie")
	} else {
		log.Println("Trying to fetch cookie id")
		sessionID, _ := url.QueryUnescape(cookie.Value)
		session, _ = manager.SessionRead(sessionID)
	}
	return
}

func (manager *SessionManager) SessionRead(sid string) (Session, error) {
	if element, ok := manager.sessions[sid]; ok {
		return *element, nil
	}
	session, err := manager.sessionInit(sid)
	return session, err

}

func (st *Session) Set(key, value interface{}) error {
	log.Printf("Trying to set the value of the key %v as value %v", key, value)
	st.value[key] = value
	return nil
}

func (manager *SessionManager) SessionDestroy(sessionID string) error {
	if _, ok := manager.sessions[sessionID]; ok {
		delete(manager.sessions, sessionID)
		return nil
	}
	return nil
}

func (st *Session) Get(key interface{}) interface{} {
	if v, ok := st.value[key]; ok {
		return v
	}
	return nil
}

func (st *Session) Delete(key interface{}) error {
	delete(st.value, key)
	return nil
}

func (st *Session) SessionID() string {
	return st.sessionID
}

func (manager *SessionManager) getSessionID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
