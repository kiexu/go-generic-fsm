package gfsm

// GraphConfig Implement it to generate *gfsm.Graph by various config
type GraphConfig[T, S comparable, U, V any] interface {
	NewG() (*Graph[T, S, U, V], error)
}
