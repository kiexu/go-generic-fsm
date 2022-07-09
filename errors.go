package gfsm

import (
	"fmt"
)

// DuplicateStateAndEventErr Pair of state and event is not unique
type DuplicateStateAndEventErr[T, S comparable] struct {
	State T
	Event S
}

func (e DuplicateStateAndEventErr[T, S]) Error() string {
	return fmt.Sprintf("pair of state %v and event %v is duplicated", e.State, e.Event)
}

// StateNotExistErr State is not in the Graph
type StateNotExistErr[T comparable] struct {
	State T
}

func (e StateNotExistErr[T]) Error() string {
	return fmt.Sprintf("state %v does not exist", e.State)
}

// InvalidEventErr Event do nothing on given state
type InvalidEventErr[T, S comparable] struct {
	State T
	Event S
}

func (e InvalidEventErr[T, S]) Error() string {
	return fmt.Sprintf("event %v inappropriate in current state %v", e.Event, e.State)
}
