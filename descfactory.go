package gfsm

type (
	// DefGFactory default factory with basic config struct
	DefGFactory[T, S comparable, U, V any] struct {
		DescList     []*DescCell[T, S, U, V] // Required. Describe FSM graph
		VertexValMap map[T]V                 // Optional. Store custom value in vertex
	}

	// DescCell describe one eventE
	DescCell[T, S comparable, U, V any] struct {
		EventVal     S
		FromStatus   []T
		ToStatus     T
		EdgeStoreVal U // Every edge's EdgeStoreVal in this cell will be assigned this field
	}
)

// Ensure interface implement
var _ GraphFactory[struct{}, struct{}, struct{}, struct{}] = new(DefGFactory[struct{}, struct{}, struct{}, struct{}])

// NewG New a Graph
func (fac *DefGFactory[T, S, U, V]) NewG() *Graph[T, S, U, V] {

	g := &Graph[T, S, U, V]{
		stoV: make(map[T]*Vertex[T, V]),
	}

	// Init itoV
	statusValSet := make(map[T]struct{})
	for _, desc := range fac.DescList {
		if _, ok := statusValSet[desc.ToStatus]; !ok {
			g.itoV = append(g.itoV, fac.newV(desc.ToStatus))
			statusValSet[desc.ToStatus] = struct{}{}
		}
		for _, fs := range desc.FromStatus {
			if _, ok := statusValSet[fs]; !ok {
				g.itoV = append(g.itoV, fac.newV(fs))
				statusValSet[fs] = struct{}{}
			}
		}
	}

	// Init idx and stoV
	// Idx starts with 0
	for i, v := range g.itoV {
		v.idx = i
		g.stoV[v.statusVal] = v
	}

	// initial adj
	vl := len(g.itoV)
	g.adj = make([]*EdgeCollection[T, S, U, V], vl, vl)
	for _, d := range fac.DescList {
		toIdx := g.VertexByStatus(d.ToStatus).idx
		for _, s := range d.FromStatus {
			fromIdx := g.VertexByStatus(s).idx
			if g.adj[fromIdx] == nil {
				g.adj[fromIdx] = &EdgeCollection[T, S, U, V]{
					eList: make([]*Edge[T, S, U, V], 0),
					eFast: make(map[S][]*Edge[T, S, U, V]),
				}
			}
			e := &Edge[T, S, U, V]{
				fromV:    g.itoV[fromIdx],
				toV:      g.itoV[toIdx],
				eventVal: d.EventVal,
				storeVal: d.EdgeStoreVal,
			}
			g.adj[fromIdx].addE(e)
		}
	}

	return g
}

// newV Without idx, autofill storeVal
func (fac *DefGFactory[T, S, U, V]) newV(status T) *Vertex[T, V] {
	genV := &Vertex[T, V]{
		statusVal: status,
	}
	if storeVal, ok := fac.VertexValMap[status]; ok {
		genV.storeVal = storeVal
	}
	return genV
}
