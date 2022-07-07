package gfsm

import (
	"sync"
)

type (
	// FSM the FSM itself
	FSM[T, S comparable, U, V any] struct {
		g          *Graph[T, S, U, V]     // Graph is config of FSM. It should be immutable
		prevStatus T                      // Last status
		currStatus T                      // Now status
		currEdge   *Edge[T, S, U, V]      // For advanced usages
		callbacks  *CallBacks[T, S, U, V] // Callbacks
		noSync     bool                   // If true, this FSM will not be thread-safe
		mutex      sync.Mutex             // RW-lock
	}

	// CallBacks do something while eventE is triggering
	CallBacks[T, S comparable, U, V any] struct {
		beforeStatusChange func(*Event[T, S, U, V]) error
		afterStatusChange  func(*Event[T, S, U, V]) error
		onDefer            func(*Event[T, S, U, V], error)
	}

	// Event packaging an eventE
	Event[T, S comparable, U, V any] struct {
		fSM      *FSM[T, S, U, V]  // Pointer to fSM
		eventVal S                 // raw input event value
		args     []interface{}     // Args to pass to callbacks
		eventE   *Edge[T, S, U, V] // Event value. eg. string or integer
	}
)

// NewFsm new a tread-safe FSM
func NewFsm[T, S comparable, U, V any](g *Graph[T, S, U, V], initStatus T) *FSM[T, S, U, V] {
	return &FSM[T, S, U, V]{
		g:          g,
		currStatus: initStatus,
	}
}

// Trigger Trigger an eventE by eventE value
func (f *FSM[T, S, U, V]) Trigger(eventVal S, args ...interface{}) (e *Event[T, S, U, V], err error) {

	if !f.noSync {
		f.mutex.Lock()
		defer f.mutex.Unlock()
	}

	// Initial eventE without toV
	e = &Event[T, S, U, V]{
		fSM:      f,
		eventVal: eventVal,
		args:     args,
	}

	defer func() {
		if f.callbacks != nil && f.callbacks.onDefer != nil {
			f.callbacks.onDefer(e, err)
		}
	}()

	// Try to get next one edge
	edge, err := f.g.NextEdge(f.currStatus, eventVal)
	if err != nil {
		return e, err
	}

	// Fill Trigger
	e.eventE = edge
	f.currEdge = edge

	// Before status change
	if f.callbacks != nil && f.callbacks.beforeStatusChange != nil {
		err = f.callbacks.beforeStatusChange(e)
		if err != nil {
			return e, err
		}
	}

	// Assign old and new status
	f.prevStatus = f.currStatus
	f.currStatus = f.g.VertexByIdx(edge.toV.idx).statusVal

	// After status change
	if f.callbacks != nil && f.callbacks.afterStatusChange != nil {
		err = f.callbacks.afterStatusChange(e)
		if err != nil {
			return e, err
		}
	}

	return e, nil
}

// PeekStatuses Peek an edge by eventE value on given status
func (f *FSM[T, S, U, V]) PeekStatuses(status T, eventVal S) []T {

	// Try to get next one edge
	edges, err := f.g.NextEdges(status, eventVal)
	if err != nil {
		return make([]T, 0)
	}

	resp := make([]T, len(edges), len(edges))
	for i, e := range edges {
		resp[i] = e.ToV().StatusVal()
	}

	return resp
}

// FSM Getter And Setter

func (f *FSM[T, S, U, V]) G() *Graph[T, S, U, V] {
	return f.g
}

func (f *FSM[T, S, U, V]) SetG(g *Graph[T, S, U, V]) {
	if !f.noSync {
		f.mutex.Lock()
		defer f.mutex.Unlock()
	}
	f.g = g
}

func (f *FSM[T, S, U, V]) PrevStatus() T {
	if !f.noSync {
		f.mutex.Lock()
		defer f.mutex.Unlock()
	}
	return f.prevStatus
}

// CurrStatus Get current status
func (f *FSM[T, S, U, V]) CurrStatus() T {
	if !f.noSync {
		f.mutex.Lock()
		defer f.mutex.Unlock()
	}
	return f.currStatus
}

// ForceSetCurrStatus prevStatus will be overwritten
// it will not modify f.currEdge. not recommended
func (f *FSM[T, S, U, V]) ForceSetCurrStatus(currStatus T) {
	if !f.noSync {
		f.mutex.Lock()
		defer f.mutex.Unlock()
	}
	f.prevStatus = f.currStatus
	f.currStatus = currStatus
}

func (f *FSM[T, S, U, V]) CurrEdge() *Edge[T, S, U, V] {
	return f.currEdge
}

func (f *FSM[T, S, U, V]) Callbacks() *CallBacks[T, S, U, V] {
	return f.callbacks
}

// SetCallbacks custom callbacks
func (f *FSM[T, S, U, V]) SetCallbacks(callbacks *CallBacks[T, S, U, V]) {
	if !f.noSync {
		f.mutex.Lock()
		defer f.mutex.Unlock()
	}
	f.callbacks = callbacks
}

func (f *FSM[T, S, U, V]) NoSync() bool {
	return f.noSync
}

// SetNoSync can force close rw-lock. not recommended.
func (f *FSM[T, S, U, V]) SetNoSync(noSync bool) {
	f.noSync = noSync
}

// CallBacks Getter And Setter

func (c *CallBacks[T, S, U, V]) BeforeStatusChange() func(*Event[T, S, U, V]) error {
	return c.beforeStatusChange
}

func (c *CallBacks[T, S, U, V]) SetBeforeStatusChange(beforeStatusChange func(*Event[T, S, U, V]) error) {
	c.beforeStatusChange = beforeStatusChange
}

func (c *CallBacks[T, S, U, V]) AfterStatusChange() func(*Event[T, S, U, V]) error {
	return c.afterStatusChange
}

func (c *CallBacks[T, S, U, V]) SetAfterStatusChange(afterStatusChange func(*Event[T, S, U, V]) error) {
	c.afterStatusChange = afterStatusChange
}

func (c *CallBacks[T, S, U, V]) OnDefer() func(*Event[T, S, U, V], error) {
	return c.onDefer
}

func (c *CallBacks[T, S, U, V]) SetOnDefer(onDefer func(*Event[T, S, U, V], error)) {
	c.onDefer = onDefer
}

// Event Getter And Setter

func (e *Event[T, S, U, V]) FSM() *FSM[T, S, U, V] {
	return e.fSM
}

// EventVal Get the raw input event value
func (e *Event[T, S, U, V]) EventVal() S {
	return e.eventVal
}

func (e *Event[T, S, U, V]) EventE() *Edge[T, S, U, V] {
	return e.eventE
}

func (e *Event[T, S, U, V]) Args() []interface{} {
	return e.args
}

func (e *Event[T, S, U, V]) FromV() *Vertex[T, V] {
	if e.eventE != nil {
		return e.eventE.fromV
	}
	return nil
}

func (e *Event[T, S, U, V]) ToV() *Vertex[T, V] {
	if e.eventE != nil {
		return e.eventE.toV
	}
	return nil
}

func (e *Event[T, S, U, V]) FromStatus() (resp T) {
	fromV := e.FromV()
	if fromV != nil {
		return fromV.statusVal
	}
	return resp
}

func (e *Event[T, S, U, V]) ToStatus() (resp T) {
	toV := e.ToV()
	if toV != nil {
		return toV.statusVal
	}
	return resp
}
