package gfsm

// Vertex idx start with number 0
type Vertex[T comparable, V any] struct {
	idx      int // Vertex idx. Auto generated based on unique stateVal
	stateVal T   // State value. Need to be unique
	storeVal V   // Anything you want
}

func (v *Vertex[T, V]) Idx() int {
	return v.idx
}

func (v *Vertex[T, V]) SetIdx(idx int) {
	v.idx = idx
}

func (v *Vertex[T, V]) StateVal() T {
	return v.stateVal
}

func (v *Vertex[T, V]) SetStateVal(stateVal T) {
	v.stateVal = stateVal
}

func (v *Vertex[T, V]) StoreVal() V {
	return v.storeVal
}

func (v *Vertex[T, V]) SetStoreVal(storeVal V) {
	v.storeVal = storeVal
}
