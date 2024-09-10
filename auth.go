package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Username             string
	Expiry               time.Time
	SolvedFirstChallenge bool
}

var (
	sessions     = make(map[string]Session)
	sessionMutex sync.RWMutex
)

func createSession(username string) string {
	sessionToken := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	sessionMutex.Lock()
	sessions[sessionToken] = Session{
		Username: username,
		Expiry:   expiresAt,
	}
	sessionMutex.Unlock()

	return sessionToken
}

func getSession(sessionToken string) *Session {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()

	session, exists := sessions[sessionToken]
	if !exists {
		return nil
	}

	if session.Expiry.Before(time.Now()) {
		delete(sessions, sessionToken)
		return nil
	}

	return &session
}

func getUsernameFromSession(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return "", err
	}

	sessionToken := cookie.Value
	userSession := getSession(sessionToken)
	if userSession == nil {
		return "", fmt.Errorf("invalid session")
	}

	return userSession.Username, nil
}

func checkFirstFlag(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false
	}
	session := getSession(cookie.Value)
	return session != nil && session.SolvedFirstChallenge
}

func markSessionSolvedFirstChallenge(sessionToken string) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	if session, exists := sessions[sessionToken]; exists {
		session.SolvedFirstChallenge = true
		sessions[sessionToken] = session
	}
}

func updateSession(sessionToken, username string) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	if session, exists := sessions[sessionToken]; exists {
		session.Username = username
		session.Expiry = time.Now().Add(24 * time.Hour)
		sessions[sessionToken] = session
	}
}
