package scanner

// DefaultRegistry returns a Registry with all built-in scanners.
func DefaultRegistry() *Registry {
	return NewRegistry(
		NewPodScanner(),
		NewDeploymentScanner(),
		NewServiceScanner(),
		NewNodeScanner(),
		NewEventScanner(),
	)
}
