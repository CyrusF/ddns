package server

import (
	"sync"
	"time"
)

type IPStore struct {
	mu        sync.RWMutex
	ip        string
	updatedAt time.Time
}

func NewIPStore() *IPStore {
	return &IPStore{}
}

func (s *IPStore) Set(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ip = ip
	s.updatedAt = time.Now()
}

func (s *IPStore) Get() (ip string, updatedAt time.Time, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ip == "" {
		return "", time.Time{}, false
	}
	return s.ip, s.updatedAt, true
}
