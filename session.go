package mcpobtrace

import (
	"sync"
)

// SessionManager manages per-session state for multi-session transports (SSE, HTTP).
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*SessionState
}

// SessionState holds state associated with a single MCP session.
type SessionState struct {
	// Config is the per-session Obtrace configuration.
	Config *ObtraceConfig
}

// NewSessionManager creates a new SessionManager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*SessionState),
	}
}

// Get retrieves the session state for the given session ID.
func (sm *SessionManager) Get(sessionID string) (*SessionState, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	s, ok := sm.sessions[sessionID]
	return s, ok
}

// Set stores the session state for the given session ID.
func (sm *SessionManager) Set(sessionID string, state *SessionState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[sessionID] = state
}

// Delete removes the session state for the given session ID.
func (sm *SessionManager) Delete(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}
