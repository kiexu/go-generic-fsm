package demo

import (
	"github.com/kiexu/go-generic-fsm"
)

var demoFac = &gfsm.DefConfig[string, string, string, gfsm.NA]{
	DescList: []*gfsm.DescCell[string, string, string, gfsm.NA]{
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
	demoFsm, _ := gfsm.NewFsm[string, string, string, gfsm.NA](demoFac, "initial")
	_, _ = demoFsm.Trigger("payEvent")
}
