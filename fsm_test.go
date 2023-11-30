package fsm

import (
	"fmt"
	"gotest.tools/v3/assert"
	"reflect"
	"sort"
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

	descFac = &DefConfig[nodeState, eventVal, edgeVal, nodeVal]{
		DescList: []*DescCell[nodeState, eventVal, edgeVal, nodeVal]{
			{
				EventVal:      payEvent,
				FromState:     []nodeState{initial},
				ToState:       paid,
				EventStoreVal: commonEdgeVal,
			},
			{
				EventVal:      deliverEvent,
				FromState:     []nodeState{paid},
				ToState:       delivering,
				EventStoreVal: commonEdgeVal,
			},
			{
				EventVal:      receiveEvent,
				FromState:     []nodeState{delivering},
				ToState:       done,
				EventStoreVal: commonEdgeVal,
			},
			{
				EventVal:      cancelEvent,
				FromState:     []nodeState{paid, delivering},
				ToState:       canceled,
				EventStoreVal: commonEdgeVal,
			},
			{
				EventVal:      readyEvent,
				FromState:     []nodeState{done, canceled},
				ToState:       initial,
				EventStoreVal: commonEdgeVal,
			},
		},
		StatusValMap: map[nodeState]nodeVal{
			initial:    initial,
			paid:       paid,
			delivering: delivering,
			done:       done,
			canceled:   canceled,
		},
	}

	nonLoopFac = &DefConfig[nodeState, eventVal, edgeVal, nodeVal]{
		DescList: []*DescCell[nodeState, eventVal, edgeVal, nodeVal]{
			{
				EventVal:      payEvent,
				FromState:     []nodeState{initial, paid}, // self loop
				ToState:       paid,
				EventStoreVal: commonEdgeVal,
			},
			{
				EventVal:      deliverEvent,
				FromState:     []nodeState{paid},
				ToState:       delivering,
				EventStoreVal: commonEdgeVal,
			},
			{
				EventVal:      receiveEvent,
				FromState:     []nodeState{delivering},
				ToState:       done,
				EventStoreVal: commonEdgeVal,
			},
			{
				EventVal:      cancelEvent,
				FromState:     []nodeState{paid, delivering},
				ToState:       canceled,
				EventStoreVal: commonEdgeVal,
			},
		},
	}
)

var demoFac = &DefConfig[string, string, string, NA]{
	DescList: []*DescCell[string, string, string, NA]{
		{
			EventVal:      "payEvent",
			FromState:     []string{"initial"}, // Multiple fromState leads to one toState
			ToState:       "paid",              // toState
			EventStoreVal: "Thanks",            // SMS message
		},
		{
			EventVal:      "deliverEvent",
			FromState:     []string{"paid"},
			ToState:       "done",
			EventStoreVal: "Coming",
		},
		{
			EventVal:      "readyEvent",
			FromState:     []string{"done", "canceled"},
			ToState:       "initial",
			EventStoreVal: "ResetOK",
		},
		{
			EventVal:      "cancelEvent",
			FromState:     []string{"paid"},
			ToState:       "canceled",
			EventStoreVal: "CancelOK",
		},
	},
}

func assertSliceEquals[T any](t *testing.T, a [][]T, b [][]T) {
	assert.Equal(t, len(a), len(b))
	l := len(a)
	aStrs := make([]string, 0, l)
	bStrs := make([]string, 0, l)
	for i := 0; i < l; i += 1 {
		aStrs = append(aStrs, tSumSli(a[i]))
		bStrs = append(bStrs, tSumSli(b[i]))
	}
	sort.Strings(aStrs)
	sort.Strings(bStrs)
	assert.DeepEqual(t, aStrs, bStrs)
}

func tSumSli[T any](input []T) string {
	resp := ""
	for i := 0; i < len(input); i += 1 {
		resp += fmt.Sprintf("%v||", input[i])
	}
	return resp
}

func TestMain(m *testing.M) {
	m.Run()
}

func BenchmarkFSM_Event(b *testing.B) {
	g, _ := descFac.NewG()
	testFSM := NewFsmByG[nodeState, eventVal, edgeVal, nodeVal](g, initial)
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
	testFSM, _ := NewFsm[nodeState, eventVal, edgeVal, nodeVal](descFac, initial)
	testFSM.SetCallbacks(&Callbacks[nodeState, eventVal, edgeVal, nodeVal]{
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
	testFSM := NewFsmByG[nodeState, eventVal, edgeVal, nodeVal](g, initial)
	testFSM.SetCallbacks(&Callbacks[nodeState, eventVal, edgeVal, nodeVal]{
		afterStateChange: func(e *Event[nodeState, eventVal, edgeVal, nodeVal]) error {
			return e.EventE().storeVal(e, w, t) // call a custom function in event store value
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

	type args[T, S comparable] struct {
		initial  T
		eventVal S
	}
	tests := []struct {
		name string
		args args[nodeState, eventVal]
		want bool
	}{

		{
			name: "ready",
			args: args[nodeState, eventVal]{
				initial:  initial,
				eventVal: payEvent,
			},
			want: true,
		},
		{
			name: "force done",
			args: args[nodeState, eventVal]{
				initial:  paid,
				eventVal: receiveEvent,
			},
			want: false,
		},
		{
			name: "deliver",
			args: args[nodeState, eventVal]{
				initial:  paid,
				eventVal: deliverEvent,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFSM, _ := NewFsm[nodeState, eventVal, edgeVal, nodeVal](descFac, tt.args.initial)
			if got := testFSM.CanTrigger(tt.args.eventVal); got != tt.want {
				t.Errorf("CanTrigger()=%v||want=%v", got, tt.want)
			}
		})
	}
}

func TestFSM_CanMigrate(t *testing.T) {

	type args[T comparable] struct {
		initial  T
		toStatus T
	}
	tests := []struct {
		name string
		args args[nodeState]
		want bool
	}{

		{
			name: "1",
			args: args[nodeState]{
				initial:  initial,
				toStatus: paid,
			},
			want: true,
		},
		{
			name: "2",
			args: args[nodeState]{
				initial:  canceled,
				toStatus: paid,
			},
			want: false,
		},
		{
			name: "3",
			args: args[nodeState]{
				initial:  paid,
				toStatus: done,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFSM, _ := NewFsm[nodeState, eventVal, edgeVal, nodeVal](nonLoopFac, tt.args.initial)
			if got := testFSM.CanMigrate(tt.args.toStatus); got != tt.want {
				t.Errorf("CanMigrate()=%v||want=%v", got, tt.want)
			}
		})
	}
}
