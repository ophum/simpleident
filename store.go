package main

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Store[T any] struct {
	records         map[uuid.UUID]*T
	mu              sync.RWMutex
	persistFilename string
	deepCopyFn      func(a *T) *T
}

func NewStore[T any](persistFilename string, deepCopyFn func(a *T) *T) *Store[T] {
	s := &Store[T]{
		records:         map[uuid.UUID]*T{},
		mu:              sync.RWMutex{},
		persistFilename: persistFilename,
		deepCopyFn:      deepCopyFn,
	}
	s.load()
	return s
}

func (s *Store[T]) save() error {
	f, err := os.Create(s.persistFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(s.records); err != nil {
		return err
	}
	return nil
}

func (s *Store[T]) load() error {
	f, err := os.Open(s.persistFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&s.records); err != nil {
		return err
	}
	return nil
}

func (s *Store[T]) List() ([]*T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ret := make([]*T, 0, len(s.records))
	for _, a := range s.records {
		ret = append(ret, s.deepCopyFn(a))
	}
	return ret, nil
}

func (s *Store[T]) Get(id uuid.UUID) (*T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	d, ok := s.records[id]
	if !ok {
		return nil, errors.New("not found")
	}

	return s.deepCopyFn(d), nil
}

func (s *Store[T]) Find(conds func(a *T) bool) (*T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.records {
		if conds(r) {
			return s.deepCopyFn(r), nil
		}
	}
	return nil, errors.New("not found")
}

func (s *Store[T]) Add(id uuid.UUID, data *T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.records[id]
	if exists {
		return errors.New("already exists")
	}

	s.records[id] = data

	return s.save()
}

func (s *Store[T]) Update(id uuid.UUID, data *T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.records[id]
	if !exists {
		return errors.New("not found")
	}

	s.records[id] = data

	return s.save()
}
