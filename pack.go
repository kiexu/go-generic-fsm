package fsm

var (
	// visualizationPack
	// https://github.com/kiexu/go-generic-fsm-visual-pack
	visualizationPack Visualizer = nil
)

// SetVisualizer for pack self-init
func SetVisualizer(v Visualizer) {
	visualizationPack = v
}

func GetVisualizer() Visualizer {
	return visualizationPack
}
