package gfsm

import (
	"reflect"
	"testing"
)

const (
	// State
	initial    = iota // Ready to take an order. Order can be initial when last order done or canceled
	paid              // Order is paid and wait to deliver
	delivering        // Courier delivering
	done              // Received the ordered goods
	canceled          // Cancel order valid on paid and delivering

	// Events
	payEvent     = "payEvent"
	deliverEvent = "deliverEvent"
	receiveEvent = "receiveEvent"
	cancelEvent  = "cancelEvent"
	readyEvent   = "readyEvent"
)

type (
	nodeState int                                                                             // node state type in testing
	eventVal  string                                                                          // eventE type in testing
	nodeVal   int                                                                             // node value type in testing
	edgeVal   func(*Event[nodeState, eventVal, edgeVal, nodeVal], *wrapper, *testing.T) error // edge value type in testing

	wrapper struct {
		currEvent eventVal
		fromState nodeState
		toState   nodeState
	}
)

var (
	commonEdgeVal edgeVal = func(e *Event[nodeState, eventVal, edgeVal, nodeVal], w *wrapper, t *testing.T) error {
		w.currEvent = e.EventE().eventVal
		w.fromState = e.EventE().FromV().StateVal()
		w.toState = e.EventE().ToV().StateVal()
		t.Logf("commonEdgeVal||event=%v||from=%v||to=%v\n", w.currEvent, w.fromState, w.toState)
		return nil
	}

	descFac = &DefGFactory[nodeState, eventVal, edgeVal, nodeVal]{
		DescList: []*DescCell[nodeState, eventVal, edgeVal, nodeVal]{
			{
				EventVal:     payEvent,
				FromState:    []nodeState{initial},
				ToState:      paid,
				EdgeStoreVal: commonEdgeVal,
			},
			{
				EventVal:     deliverEvent,
				FromState:    []nodeState{paid},
				ToState:      delivering,
				EdgeStoreVal: commonEdgeVal,
			},
			{
				EventVal:     receiveEvent,
				FromState:    []nodeState{delivering},
				ToState:      done,
				EdgeStoreVal: commonEdgeVal,
			},
			{
				EventVal:     cancelEvent,
				FromState:    []nodeState{paid, delivering},
				ToState:      canceled,
				EdgeStoreVal: commonEdgeVal,
			},
			{
				EventVal:     readyEvent,
				FromState:    []nodeState{done, canceled},
				ToState:      initial,
				EdgeStoreVal: commonEdgeVal,
			},
		},
		VertexValMap: map[nodeState]nodeVal{
			initial:    initial,
			paid:       paid,
			delivering: delivering,
			done:       done,
			canceled:   canceled,
		},
	}
)

func TestMain(m *testing.M) {
	m.Run()
}

func BenchmarkFSM_Event(b *testing.B) {
	g, _ := descFac.NewG()
	testFSM := NewFsm[nodeState, eventVal, edgeVal, nodeVal](g, initial)
	tests := generateNonErrorTests()
	tl := len(tests)
	var err error
	b.ResetTimer()
	for i := 0; i < b.N; i += 1 {
		_, err = testFSM.Trigger(tests[i%tl].eventName)
		if err != nil {
			b.Fatalf("BenchmarkFSM_Event_Fatal||err=%v", err)
		}
	}
	b.ResetTimer()
}

// TestFSM_Event_No_Error Test regular flow
func TestFSM_Event_No_Error(t *testing.T) {

	w := &wrapper{}
	g, _ := descFac.NewG()
	testFSM := NewFsm[nodeState, eventVal, edgeVal, nodeVal](g, initial)
	testFSM.SetCallbacks(&CallBacks[nodeState, eventVal, edgeVal, nodeVal]{
		afterStateChange: func(e *Event[nodeState, eventVal, edgeVal, nodeVal]) error {
			return e.EventE().storeVal(e, w, t) // a sample to run in-config callback functions
		},
	})

	tests := generateNonErrorTests()
	for _, tt := range tests {
		t.Run(string(tt.eventName), func(t *testing.T) {
			_, err := testFSM.Trigger(tt.eventName)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestFSM_Event_No_Error||error=%v||wantErr=%v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(w, tt.wantW) {
				t.Errorf("TestFSM_Event_No_Error||w=%v||wantW=%v", w, tt.wantW)
			}
		})
	}
}

