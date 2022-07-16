package fsm

import (
	"github.com/kiexu/go-generic-collection"
	"github.com/kiexu/go-generic-collection/hashset"
)

type (
	// DefConfig Default factory with basic config struct
	// As a regular FSM, {stateVal, eventVal} need to be unique
	DefConfig[T, S comparable, U, V any] struct {
		DescList     []*DescCell[T, S, U, V] // Required. Describe FSM graph
		StatusValMap map[T]V                 // Optional. Store custom value in abstract status
	}

	// DescCell Describe one eventE
	DescCell[T, S comparable, U, V any] struct {
		EventVal      S
		FromState     []T
		ToState       T
		EventStoreVal U // Every edge's EventStoreVal in this cell will be assigned this field
	}

	// stateEvent Deduplication helper
	stateEvent[T, S comparable] struct {
		stateVal T
		eventVal S
	}
)

// Ensure interface implement
var _ GraphConfig[struct{}, struct{}, struct{}, struct{}] = new(DefConfig[struct{}, struct{}, struct{}, struct{}])

// NewG New a Graph
func (fac *DefConfig[T, S, U, V]) NewG() (*Graph[T, S, U, V], error) {

	g := &Graph[T, S, U, V]{
		stoV: make(map[T]*Vertex[T, V]),
	}

	// Init itoV
	var stateValSet gcollection.Set[T] = hashset.NewHashSet[T]()
	for _, desc := range fac.DescList {
		if ok := stateValSet.Add(desc.ToState); ok {
			g.itoV = append(g.itoV, fac.newV(desc.ToState))
		}
		for _, fs := range desc.FromState {
			if ok := stateValSet.Add(fs); ok {
				g.itoV = append(g.itoV, fac.newV(fs))
			}
		}
	}

	// Init idx and stoV
	// Idx starts with 0
	for i, v := range g.itoV {
		v.idx = i
		g.stoV[v.stateVal] = v
	}

	// initial adj
	vl := len(g.itoV)
	var stateEventSet gcollection.Set[stateEvent[T, S]] = hashset.NewHashSet[stateEvent[T, S]]()
	g.adj = make([]*EdgeCollection[T, S, U, V], vl, vl)
	for _, d := range fac.DescList {
		toIdx := g.VertexByState(d.ToState).idx
		for _, s := range d.FromState {
			fromIdx := g.VertexByState(s).idx
			if g.adj[fromIdx] == nil {
				g.adj[fromIdx] = &EdgeCollection[T, S, U, V]{
					eList: make([]*Edge[T, S, U, V], 0),
					eFast: make(map[S][]*Edge[T, S, U, V]),
				}
			}
			uniqSE := stateEvent[T, S]{
				stateVal: s,
				eventVal: d.EventVal,
			}
			if ok := stateEventSet.Add(uniqSE); !ok {
				return nil, &DuplicateStateAndEventErr[T, S]{State: s, Event: d.EventVal}
			}
			e := &Edge[T, S, U, V]{
				fromV:    g.itoV[fromIdx],
				toV:      g.itoV[toIdx],
				eventVal: d.EventVal,
				storeVal: d.EventStoreVal,
			}
			g.adj[fromIdx].addE(e)
		}
	}

	return g, nil
}

// newV Without idx, autofill storeVal
func (fac *DefConfig[T, S, U, V]) newV(state T) *Vertex[T, V] {
	genV := &Vertex[T, V]{
		stateVal: state,
	}
	if storeVal, ok := fac.StatusValMap[state]; ok {
		genV.storeVal = storeVal
	}
	return genV
}
