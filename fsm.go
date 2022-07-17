package fsm

import (
	"fmt"
	"sync"
)

type (
	// FSM the FSM itself
	FSM[T, S comparable, U, V any] struct {
		g         *Graph[T, S, U, V]     // Graph is config of FSM. It should be immutable
		prevState T                      // Last state
		currState T                      // Now state
		currEdge  *Edge[T, S, U, V]      // For advanced usages
		callbacks *Callbacks[T, S, U, V] // Callbacks
		noSync    bool                   // If true, Trigger() and some other methods will not be thread-safe
		mutex     sync.Mutex             // RW-lock
	}

	// Callbacks do something while eventE is triggering
	Callbacks[T, S comparable, U, V any] struct {
		onEntry           func(*Event[T, S, U, V]) error
		beforeStateChange func(*Event[T, S, U, V]) error
		afterStateChange  func(*Event[T, S, U, V]) error
		onDefer           func(*Event[T, S, U, V], error)
	}

	// Event packaging an eventE
	Event[T, S comparable, U, V any] struct {
		fSM      *FSM[T, S, U, V]  // Pointer to fSM
		eventVal S                 // raw input event value
		args     []interface{}     // Args to pass to callbacks
		eventE   *Edge[T, S, U, V] // An Edge for advanced access
	}

	// VisualGenerator Type of interaction with visualization power pack
	VisualGenerator func() *FSM[string, string, string, string]

	// NA placeholder of unused type
	NA struct{}
)

// NewFsm new an FSM by desc
func NewFsm[T, S comparable, U, V any](desc GraphConfig[T, S, U, V], initState T) (*FSM[T, S, U, V], error) {
	g, err := desc.NewG()
	if err != nil {
		return nil, err
	}
	return NewFsmByG(g, initState), nil
}

// NewFsmByG new an FSM by given graph
func NewFsmByG[T, S comparable, U, V any](g *Graph[T, S, U, V], initState T) *FSM[T, S, U, V] {
	return &FSM[T, S, U, V]{
		g:         g,
		currState: initState,
	}
}

// Trigger To trigger an eventE by eventE value
// Thread safe if f.noSync == false
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

	// Callback on entry
	if f.callbacks != nil && f.callbacks.onEntry != nil {
		err = f.callbacks.onEntry(e)
		if err != nil {
			return e, err
		}
	}

	defer func() {
		// Callback on defer
		if f.callbacks != nil && f.callbacks.onDefer != nil {
			f.callbacks.onDefer(e, err)
		}
	}()

	// Try to get next one edge
	edge, err := f.g.NextEdge(f.currState, eventVal)
	if err != nil {
		return e, err
	}

	// Fill Trigger
	e.eventE = edge
	f.currEdge = edge

	// Before state change
	if f.callbacks != nil && f.callbacks.beforeStateChange != nil {
		err = f.callbacks.beforeStateChange(e)
		if err != nil {
			return e, err
		}
	}

	// Assign old and new state
	f.prevState = f.currState
	f.currState = f.g.VertexByIdx(edge.toV.idx).stateVal

	// After state change
	if f.callbacks != nil && f.callbacks.afterStateChange != nil {
		err = f.callbacks.afterStateChange(e)
		if err != nil {
			return e, err
		}
	}

	return e, nil
}

// CanTrigger Whether given eventVal can trigger event
func (f *FSM[T, S, U, V]) CanTrigger(eventVal S) bool {
	_, ok := f.PeekState(f.CurrState(), eventVal)
	return ok
}

// PeekState Peek a state by prev state and event
func (f *FSM[T, S, U, V]) PeekState(state T, eventVal S) (T, bool) {
	// Try to get next one edge
	edge, err := f.g.NextEdge(state, eventVal)
	if err != nil {
		var resp T
		return resp, false
	}

	return edge.toV.stateVal, true
}

// CanMigrate judge if current state can migrate to given toState by one or more step
func (f *FSM[T, S, U, V]) CanMigrate(toState T) bool {
	return f.g.HasPathTo(f.currState, toState)
}

func (f *FSM[T, S, U, V]) PrevState() T {
	return f.prevState
}

// CurrState Get current state
func (f *FSM[T, S, U, V]) CurrState() T {
	return f.currState
}

// OpenVisualization active visualization
// Users need to read the result fields assigned into the wrapper
// according to specific type of visualization pack
// DO NOT forget to call CloseVisualization() if you want FSM to be GC,
// Because OpenVisualization() implicitly passes the pointer of FSM to visual-pack's static variable
func (f *FSM[T, S, U, V]) OpenVisualization(wrapper *VisualOpenWrapper) error {

	// check if visual pack imported and initialed
	if visualizationPack == nil {
		return &VisualPackNotInitErr{}
	}

	return visualizationPack.Open(&VisualOpenPackWrapper{
		VisualOpenWrapper: wrapper,
		VisualGen:         f.getVisualGenerator(),
	})
}

