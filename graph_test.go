package fsm

import (
	"fmt"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	"sort"
	"strings"
	"testing"
)

func wantStateTestFormatter(g *Graph[nodeState, eventVal, edgeVal, nodeVal], states [][]nodeState) (resp []string) {
	for _, state := range states {
		idxStrList := make([]string, 0)
		for _, s := range state {
			idxStrList = append(idxStrList, fmt.Sprintf("%v", g.VertexByState(s).idx))
		}
		resp = append(resp, strings.Join(idxStrList, "-"))
	}
	sort.Strings(resp)
	return resp
}

func wantIdxTestFormatter(indexes [][]int) (resp []string) {
	for _, index := range indexes {
		idxStrList := make([]string, 0)
		for _, i := range index {
			idxStrList = append(idxStrList, fmt.Sprintf("%v", i))
		}
		resp = append(resp, strings.Join(idxStrList, "-"))
	}
	sort.Strings(resp)
	return resp
}

func wantEdgeEventTestFormatter(edges [][]*Edge[nodeState, eventVal, edgeVal, nodeVal]) (resp [][]eventVal) {
	resp = make([][]eventVal, 0)
	for i := 0; i < len(edges); i += 1 {
		edge := edges[i]
		events := make([]eventVal, 0)
		for j := 0; j < len(edge); j += 1 {
			events = append(events, edge[j].eventVal)
		}
		resp = append(resp, events)
	}
	return resp
}

