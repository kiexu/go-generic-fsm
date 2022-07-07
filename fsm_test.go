package gfsm

import (
	"reflect"
	"testing"
)

const (
	// Status
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
	nodeStatus int                                                                              // node status type in testing
	eventVal   string                                                                           // eventE type in testing
	nodeVal    int                                                                              // node value type in testing
	edgeVal    func(*Event[nodeStatus, eventVal, edgeVal, nodeVal], *wrapper, *testing.T) error // edge value type in testing

	wrapper struct {
		currEvent  eventVal
		fromStatus nodeStatus
		toStatus   nodeStatus
	}
)

var (
	commonEdgeVal edgeVal = func(e *Event[nodeStatus, eventVal, edgeVal, nodeVal], w *wrapper, t *testing.T) error {
		w.currEvent = e.EventE().eventVal
		w.fromStatus = e.EventE().FromV().StatusVal()
		w.toStatus = e.EventE().ToV().StatusVal()
		t.Logf("commonEdgeVal||event=%v||from=%v||to=%v\n", w.currEvent, w.fromStatus, w.toStatus)
		return nil
	}

	descFac = &DefGFactory[nodeStatus, eventVal, edgeVal, nodeVal]{
		DescList: []*DescCell[nodeStatus, eventVal, edgeVal, nodeVal]{
			{
				EventVal:     payEvent,
				FromStatus:   []nodeStatus{initial},
				ToStatus:     paid,
				EdgeStoreVal: commonEdgeVal,
			},
			{
				EventVal:     deliverEvent,
				FromStatus:   []nodeStatus{paid},
				ToStatus:     delivering,
				EdgeStoreVal: commonEdgeVal,
			},
			{
				EventVal:     receiveEvent,
				FromStatus:   []nodeStatus{delivering},
				ToStatus:     done,
				EdgeStoreVal: commonEdgeVal,
			},
			{
				EventVal:     cancelEvent,
				FromStatus:   []nodeStatus{paid, delivering},
				ToStatus:     canceled,
				EdgeStoreVal: commonEdgeVal,
			},
			{
				EventVal:     readyEvent,
				FromStatus:   []nodeStatus{done, canceled},
				ToStatus:     initial,
				EdgeStoreVal: commonEdgeVal,
			},
		},
		VertexValMap: map[nodeStatus]nodeVal{
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
	testFSM := NewFsm[nodeStatus, eventVal, edgeVal, nodeVal](descFac.NewG(), initial)
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
	testFSM := NewFsm[nodeStatus, eventVal, edgeVal, nodeVal](descFac.NewG(), initial)
	testFSM.SetCallbacks(&CallBacks[nodeStatus, eventVal, edgeVal, nodeVal]{
		afterStatusChange: func(e *Event[nodeStatus, eventVal, edgeVal, nodeVal]) error {
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
	testFSM := NewFsm[nodeStatus, eventVal, edgeVal, nodeVal](descFac.NewG(), initial)
	testFSM.SetCallbacks(&CallBacks[nodeStatus, eventVal, edgeVal, nodeVal]{
		afterStatusChange: func(e *Event[nodeStatus, eventVal, edgeVal, nodeVal]) error {
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
				currEvent:  payEvent,
				fromStatus: initial,
				toStatus:   paid,
			},
			false,
		},
		{
			deliverEvent,
			&wrapper{
				currEvent:  deliverEvent,
				fromStatus: paid,
				toStatus:   delivering,
			},
			false,
		},
		{
			receiveEvent,
			&wrapper{
				currEvent:  receiveEvent,
				fromStatus: delivering,
				toStatus:   done,
			},
			false,
		},
		{
			readyEvent,
			&wrapper{
				currEvent:  readyEvent,
				fromStatus: done,
				toStatus:   initial,
			},
			false,
		},
		{
			payEvent,
			&wrapper{
				currEvent:  payEvent,
				fromStatus: initial,
				toStatus:   paid,
			},
			false,
		},
		{
			cancelEvent,
			&wrapper{
				currEvent:  cancelEvent,
				fromStatus: paid,
				toStatus:   canceled,
			},
			false,
		},
		{
			readyEvent,
			&wrapper{
				currEvent:  readyEvent,
				fromStatus: canceled,
				toStatus:   initial,
			},
			false,
		},
		{
			payEvent,
			&wrapper{
				currEvent:  payEvent,
				fromStatus: initial,
				toStatus:   paid,
			},
			false,
		},
		{
			deliverEvent,
			&wrapper{
				currEvent:  deliverEvent,
				fromStatus: paid,
				toStatus:   delivering,
			},
			false,
		},
		{
			cancelEvent,
			&wrapper{
				currEvent:  cancelEvent,
				fromStatus: delivering,
				toStatus:   canceled,
			},
			false,
		},
		{
			readyEvent,
			&wrapper{
				currEvent:  readyEvent,
				fromStatus: canceled,
				toStatus:   initial,
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
				currEvent:  payEvent,
				fromStatus: initial,
				toStatus:   paid,
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
				currEvent:  deliverEvent,
				fromStatus: paid,
				toStatus:   delivering,
			},
			false,
		},
		{
			receiveEvent,
			&wrapper{
				currEvent:  receiveEvent,
				fromStatus: delivering,
				toStatus:   done,
			},
			false,
		},
	}
	return tests
}

func TestFSM_PeekStatuses(t *testing.T) {
	// todo
}
