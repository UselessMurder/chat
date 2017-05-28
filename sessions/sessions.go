package sessions

import (
	"errors"
	"github.com/murder/chat/hashgenerator"
	"net/http"
	"time"
)

const (
	COOKIE_NAME = "chatsessionid"
)

type Session struct {
	SessionHash string
	UserId      uint64
	RoomId      uint64
	UserName    string
	Status      uint8
	ExpireTime  time.Time
}

type SessionList struct {
	sessions   map[string]*Session
	setChan    chan *Session
	getChan    chan string
	tubeChan   chan *Session
	removeChan chan string
	expireChan chan struct{}
	doneChan   chan struct{}
}

func (sl *SessionList) OpenSessionManager() {

	sl.sessions = make(map[string]*Session)
	sl.setChan = make(chan *Session)
	sl.tubeChan = make(chan *Session)
	sl.getChan = make(chan string)
	sl.removeChan = make(chan string)
	sl.doneChan = make(chan struct{})
	sl.expireChan = make(chan struct{})

	go func() {
		for {
			select {
			case currentSession := <-sl.setChan:
				sl.sessions[currentSession.SessionHash] = currentSession
			case sessionId := <-sl.getChan:
				sl.tubeChan <- sl.sessions[sessionId]
			case sessionId := <-sl.removeChan:
				delete(sl.sessions, sessionId)
			case <-sl.expireChan:
				currentTime := time.Now()
				for _, currentSession := range sl.sessions {
					if currentTime.After(currentSession.ExpireTime) {
						delete(sl.sessions, currentSession.SessionHash)
					}
				}
			case <-sl.doneChan:
				return
			}
		}
	}()

	go func() {
		for {
			time.Sleep(5 * time.Minute)
			sl.expireChan <- struct{}{}
		}
	}()
}

func (sl *SessionList) CloseSessionManager() {
	sl.doneChan <- struct{}{}
}

func (sl *SessionList) SetSession(currentSession *Session) {
	sl.setChan <- currentSession
}

func (sl *SessionList) GetSession(sessionsHash string) (error, *Session) {
	sl.getChan <- sessionsHash
	currentSession := <-sl.tubeChan
	var err error
	if currentSession == nil {
		err = errors.New("Invalid session!")
	} else {
		currentSession.ExpireTime = time.Now().Add(24 * time.Hour)
	}

	return err, currentSession
}

func (sl *SessionList) GetCookie(r *http.Request, w http.ResponseWriter) string {

	cookie, err := r.Cookie(COOKIE_NAME)

	if err != nil {

		hash, _ := hashgenerator.GenerateHash28(time.Now().String(), "User")
		t := time.Now().Add(24 * time.Hour)

		sl.setChan <- &Session{hash, 0, 0, "Guest", 0, t}

		cookie = &http.Cookie{
			Name:    COOKIE_NAME,
			Value:   hash,
			Expires: t,
		}

	} else {
		cookie.Expires = time.Now().Add(24 * time.Hour)
	}

	http.SetCookie(w, cookie)

	return cookie.Value
}
