package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Therapist struct {
	ID           string    `json:"id"`
	DisplayName  string    `json:"display_name"`
	Department   string    `json:"department"`
	PasswordHash string    `json:"-"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}

type Session struct {
	Token     string    `json:"token"`
	Therapist Therapist `json:"therapist"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Directory struct {
	mu       sync.RWMutex
	users    map[string]Therapist
	sessions map[string]Session
	now      func() time.Time
	secret   string
}

func NewDirectory(now func() time.Time, secret string) *Directory {
	if now == nil {
		now = time.Now
	}
	if strings.TrimSpace(secret) == "" {
		secret = "rehab-followup-local"
	}
	return &Directory{users: make(map[string]Therapist), sessions: make(map[string]Session), now: now, secret: secret}
}

func (d *Directory) AddTherapist(id, name, department, password string) (Therapist, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	department = strings.TrimSpace(department)
	if id == "" || name == "" || department == "" || password == "" {
		return Therapist{}, errors.New("therapist credentials are required")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.users[id]; exists {
		return Therapist{}, fmt.Errorf("therapist %q already exists", id)
	}
	user := Therapist{ID: id, DisplayName: name, Department: department, PasswordHash: hashPassword(password), Active: true, CreatedAt: d.now()}
	d.users[id] = user
	return user, nil
}

func (d *Directory) Login(id, password string) (Session, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	user, ok := d.users[strings.TrimSpace(id)]
	if !ok || !user.Active || user.PasswordHash != hashPassword(password) {
		return Session{}, errors.New("invalid therapist credentials")
	}
	now := d.now()
	token := d.tokenFor(user.ID, now)
	session := Session{Token: token, Therapist: user, ExpiresAt: now.Add(8 * time.Hour)}
	d.sessions[token] = session
	return session, nil
}

func (d *Directory) Authenticate(token string) (Therapist, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	session, ok := d.sessions[strings.TrimSpace(token)]
	if !ok || !session.ExpiresAt.After(d.now()) {
		return Therapist{}, errors.New("session is invalid or expired")
	}
	return session.Therapist, nil
}

func (d *Directory) Logout(token string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.sessions[strings.TrimSpace(token)]; !ok {
		return false
	}
	delete(d.sessions, strings.TrimSpace(token))
	return true
}

func (d *Directory) ActiveSessions() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	now := d.now()
	count := 0
	for _, session := range d.sessions {
		if session.ExpiresAt.After(now) {
			count++
		}
	}
	return count
}

func (d *Directory) tokenFor(id string, now time.Time) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%s", id, now.UnixNano(), d.secret)))
	return hex.EncodeToString(hash[:])
}

func hashPassword(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}
