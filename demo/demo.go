package demo

import (
	"github.com/kiexu/go-generic-fsm"
)

var demoFac = &fsm.DefConfig[string, string, string, fsm.NA]{
	DescList: []*fsm.DescCell[string, string, string, fsm.NA]{
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

func demo() {
	demoFsm, _ := fsm.NewFsm[string, string, string, fsm.NA](demoFac, "initial")
	_, _ = demoFsm.Trigger("payEvent")
}
