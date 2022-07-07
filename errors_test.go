package gfsm

import (
	"reflect"
	"testing"
)

func TestStatusNotExistErr(t *testing.T) {

	t.Run("StatusNotExistErr", func(t *testing.T) {
		_, err := descFac.NewG().NextEdge(2333, payEvent)
		if err == nil {
			t.Errorf("TestStatusNotExistErr||err=nil")
			return
		}
		if _, ok := err.(*StatusNotExistErr[nodeStatus]); !ok {
			t.Errorf("TestStatusNotExistErr||type=%v||want=%v", reflect.TypeOf(err), "StatusNotExistErr")
			return
		}
	})
}

func TestInvalidEventErr(t *testing.T) {

	testFSM := NewFsm[nodeStatus, eventVal, edgeVal, nodeVal](descFac.NewG(), initial)

	t.Run("InvalidEventErr", func(t *testing.T) {
		_, err := testFSM.Trigger("not exist event")
		if err == nil {
			t.Errorf("InvalidEventErr||err=nil")
			return
		}
		if _, ok := err.(*InvalidEventErr[nodeStatus, eventVal]); !ok {
			t.Errorf("TestInvalidEventErr||type=%v||want=%v", reflect.TypeOf(err), "InvalidEventErr")
			return
		}
	})
}