func TestFSM_Event_Contain_Error(t *testing.T) {

	w := &wrapper{}
	g, _ := descFac.NewG()
	testFSM := NewFsm[nodeState, eventVal, edgeVal, nodeVal](g, initial)
	testFSM.SetCallbacks(&CallBacks[nodeState, eventVal, edgeVal, nodeVal]{
		afterStateChange: func(e *Event[nodeState, eventVal, edgeVal, nodeVal]) error {
			return e.EventE().storeVal(e, w, t)
		},
	})

	tests := genTestsContainError()
	for _, tt := range tests {
		t.Run(string(tt.eventName), func(t *testing.T) {
			_, err := testFSM.Trigger(tt.eventName)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestFSM_Event_Contain_Error||error=%v||wantErr=%v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(w, tt.wantW) {
				t.Errorf("TestFSM_Event_Contain_Error||w=%v||want=%v", w, tt.wantW)
			}
		})
	}
}

func generateNonErrorTests() []struct {
	eventName eventVal
	wantW     *wrapper
	wantErr   bool
} {
	tests := []struct {
		eventName eventVal
		wantW     *wrapper
		wantErr   bool
	}{
		{
			payEvent,
			&wrapper{
				currEvent: payEvent,
				fromState: initial,
				toState:   paid,
			},
			false,
		},
		{
			deliverEvent,
			&wrapper{
				currEvent: deliverEvent,
				fromState: paid,
				toState:   delivering,
			},
			false,
		},
		{
			receiveEvent,
			&wrapper{
				currEvent: receiveEvent,
				fromState: delivering,
				toState:   done,
			},
			false,
		},
		{
			readyEvent,
			&wrapper{
				currEvent: readyEvent,
				fromState: done,
				toState:   initial,
			},
			false,
		},
		{
			payEvent,
			&wrapper{
				currEvent: payEvent,
				fromState: initial,
				toState:   paid,
			},
			false,
		},
		{
			cancelEvent,
			&wrapper{
				currEvent: cancelEvent,
				fromState: paid,
				toState:   canceled,
			},
			false,
		},
		{
			readyEvent,
			&wrapper{
				currEvent: readyEvent,
				fromState: canceled,
				toState:   initial,
			},
			false,
		},
		{
			payEvent,
			&wrapper{
				currEvent: payEvent,
				fromState: initial,
				toState:   paid,
			},
			false,
		},
		{
			deliverEvent,
			&wrapper{
				currEvent: deliverEvent,
				fromState: paid,
				toState:   delivering,
			},
			false,
		},
		{
			cancelEvent,
			&wrapper{
				currEvent: cancelEvent,
				fromState: delivering,
				toState:   canceled,
			},
			false,
		},
		{
			readyEvent,
			&wrapper{
				currEvent: readyEvent,
				fromState: canceled,
				toState:   initial,
			},
			false,
		},
	}
	return tests
}

func genTestsContainError() []struct {
	eventName eventVal
	wantW     *wrapper
	wantErr   bool
} {
	tests := []struct {
		eventName eventVal
		wantW     *wrapper
		wantErr   bool
	}{
		{
			payEvent,
			&wrapper{
				currEvent: payEvent,
				fromState: initial,
				toState:   paid,
			},
			false,
		},
		{
			payEvent, // can not pay twice
			nil,
			true,
		},
		{
			deliverEvent,
			&wrapper{
				currEvent: deliverEvent,
				fromState: paid,
				toState:   delivering,
			},
			false,
		},
		{
			receiveEvent,
			&wrapper{
				currEvent: receiveEvent,
				fromState: delivering,
				toState:   done,
			},
			false,
		},
	}
	return tests
}

func TestFSM_CanTrigger(t *testing.T) {

	g, _ := descFac.NewG()
	testFSM := NewFsm[nodeState, eventVal, edgeVal, nodeVal](g, initial)

	type args[T, S comparable] struct {
		eventVal  S
		forceNext T
	}
	tests := []struct {
		name string
		args args[nodeState, eventVal]
		want bool
	}{

		{
			name: "ready",
			args: args[nodeState, eventVal]{
				eventVal:  payEvent,
				forceNext: paid,
			},
			want: true,
		},
		{
			name: "force done",
			args: args[nodeState, eventVal]{
				eventVal:  receiveEvent,
				forceNext: paid,
			},
			want: false,
		},
		{
			name: "deliver",
			args: args[nodeState, eventVal]{
				eventVal:  deliverEvent,
				forceNext: done,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testFSM.CanTrigger(tt.args.eventVal); got != tt.want {
				t.Errorf("CanTrigger() = %v, want %v", got, tt.want)
			}
			testFSM.ForceSetCurrState(tt.args.forceNext)
		})
	}
}
