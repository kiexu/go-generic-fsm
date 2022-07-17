package fsm

type (
	// EdgeCollection fast query supported
	EdgeCollection[T, S comparable, U, V any] struct {
		eList []*Edge[T, S, U, V]       // Regular edge list
		eFast map[S][]*Edge[T, S, U, V] // Redundancy edge map for O(1) query
	}

	// Edge Event value included
	Edge[T, S comparable, U, V any] struct {
		fromV    *Vertex[T, V] // From vertex
		toV      *Vertex[T, V] // To vertex
		eventVal S             // Event value. Not unique
		storeVal U             // Anything you want. e.g. Real callback function(use Callbacks to invoke)
	}
)

// EdgeCollection

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

func (c *EdgeCollection[T, S, U, V]) EList() []*Edge[T, S, U, V] {
	return c.eList
}

func (c *EdgeCollection[T, S, U, V]) EFast() map[S][]*Edge[T, S, U, V] {
	return c.eFast
}

// Edge

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
