package gfsm

type (
	// Graph the graph in FSM
	// T type of state(vertex) value
	// S type of eventE value
	// U type of object stored in edge
	// V type of object stored in vertex
	Graph[T, S comparable, U, V any] struct {
		adj  []*EdgeCollection[T, S, U, V] // Adjacency table
		stoV map[T]*Vertex[T, V]           // Status value -> Vertex
		itoV []*Vertex[T, V]               // Status idx -> Vertex
	}

	// 	EdgeCollection fast query supported
	EdgeCollection[T, S comparable, U, V any] struct {
		eList []*Edge[T, S, U, V]       // Regular edge list
		eFast map[S][]*Edge[T, S, U, V] // Redundancy edge map for O(1) query
	}

	// Edge Event value included
	Edge[T, S comparable, U, V any] struct {
		fromV    *Vertex[T, V] // From vertex
		toV      *Vertex[T, V] // To vertex
		eventVal S             // Event value. Need not be unique
		storeVal U             // Anything you want. e.g. Real callback function(use CallBacks to invoke)
	}

	// Vertex idx start with number 1
	Vertex[T comparable, V any] struct {
		idx       int // Vertex idx. Auto generated based on unique statusVal
		statusVal T   // Status value. Need to be unique
		storeVal  V   // Anything you want
	}
)

// Graph Methods

// NextEdges query all edges by status and eventE name
func (g *Graph[T, S, U, V]) NextEdges(fromStatus T, eventName S) ([]*Edge[T, S, U, V], error) {
	fromV := g.VertexByStatus(fromStatus)
	if fromV == nil {
		return nil, &StatusNotExistErr[T]{Status: fromStatus}
	}
	eList := g.adj[fromV.idx].EdgeByEventVal(eventName)
	if len(eList) == 0 {
		return nil, &InvalidEventErr[T, S]{Event: eventName, Status: fromStatus}
	}
	return eList, nil
}

// NextEdge query first valid edge by status and eventE name
func (g *Graph[T, S, U, V]) NextEdge(fromStatus T, eventName S) (*Edge[T, S, U, V], error) {
	edges, err := g.NextEdges(fromStatus, eventName)
	if err != nil {
		return nil, err
	}
	return edges[0], nil
}

// addE add an edge to EdgeCollection
func (c *EdgeCollection[T, S, U, V]) addE(e *Edge[T, S, U, V]) {
	if e == nil {
		return
	}
	c.eList = append(c.eList, e)
	c.eFast[e.eventVal] = append(c.eFast[e.eventVal], e)
}

// EdgeByEventVal get eventE value by eventE value
func (c *EdgeCollection[T, S, U, V]) EdgeByEventVal(eventVal S) []*Edge[T, S, U, V] {
	return c.eFast[eventVal]
}

// VertexByStatus get vertex by status value
func (g *Graph[T, S, U, V]) VertexByStatus(statusVal T) *Vertex[T, V] {
	return g.stoV[statusVal]
}

// VertexByIdx get vertex by status idx
func (g *Graph[T, S, U, V]) VertexByIdx(idx int) *Vertex[T, V] {
	return g.itoV[idx]
}

// Graph Getter And Setter

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

// Edge Getter And Setter

func (e *Edge[T, S, U, V]) FromV() *Vertex[T, V] {
	return e.fromV
}

func (e *Edge[T, S, U, V]) SetFromV(fromV *Vertex[T, V]) {
	e.fromV = fromV
}

func (e *Edge[T, S, U, V]) ToV() *Vertex[T, V] {
	return e.toV
}

func (e *Edge[T, S, U, V]) SetToV(toV *Vertex[T, V]) {
	e.toV = toV
}

func (e *Edge[T, S, U, V]) EventVal() S {
	return e.eventVal
}

func (e *Edge[T, S, U, V]) SetEventVal(eventVal S) {
	e.eventVal = eventVal
}

func (e *Edge[T, S, U, V]) StoreVal() U {
	return e.storeVal
}

func (e *Edge[T, S, U, V]) SetStoreVal(storeVal U) {
	e.storeVal = storeVal
}

// Vertex Getter And Setter

func (v *Vertex[T, V]) Idx() int {
	return v.idx
}

func (v *Vertex[T, V]) SetIdx(idx int) {
	v.idx = idx
}

func (v *Vertex[T, V]) StatusVal() T {
	return v.statusVal
}

func (v *Vertex[T, V]) SetStatusVal(statusVal T) {
	v.statusVal = statusVal
}

func (v *Vertex[T, V]) StoreVal() V {
	return v.storeVal
}

func (v *Vertex[T, V]) SetStoreVal(storeVal V) {
	v.storeVal = storeVal
}