// CloseVisualization close visualization and release resources
// Usage follow specific type of visualization pack
func (f *FSM[T, S, U, V]) CloseVisualization(wrapper *VisualCloseWrapper) error {

	// check if visual pack imported and initialed
	if visualizationPack == nil {
		return &VisualPackNotInitErr{}
	}

	return visualizationPack.Close(&VisualClosePackWrapper{
		VisualCloseWrapper: wrapper,
	})
}

// getVisualGenerator transfer FSM to all-string type (without Callbacks) for Visualizer
// Considering the limited demand for consistency in visualization, no synchronize lock here.
// Since Graph is immutable, this will hardly cause Graph problems,
// But it may lead to visual confusion between the fromState and toState in rare cases.
func (f *FSM[T, S, U, V]) getVisualGenerator() VisualGenerator {
	return func() *FSM[string, string, string, string] {
		l := len(f.g.itoV)
		og := &Graph[string, string, string, string]{
			adj:  make([]*EdgeCollection[string, string, string, string], l, l), // Adjacency table
			stoV: make(map[string]*Vertex[string, string], l),                   // State value -> Vertex
			itoV: make([]*Vertex[string, string], l, l),
		}
		for i, v := range f.g.itoV {
			og.itoV[i] = &Vertex[string, string]{
				idx:      v.idx,
				stateVal: fmt.Sprintf("%v", v.stateVal),
				storeVal: fmt.Sprintf("%v", v.storeVal),
			}
		}
		for i, v := range og.itoV {
			og.stoV[v.stateVal] = v
			if og.adj[i] == nil {
				og.adj[i] = &EdgeCollection[string, string, string, string]{
					eList: make([]*Edge[string, string, string, string], 0),
					eFast: make(map[string][]*Edge[string, string, string, string], 0),
				}
			}
			for _, e := range f.g.adj[i].eList {
				og.adj[i].addE(&Edge[string, string, string, string]{
					fromV:    og.itoV[e.fromV.idx],
					toV:      og.itoV[e.toV.idx],
					eventVal: fmt.Sprintf("%v", e.eventVal),
					storeVal: fmt.Sprintf("%v", e.storeVal),
				})
			}
		}
		return &FSM[string, string, string, string]{
			g:         og,
			prevState: fmt.Sprintf("%v", f.prevState),
			currState: fmt.Sprintf("%v", f.currState),
		}
	}
}

// ForceSetCurrState prevState will be overwritten
// It will not modify f.currEdge. not recommended
// Thread safe if f.noSync == false
func (f *FSM[T, S, U, V]) ForceSetCurrState(currState T) {
	if !f.noSync {
		f.mutex.Lock()
		defer f.mutex.Unlock()
	}
	f.prevState = f.currState
	f.currState = currState
}

// FSM Getter And Setter

func (f *FSM[T, S, U, V]) G() *Graph[T, S, U, V] {
	return f.g
}

func (f *FSM[T, S, U, V]) Callbacks() *Callbacks[T, S, U, V] {
	return f.callbacks
}

func (f *FSM[T, S, U, V]) CurrEdge() *Edge[T, S, U, V] {
	return f.currEdge
}

// SetCallbacks custom callbacks
func (f *FSM[T, S, U, V]) SetCallbacks(callbacks *Callbacks[T, S, U, V]) {
	f.callbacks = callbacks
}

func (f *FSM[T, S, U, V]) NoSync() bool {
	return f.noSync
}

// SetNoSync can force close rw-lock. not recommended.
func (f *FSM[T, S, U, V]) SetNoSync(noSync bool) {
	f.noSync = noSync
}

// Callbacks Getter And Setter

func (c *Callbacks[T, S, U, V]) BeforeStateChange() func(*Event[T, S, U, V]) error {
	return c.beforeStateChange
}

func (c *Callbacks[T, S, U, V]) SetBeforeStateChange(beforeStateChange func(*Event[T, S, U, V]) error) {
	c.beforeStateChange = beforeStateChange
}

func (c *Callbacks[T, S, U, V]) AfterStateChange() func(*Event[T, S, U, V]) error {
	return c.afterStateChange
}

func (c *Callbacks[T, S, U, V]) SetAfterStateChange(afterStateChange func(*Event[T, S, U, V]) error) {
	c.afterStateChange = afterStateChange
}

func (c *Callbacks[T, S, U, V]) OnDefer() func(*Event[T, S, U, V], error) {
	return c.onDefer
}

func (c *Callbacks[T, S, U, V]) SetOnDefer(onDefer func(*Event[T, S, U, V], error)) {
	c.onDefer = onDefer
}

// Event Getter And Setter

// FSM In concurrent usage, please use FromState and ToState after Trigger to get state, not this method
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

func (e *Event[T, S, U, V]) FromState() (resp T) {
	fromV := e.FromV()
	if fromV != nil {
		return fromV.stateVal
	}
	return resp
}

func (e *Event[T, S, U, V]) ToState() (resp T) {
	toV := e.ToV()
	if toV != nil {
		return toV.stateVal
	}
	return resp
}
