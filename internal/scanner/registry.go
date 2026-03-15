package scanner

// DefaultRegistry returns a Registry with all built-in scanners.
func DefaultRegistry() *Registry {
	return NewRegistry(
		NewPodScanner(),
		NewDeploymentScanner(),
		NewServiceScanner(),
		NewNodeScanner(),
		NewEventScanner(),
		NewClusterScanner(),

		// Configuration & Storage
		NewConfigMapScanner(),
		NewSecretScanner(),
		NewPersistentVolumeScanner(),
		NewPersistentVolumeClaimScanner(),
		NewServiceAccountScanner(),

		// Advanced Workloads
		NewStatefulSetScanner(),
		NewDaemonSetScanner(),
		NewReplicaSetScanner(),
		NewJobScanner(),
		NewCronJobScanner(),

		// Networking & Autoscaling
		NewIngressScanner(),
		NewNetworkPolicyScanner(),
		NewHorizontalPodAutoscalerScanner(),

		// Security & RBAC
		NewRoleScanner(),
		NewClusterRoleScanner(),
		NewRoleBindingScanner(),
		NewClusterRoleBindingScanner(),

		// Specialized
		NewPodDisruptionBudgetScanner(),
		NewResourceQuotaScanner(),
	)
}
