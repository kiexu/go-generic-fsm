package gfsm

// GraphFactory Implement it to generate *gfsm.Graph by various config
type GraphFactory[T, S comparable, U, V any] interface {
	NewG() *Graph[T, S, U, V]
}
