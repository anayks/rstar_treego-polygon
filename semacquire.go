package main

import (
	"errors"
	"time"
)

var (
	ErrNoTickets     = errors.New("deadlock: semaphor was locked so long")
	ErrIllegalUnlock = errors.New("can't unlock not locked structure")
)

type SemInterface interface {
	Acquire() func()
	Release() func()
}

type Semacquire struct {
	sem     chan struct{}
	timeout time.Duration
}

func (s *Semacquire) Acquire() error {
	select {
	case s.sem <- struct{}{}:
		{
			return nil
		}
	}
}

func (s *Semacquire) Release() error {
	select {
	case <-s.sem:
		{
			return nil
		}
	}
}

func SemNew(timeout time.Duration) *Semacquire {
	return &Semacquire{
		sem:     make(chan struct{}, 1),
		timeout: timeout,
	}
}
