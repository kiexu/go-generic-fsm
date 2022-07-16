package fsm

// GraphConfig Implement it to generate *fsm.Graph by various config
type GraphConfig[T, S comparable, U, V any] interface {
	NewG() (*Graph[T, S, U, V], error)
}
