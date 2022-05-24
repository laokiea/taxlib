package breaker

import (
	"errors"
	"time"
)

const (
	StateClosed = iota
	StateHalfOpen
	StateOpen
)

var (
	ErrOpenState       = errors.New("open state")
	ErrTooManyRequests = errors.New("too many requests")
)

const defaultTimeout = 5

type State int

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown state"
	}
}

type BreakerOpt interface {
	Apply(*Breaker)
}

type TriggerOpt struct {
	Func func(*Breaker) bool
}

func (o TriggerOpt) Apply(b *Breaker) {
	b.Trigger = o.Func
}

type Breaker struct {
	Success           int
	Faild             int
	SuccessiveSuccess int
	SuccessiveFaild   int
	requestCount      int
	Timeout           int
	state             State
	Trigger           func(*Breaker) bool
	expiry            time.Time
	maxRequest        int
}

func defaultTrigger(b *Breaker) bool {
	return b.SuccessiveFaild > 10
}

func NewBreaker(options ...BreakerOpt) *Breaker {
	breaker := &Breaker{}
	for _, opt := range options {
		opt.Apply(breaker)
	}
	if breaker.Trigger == nil {
		breaker.Trigger = defaultTrigger
	}
	if breaker.Timeout <= 0 {
		breaker.Timeout = defaultTimeout
	}
	if breaker.maxRequest <= 0 {
		breaker.maxRequest = 2
	}
	breaker.expiry = time.Now().Add(time.Duration(breaker.Timeout) * time.Second)
	return breaker
}

func (b *Breaker) State() string {
	return b.state.String()
}

func (b *Breaker) SetState(state State) {
	b.requestCount = 0
	b.Faild = 0
	b.Success = 0
	b.SuccessiveFaild = 0
	b.SuccessiveSuccess = 0
	b.state = state
}

func (b *Breaker) CheckState() {
	if !b.expiry.IsZero() && time.Now().After(b.expiry) && b.state == StateOpen {
		b.SetState(StateHalfOpen)
		b.expiry = time.Now().Add(time.Duration(b.Timeout) * time.Second)
	}
}

func (b *Breaker) Do(f func() (interface{}, error)) (interface{}, error) {
	b.CheckState()
	if b.state == StateOpen {
		return nil, ErrOpenState
	}
	if b.state == StateHalfOpen && b.requestCount > b.maxRequest {
		return nil, ErrTooManyRequests
	}
	result, err := f()
	b.requestCount++
	if err != nil {
		b.SuccessiveSuccess = 0
		b.Faild++
		b.SuccessiveFaild++
		if b.Trigger(b) {
			b.SetState(StateOpen)
		} else if b.state == StateHalfOpen {
			b.SetState(StateOpen)
		}
	} else {
		b.SuccessiveFaild = 0
		b.Success++
		b.SuccessiveSuccess++
		if b.state == StateHalfOpen && b.SuccessiveSuccess > b.maxRequest {
			b.SetState(StateClosed)
		}
	}
	return result, err
}
