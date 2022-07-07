package gfsm

import (
	"fmt"
)

// StatusNotExistErr status is not in the Graph
type StatusNotExistErr[T comparable] struct {
	Status T
}

func (e StatusNotExistErr[T]) Error() string {
	return fmt.Sprintf("status %v does not exist", e.Status)
}

// InvalidEventErr Event do nothing on given status
type InvalidEventErr[T, S comparable] struct {
	Event  S
	Status T
}

func (e InvalidEventErr[T, S]) Error() string {
	return fmt.Sprintf("event %v inappropriate in current state %v", e.Event, e.Status)
}
