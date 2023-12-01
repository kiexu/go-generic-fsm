package fsm

import (
	"github.com/kiexu/go-generic-collection/stack"
)

const (
	PathOptNa      = 0
	PathOptAllPath = 1 << iota
	PathOptRing
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
	resp, err := g.pathTo(fromState, toState, PathOptNa)
	if err != nil {
		return false
	}
	return len(resp) > 0
}

// AllPathTo Find all path from fromState to toState. (fromState, toState]
func (g *Graph[T, S, U, V]) AllPathTo(fromState T, toState T) [][]int {
	resp, err := g.pathTo(fromState, toState, PathOptAllPath|PathOptRing)
	if err != nil {
		return make([][]int, 0)
	}
	return resp
}

// AllPathEdgesTo Find all Edges from fromState to toState
func (g *Graph[T, S, U, V]) AllPathEdgesTo(fromState T, toState T) [][]*Edge[T, S, U, V] {
	resp := make([][]*Edge[T, S, U, V], 0)
	paths := g.AllPathTo(fromState, toState)
	if len(paths) == 0 {
		return resp
	}
	for i := 0; i < len(paths); i += 1 {
		path := paths[i]
		if len(path) < 1 {
			continue
		}
		edges := make([]*Edge[T, S, U, V], 0)
		eCollection := g.Adj()[g.VertexByState(fromState).idx]
		for j := 0; j < len(path); j += 1 {
			for k := 0; k < len(eCollection.eList); k += 1 {
				if eCollection.eList[k].toV.idx == path[j] {
					edges = append(edges, eCollection.eList[k])
					eCollection = g.Adj()[path[j]]
					break
				}
			}
		}
		resp = append(resp, edges)
	}
	return resp
}

// pathTo Find (all) path
// ring: resp with maximum of one loop
func (g *Graph[T, S, U, V]) pathTo(fromState T, toState T, optFlag int) ([][]int, error) {
	fromV := g.VertexByState(fromState)
	if fromV == nil {
		return nil, &StateNotExistErr[T]{State: fromState}
	}
	toV := g.VertexByState(toState)
	if toV == nil {
		return nil, &StateNotExistErr[T]{State: toState}
	}

	wrapper := &pathWrapper{}
	visited := make([]int8, len(g.itoV))
	visited[fromV.idx] = 1
	st := stack.NewStack[int]()
	if g.adj[fromV.idx] != nil {
		for _, edges := range g.adj[fromV.idx].eList {
			visited[edges.toV.idx] += 1
			st.Push(edges.toV.idx)
			g.pathDfs(st, toV.idx, visited, wrapper, optFlag)
			st.Pop()
			visited[edges.toV.idx] -= 1
		}
	}

	return wrapper.path, nil
}

// pathDfs Find (all) path
func (g *Graph[T, S, U, V]) pathDfs(st *stack.Stack[int], toIdx int, visited []int8, w *pathWrapper, optFlag int) (abort bool) {

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
		return !(optFlag&PathOptAllPath > 0)
	}

	if g.adj[idx] == nil {
		return
	}
	var maxVisit int8
	if optFlag&PathOptRing > 0 {
		maxVisit = 2
	} else {
		maxVisit = 1
	}
	for _, edges := range g.adj[idx].eList {
		if visited[edges.toV.idx] >= maxVisit {
			continue
		}
		visited[edges.toV.idx] += 1
		st.Push(edges.toV.idx)
		if g.pathDfs(st, toIdx, visited, w, optFlag) {
			return true
		}
		st.Pop()
		visited[edges.toV.idx] -= 1
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