func TestGraph_pathTo_1(t *testing.T) {

	g, _ := descFac.NewG()

	type args[T comparable] struct {
		fromState T
		toState   T
		allPath   bool
		ring      bool
	}
	tests := []struct {
		name    string
		args    args[nodeState]
		want    [][]nodeState
		wantErr bool
	}{
		{
			name: "initial-paid",
			args: args[nodeState]{
				fromState: initial,
				toState:   paid,
				allPath:   true,
				ring:      false,
			},
			want: [][]nodeState{
				{paid},
			},
		},
		{
			name: "initial-done",
			args: args[nodeState]{
				fromState: initial,
				toState:   done,
				allPath:   true,
				ring:      false,
			},
			want: [][]nodeState{
				{paid, delivering, done},
			},
		},
		{
			name: "initial-initial-ring",
			args: args[nodeState]{
				fromState: initial,
				toState:   initial,
				allPath:   true,
				ring:      true,
			},
			want: [][]nodeState{
				{paid, delivering, done, initial},
				{paid, canceled, initial},
				{paid, delivering, canceled, initial},
			},
		},
		{
			name: "initial-initial-one",
			args: args[nodeState]{
				fromState: initial,
				toState:   initial,
				allPath:   false,
				ring:      true,
			},
			want: [][]nodeState{
				{paid, delivering, done, initial},
				{paid, canceled, initial},
				{paid, delivering, canceled, initial},
			},
		},
		{
			name: "initial-cancel",
			args: args[nodeState]{
				fromState: initial,
				toState:   canceled,
				allPath:   true,
				ring:      false,
			},
			want: [][]nodeState{
				{paid, canceled},
				{paid, delivering, canceled},
			},
		},
		{
			name: "initial-cancel-one",
			args: args[nodeState]{
				fromState: initial,
				toState:   canceled,
				allPath:   false,
				ring:      false,
			},
			// pick one
			want: [][]nodeState{
				{paid, canceled},
				{paid, delivering, canceled},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opt int
			if tt.args.allPath {
				opt |= PathOptAllPath
			}
			if tt.args.ring {
				opt |= PathOptRing
			}
			got, err := g.pathTo(tt.args.fromState, tt.args.toState, opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("pathTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			wantList := wantStateTestFormatter(g, tt.want)
			gotList := wantIdxTestFormatter(got)
			t.Logf("pathTo()||wantList=%v||gotList=%v||allPath=%v", wantList, gotList, tt.args.allPath)
			if tt.args.allPath {
				assert.DeepEqual(t, wantList, gotList)
			} else {
				if len(wantList) == 0 {
					assert.Check(t, cmp.Len(gotList, 0))
					return
				}
				assert.Check(t, cmp.Len(gotList, 1))
				assert.Check(t, cmp.Contains(wantList, gotList[0]))
			}
		})
	}
}

func TestGraph_pathTo_2(t *testing.T) {

	g, _ := nonLoopFac.NewG()

	type args[T comparable] struct {
		fromState T
		toState   T
		allPath   bool
		ring      bool
	}
	tests := []struct {
		name    string
		args    args[nodeState]
		want    [][]nodeState
		wantErr bool
	}{
		{
			name: "initial-paid",
			args: args[nodeState]{
				fromState: initial,
				toState:   paid,
				allPath:   true,
				ring:      false,
			},
			want: [][]nodeState{
				{paid},
			},
		},
		{
			name: "initial-done",
			args: args[nodeState]{
				fromState: initial,
				toState:   done,
				allPath:   true,
				ring:      false,
			},
			want: [][]nodeState{
				{paid, delivering, done},
			},
		},
		{
			name: "initial-done-ring",
			args: args[nodeState]{
				fromState: initial,
				toState:   done,
				allPath:   true,
				ring:      true,
			},
			want: [][]nodeState{
				{paid, delivering, done},
				{paid, paid, delivering, done},
			},
		},
		{
			name: "initial-initial",
			args: args[nodeState]{
				fromState: initial,
				toState:   initial,
				allPath:   true,
				ring:      true,
			},
			want: [][]nodeState{},
		},
		{
			name: "paid-paid",
			args: args[nodeState]{
				fromState: paid,
				toState:   paid,
				allPath:   true,
				ring:      false,
			},
			want: [][]nodeState{
				{paid},
			},
		},
		{
			name: "cancel-paid",
			args: args[nodeState]{
				fromState: canceled,
				toState:   paid,
				allPath:   true,
				ring:      false,
			},
			want: [][]nodeState{},
		},
		{
			name: "cancel-paid-ring",
			args: args[nodeState]{
				fromState: canceled,
				toState:   paid,
				allPath:   true,
				ring:      true,
			},
			want: [][]nodeState{},
		},
		{
			name: "initial-cancel",
			args: args[nodeState]{
				fromState: initial,
				toState:   canceled,
				allPath:   true,
				ring:      false,
			},
			want: [][]nodeState{
				{paid, canceled},
				{paid, delivering, canceled},
			},
		},
		{
			name: "initial-cancel-ring",
			args: args[nodeState]{
				fromState: initial,
				toState:   canceled,
				allPath:   true,
				ring:      true,
			},
			want: [][]nodeState{
				{paid, paid, canceled},
				{paid, canceled},
				{paid, delivering, canceled},
				{paid, paid, delivering, canceled},
			},
		},
		{
			name: "initial-cancel-one",
			args: args[nodeState]{
				fromState: initial,
				toState:   canceled,
				allPath:   false,
				ring:      false,
			},
			// pick one
			want: [][]nodeState{
				{paid, canceled},
				{paid, delivering, canceled},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opt int
			if tt.args.allPath {
				opt |= PathOptAllPath
			}
			if tt.args.ring {
				opt |= PathOptRing
			}
			got, err := g.pathTo(tt.args.fromState, tt.args.toState, opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("pathTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			wantList := wantStateTestFormatter(g, tt.want)
			gotList := wantIdxTestFormatter(got)
			t.Logf("pathTo()||wantList=%v||gotList=%v||allPath=%v", wantList, gotList, tt.args.allPath)
			if tt.args.allPath {
				assert.DeepEqual(t, wantList, gotList)
			} else {
				if len(wantList) == 0 {
					assert.Check(t, cmp.Len(gotList, 0))
					return
				}
				assert.Check(t, cmp.Len(gotList, 1))
				assert.Check(t, cmp.Contains(wantList, gotList[0]))
			}
		})
	}
}

func TestGraph_AllPathEdgesTo_0(t *testing.T) {

	g, _ := nonLoopFac.NewG()

	type args[T comparable] struct {
		fromState T
		toState   T
	}
	tests := []struct {
		name    string
		args    args[nodeState]
		want    [][]eventVal
		wantErr bool
	}{
		{
			name: "initial-paid",
			args: args[nodeState]{
				fromState: initial,
				toState:   paid,
			},
			want: [][]eventVal{
				{payEvent},
			},
		},
		{
			name: "initial-done",
			args: args[nodeState]{
				fromState: initial,
				toState:   done,
			},
			want: [][]eventVal{
				{payEvent, deliverEvent, receiveEvent},
				{payEvent, payEvent, deliverEvent, receiveEvent},
			},
		},
		{
			name: "initial-initial",
			args: args[nodeState]{
				fromState: initial,
				toState:   initial,
			},
			want: [][]eventVal{},
		},
		{
			name: "paid-paid",
			args: args[nodeState]{
				fromState: paid,
				toState:   paid,
			},
			want: [][]eventVal{
				{payEvent},
			},
		},
		{
			name: "cancel-paid",
			args: args[nodeState]{
				fromState: canceled,
				toState:   paid,
			},
			want: [][]eventVal{},
		},
		{
			name: "initial-cancel",
			args: args[nodeState]{
				fromState: initial,
				toState:   canceled,
			},
			want: [][]eventVal{
				{payEvent, cancelEvent},
				{payEvent, payEvent, cancelEvent},
				{payEvent, deliverEvent, cancelEvent},
				{payEvent, payEvent, deliverEvent, cancelEvent},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.AllPathEdgesTo(tt.args.fromState, tt.args.toState)
			gotList := wantEdgeEventTestFormatter(got)
			t.Logf("AllPathEdgesTo()||wantList=%v||gotList=%v", tt.want, gotList)
			assert.Check(t, cmp.Len(gotList, len(tt.want)))
			assertSliceEquals(t, gotList, tt.want)
		})
	}
}

func TestGraph_AllPathEdgesTo_1(t *testing.T) {

	g, _ := complexFac.NewG()

	type args[T comparable] struct {
		fromState T
		toState   T
	}
	tests := []struct {
		name    string
		args    args[nodeState]
		want    [][]eventVal
		wantErr bool
	}{
		{
			name: "initial-paid",
			args: args[nodeState]{
				fromState: initial,
				toState:   paid,
			},
			want: [][]eventVal{
				{payEvent},
			},
		},
		{
			name: "initial-done",
			args: args[nodeState]{
				fromState: initial,
				toState:   done,
			},
			want: [][]eventVal{
				{payEvent, deliverEvent, receiveEvent},
				{payEvent, payEvent, deliverEvent, receiveEvent},
			},
		},
		{
			name: "initial-initial",
			args: args[nodeState]{
				fromState: initial,
				toState:   initial,
			},
			want: [][]eventVal{},
		},
		{
			name: "paid-paid",
			args: args[nodeState]{
				fromState: paid,
				toState:   paid,
			},
			want: [][]eventVal{
				{payEvent},
			},
		},
		{
			name: "cancel-paid",
			args: args[nodeState]{
				fromState: canceled,
				toState:   paid,
			},
			want: [][]eventVal{},
		},
		{
			name: "initial-cancel",
			args: args[nodeState]{
				fromState: initial,
				toState:   canceled,
			},
			want: [][]eventVal{
				{payEvent, cancelEvent},
				{payEvent, payEvent, cancelEvent},
				{payEvent, deliverEvent, cancelEvent},
				{payEvent, payEvent, deliverEvent, cancelEvent},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.AllPathEdgesTo(tt.args.fromState, tt.args.toState)
			gotList := wantEdgeEventTestFormatter(got)
			t.Logf("AllPathEdgesTo()||wantList=%v||gotList=%v", tt.want, gotList)
			assert.Check(t, cmp.Len(gotList, len(tt.want)))
			assertSliceEquals(t, gotList, tt.want)
		})
	}
}
