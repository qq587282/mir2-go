package session

import (
	"sync"
	"time"
)

type SessionState int

const (
	StateNone        SessionState = 0
	StateConnecting  SessionState = 1
	StateLogin       SessionState = 2
	StateSelectChar  SessionState = 3
	StatePlaying     SessionState = 4
)

type LoginSession struct {
	SessionID    int32
	Account      string
	Password     string
	CharID       int32
	CharName     string
	State        SessionState
	IP           string
	ServerAddr   string
	ConnectionID int32
	LoginTime    time.Time
	LastTime     time.Time
	GateConnID   int32
	SelGateConnID int32
	RunGateConnID int32
	SelectServer int32
	lock         sync.RWMutex
}

type SessionManager struct {
	sessions  map[int32]*LoginSession
	nextID    int32
	mu        sync.RWMutex
}

var defaultSessionManager *SessionManager

func init() {
	defaultSessionManager = NewSessionManager()
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[int32]*LoginSession),
		nextID:   1,
	}
}

func (sm *SessionManager) CreateSession(gateConnID int32, ip string) *LoginSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	sess := &LoginSession{
		SessionID:    sm.nextID,
		GateConnID:   gateConnID,
		IP:           ip,
		State:        StateConnecting,
		LoginTime:    time.Now(),
		LastTime:     time.Now(),
	}
	sm.nextID++
	sm.sessions[sess.SessionID] = sess
	return sess
}

func (sm *SessionManager) GetSession(sessionID int32) *LoginSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[sessionID]
}

func (sm *SessionManager) GetSessionByAccount(account string) *LoginSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	for _, sess := range sm.sessions {
		if sess.Account == account {
			return sess
		}
	}
	return nil
}

func (sm *SessionManager) RemoveSession(sessionID int32) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}

func (sm *SessionManager) UpdateSession(sess *LoginSession) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sess.LastTime = time.Now()
}

func (sm *SessionManager) GetOnlineCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	count := 0
	for _, sess := range sm.sessions {
		if sess.State >= StateLogin {
			count++
		}
	}
	return count
}

func GetSessionManager() *SessionManager {
	return defaultSessionManager
}
