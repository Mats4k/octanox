package octanox

import (
	"time"

	"github.com/google/uuid"
)

type StateMap map[string]bool

func (s StateMap) Generate(seconds int) string {
	state := uuid.NewString()
	s[state] = true

	go func() {
		<-time.After(time.Duration(seconds) * time.Second)
		delete(s, state)
	}()

	return state
}

func (s StateMap) Validate(state string) bool {
	_, ok := s[state]
	return ok
}

func (s StateMap) ValidateOnce(state string) bool {
	if s.Validate(state) {
		delete(s, state)
		return true
	}
	return false
}
