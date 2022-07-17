package fsm

type Visualizer interface {

	// Open visualization
	// must be thread-safe
	Open(*VisualOpenPackWrapper) error

	// Close visualization and release resources
	// must be thread-safe
	Close(*VisualClosePackWrapper) error
}

// VisualOpenPackWrapper contains all input and output params
type VisualOpenPackWrapper struct {
	// input and output
	*VisualOpenWrapper

	// FSM VisualGenerator
	VisualGen VisualGenerator
}

// VisualOpenWrapper contains all input and output(E.g. io.Reader) params
type VisualOpenWrapper struct {

	// output
	Path  *string
	Token *string // Token
}

type VisualClosePackWrapper struct {
	// input and output
	*VisualCloseWrapper
}

// VisualCloseWrapper contains all input and output(E.g. io.Reader) params of closing visualization.
type VisualCloseWrapper struct {

	// input
	Token *string // Token
}
