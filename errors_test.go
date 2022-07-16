package fsm

import (
	"reflect"
	"testing"
)

func TestStateNotExistErr(t *testing.T) {

	t.Run("StateNotExistErr", func(t *testing.T) {
		g, _ := descFac.NewG()
		_, err := g.NextEdge(2333, payEvent)
		if err == nil {
			t.Errorf("TestStateNotExistErr||err=nil")
			return
		}
		if _, ok := err.(*StateNotExistErr[nodeState]); !ok {
			t.Errorf("TestStateNotExistErr||type=%v||want=%v", reflect.TypeOf(err), "StateNotExistErr")
			return
		}
	})
}

func TestInvalidEventErr(t *testing.T) {

	g, _ := descFac.NewG()
	testFSM := NewFsmByG[nodeState, eventVal, edgeVal, nodeVal](g, initial)

	t.Run("InvalidEventErr", func(t *testing.T) {
		_, err := testFSM.Trigger("not exist event")
		if err == nil {
			t.Errorf("InvalidEventErr||err=nil")
			return
		}
		if _, ok := err.(*InvalidEventErr[nodeState, eventVal]); !ok {
			t.Errorf("TestInvalidEventErr||type=%v||want=%v", reflect.TypeOf(err), "InvalidEventErr")
			return
		}
	})
}
