package gfsm

import (
	"github.com/kiexu/go-generic-collection/stack"
)

type (
	// Graph the graph in FSM
	// T type of state(vertex) value
	// S type of event value
	// U type of object stored in edge
	// V type of object stored in vertex
	Graph[T, S comparable, U, V any] struct {
		adj  []*EdgeCollection[T, S, U, V] // Adjacency table
		stoV map[T]*Vertex[T, V]           // State value -> Vertex
		itoV []*Vertex[T, V]               // State idx -> Vertex
	}

	// pathWrapper a recursion helper
	pathWrapper struct {
		path [][]int
	}
)

// NextEdge Query first valid edge by state and event name
func (g *Graph[T, S, U, V]) NextEdge(fromState T, eventName S) (*Edge[T, S, U, V], error) {
	edges, err := g.NextEdges(fromState, eventName)
	if err != nil {
		return nil, err
	}
	return edges[0], nil
}

// NextEdges query all edges by state and eventE name
// Almost all FSMs do not support the same event which is on one state leads to multiple states,
// So this method is only limited inside the Graph
func (g *Graph[T, S, U, V]) NextEdges(fromState T, eventName S) ([]*Edge[T, S, U, V], error) {
	fromV := g.VertexByState(fromState)
	if fromV == nil {
		return nil, &StateNotExistErr[T]{State: fromState}
	}
	eList := g.adj[fromV.idx].EdgeByEventVal(eventName)
	if len(eList) == 0 {
		return nil, &InvalidEventErr[T, S]{State: fromState, Event: eventName}
	}
	return eList, nil
}

// HasPathTo Find if one state can be migrated to another state
func (g *Graph[T, S, U, V]) HasPathTo(fromState T, toState T) bool {
	return len(g.AllPathTo(fromState, toState)) > 0
}

// AllPathTo Find all path from fromState to toState
func (g *Graph[T, S, U, V]) AllPathTo(fromState T, toState T) [][]int {
	resp, err := g.pathTo(fromState, toState, true)
	if err != nil {
		return make([][]int, 0)
	}
	return resp
}

// pathTo Find (all) path
func (g *Graph[T, S, U, V]) pathTo(fromState T, toState T, allPath bool) ([][]int, error) {
	fromV := g.VertexByState(fromState)
	if fromV == nil {
		return nil, &StateNotExistErr[T]{State: fromState}
	}
	toV := g.VertexByState(toState)
	if toV == nil {
		return nil, &StateNotExistErr[T]{State: toState}
	}

	wrapper := &pathWrapper{}
	visited := make([]bool, len(g.itoV))
	st := stack.NewStack[int]()
	if g.adj[fromV.idx] != nil {
		for _, edges := range g.adj[fromV.idx].eList {
			visited[edges.toV.idx] = true
			st.Push(edges.toV.idx)
			g.pathDfs(st, toV.idx, visited, wrapper, allPath)
			st.Pop()
			visited[edges.toV.idx] = false
		}
	}

	return wrapper.path, nil
}

// pathDfs Find (all) path
func (g *Graph[T, S, U, V]) pathDfs(st *stack.Stack[int], toIdx int, visited []bool, w *pathWrapper, allPath bool) (abort bool) {

	if st.IsEmpty() {
		return
	}

	idx, _ := st.Peek()

	if idx == toIdx {
		currPath := make([]int, 0)
		st.ForEach(func(i int) {
			currPath = append(currPath, i)
		})
		w.path = append(w.path, currPath)
		return !allPath
	}

	if g.adj[idx] == nil {
		return
	}
	for _, edges := range g.adj[idx].eList {
		if visited[edges.toV.idx] {
			continue
		}
		visited[edges.toV.idx] = true
		st.Push(edges.toV.idx)
		if g.pathDfs(st, toIdx, visited, w, allPath) {
			return true
		}
		st.Pop()
		visited[edges.toV.idx] = false
	}

	return
}

// VertexByState Get vertex by state value
func (g *Graph[T, S, U, V]) VertexByState(stateVal T) *Vertex[T, V] {
	return g.stoV[stateVal]
}

// VertexByIdx get vertex by state idx
func (g *Graph[T, S, U, V]) VertexByIdx(idx int) *Vertex[T, V] {
	return g.itoV[idx]
}

func (g *Graph[T, S, U, V]) Adj() []*EdgeCollection[T, S, U, V] {
	return g.adj
}

func (g *Graph[T, S, U, V]) SetAdj(adj []*EdgeCollection[T, S, U, V]) {
	g.adj = adj
}

func (g *Graph[T, S, U, V]) StoV() map[T]*Vertex[T, V] {
	return g.stoV
}

func (g *Graph[T, S, U, V]) SetStoV(stoV map[T]*Vertex[T, V]) {
	g.stoV = stoV
}

func (g *Graph[T, S, U, V]) ItoV() []*Vertex[T, V] {
	return g.itoV
}

func (g *Graph[T, S, U, V]) SetItoV(itoV []*Vertex[T, V]) {
	g.itoV = itoV
}
