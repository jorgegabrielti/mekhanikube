package diagnosis

import "errors"

// Sentinel errors for known failure modes.
var (
	// ErrNoKubeconfig indicates no kubeconfig file was found.
	ErrNoKubeconfig = errors.New("nautikube: no kubeconfig found")

	// ErrClusterUnreachable indicates the Kubernetes API server is not reachable.
	ErrClusterUnreachable = errors.New("nautikube: cluster unreachable")

	// ErrForbidden indicates insufficient RBAC permissions.
	ErrForbidden = errors.New("nautikube: insufficient permissions")
)
