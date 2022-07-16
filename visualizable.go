package fsm

type Visualizer interface {

	// Open visualization
	// must be thread-safe
	Open(*VisualizeStartWrapper) error

	// Close visualization and release resources
	// must be thread-safe
	Close(*VisualizeStopWrapper) error
}

// VisualizeStartWrapper contains all input and output params
type VisualizeStartWrapper struct {
	// input and output
	*VisualWrapper

	// FSM VisualGenerator
	VisualGen VisualGenerator
}

// VisualWrapper contains all input and output(E.g. io.Reader) params
// fields should be all optional
type VisualWrapper struct {

	// output
	Path  *string
	Token *string // stop Token
}

type VisualizeStopWrapper struct {

	// input
	Token *string // stop Token
}
