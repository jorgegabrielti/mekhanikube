package diagnosis

import "testing"

func TestNewKnowledgeBase(t *testing.T) {
	t.Parallel()
	kb, err := NewKnowledgeBase()
	if err != nil {
		t.Fatalf("NewKnowledgeBase() error = %v", err)
	}
	if kb == nil {
		t.Fatal("NewKnowledgeBase() returned nil")
	}
	// Should have loaded at least the known entries.
	knownKeys := []string{
		"crashloopbackoff",
		"imagepullbackoff",
		"oomkilled",
		"high-restarts",
		"pod-pending",
		"container-not-ready",
		"node-not-ready",
		"node-memory-pressure",
		"node-disk-pressure",
		"node-pid-pressure",
		"node-unschedulable",
		"deployment-unavailable",
		"deployment-mismatch",
		"deployment-stuck",
		"deployment-zero-replicas",
		"service-no-endpoints",
		"service-lb-pending",
		"frequent-events",
	}
	for _, key := range knownKeys {
		if _, ok := kb.Lookup(key); !ok {
			t.Errorf("Lookup(%q) not found, expected it to be loaded", key)
		}
	}
}

func TestKnowledgeBaseLookupMissing(t *testing.T) {
	t.Parallel()
	kb, err := NewKnowledgeBase()
	if err != nil {
		t.Fatalf("NewKnowledgeBase() error = %v", err)
	}
	_, ok := kb.Lookup("non-existent-key")
	if ok {
		t.Error("Lookup() should return false for non-existent key")
	}
}

func TestKnowledgeBaseEnrich(t *testing.T) {
	t.Parallel()
	kb, err := NewKnowledgeBase()
	if err != nil {
		t.Fatalf("NewKnowledgeBase() error = %v", err)
	}

	p := Problem{
		Resource:       "Pod",
		Namespace:      "default",
		Name:           "my-pod",
		Issue:          "CrashLoopBackOff",
		Severity:       Critical,
		RemediationKey: "crashloopbackoff",
	}

	kb.Enrich(&p)

	if len(p.Remediation) == 0 {
		t.Error("Enrich() should populate Remediation commands")
	}

	// Check placeholder replacement.
	for _, cmd := range p.Remediation {
		if containsStr(cmd, "{pod}") || containsStr(cmd, "{name}") || containsStr(cmd, "{namespace}") {
			t.Errorf("Enrich() left unreplaced placeholder in command: %q", cmd)
		}
	}
}

func TestKnowledgeBaseEnrichUnknownKey(t *testing.T) {
	t.Parallel()
	kb, err := NewKnowledgeBase()
	if err != nil {
		t.Fatalf("NewKnowledgeBase() error = %v", err)
	}

	p := Problem{
		RemediationKey: "unknown-key-123",
	}

	kb.Enrich(&p)

	if len(p.Remediation) != 0 {
		t.Error("Enrich() should not populate Remediation for unknown key")
	}
}

func TestKnowledgeBaseEnrichEmptyNamespace(t *testing.T) {
	t.Parallel()
	kb, err := NewKnowledgeBase()
	if err != nil {
		t.Fatalf("NewKnowledgeBase() error = %v", err)
	}

	// Simulate a cluster-scoped resource (empty namespace)
	p := Problem{
		Resource:       "ClusterRole",
		Namespace:      "",
		Name:           "my-clusterrole",
		Issue:          "wildcard permissions",
		Severity:       High,
		RemediationKey: "rbac_cluster_wildcard",
	}

	kb.Enrich(&p)

	if len(p.Remediation) == 0 {
		t.Error("Enrich() should populate Remediation commands")
	}

	for _, cmd := range p.Remediation {
		if containsStr(cmd, "-n ") && !containsStr(cmd, "-n kube-system") && !containsStr(cmd, "-n default") {
			// Allow hardcoded namespaces in commands, but reject empty -n flags
			if containsStr(cmd, "-n  ") || containsStr(cmd, "-n \"\"") {
				t.Errorf("Enrich() produced broken namespace flag in command: %q", cmd)
			}
		}
		if containsStr(cmd, "{namespace}") {
			t.Errorf("Enrich() left unreplaced {namespace} in command: %q", cmd)
		}
	}
}

func TestKnowledgeBaseNewEntries(t *testing.T) {
	t.Parallel()
	kb, err := NewKnowledgeBase()
	if err != nil {
		t.Fatalf("NewKnowledgeBase() error = %v", err)
	}

	newKeys := []string{
		"event-secret-sync-failed",
		"event-image-pull-secret-missing",
		"event-failed-mount",
		"event-network-not-ready",
		"event-probe-failed",
		"event-hpa-misconfigured",
		"event-scaledobject-failed",
		"event-invalid-image",
		"event-endpoint-slice-failed",
		"cluster-api-unhealthy",
		"cluster-coredns-missing",
		"cluster-coredns-unhealthy",
		"cluster-node-skew",
		"service-externalname-empty",
	}
	for _, key := range newKeys {
		if _, ok := kb.Lookup(key); !ok {
			t.Errorf("Lookup(%q) not found, expected new knowledge entry to be loaded", key)
		}
	}
}

func TestDerivePodBase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		podName  string
		expected string
	}{
		{"deployment pod", "card-push-operations-orchestrator-worker-kafka-d89bcb56c-6q97v", "card-push-operations-orchestrator-worker-kafka"},
		{"simple deployment pod", "my-app-7b8f6c9d5-x9k2m", "my-app"},
		{"statefulset pod", "my-app-0", "my-app-0"},
		{"standalone pod", "my-pod", "my-pod"},
		{"daemonset pod", "kube-proxy-abc12", "kube-proxy-abc12"},
		{"two-part name", "nginx-deployment-5d8f6b9c7-abc12", "nginx-deployment"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := derivePodBase(tt.podName)
			if got != tt.expected {
				t.Errorf("derivePodBase(%q) = %q, want %q", tt.podName, got, tt.expected)
			}
		})
	}
}

func TestEnrichPrependsFindPodForDeploymentPods(t *testing.T) {
	t.Parallel()
	kb, err := NewKnowledgeBase()
	if err != nil {
		t.Fatalf("NewKnowledgeBase() error = %v", err)
	}

	p := Problem{
		Resource:       "Pod",
		Namespace:      "ms-card-push",
		Name:           "my-app-7b8f6c9d5-x9k2m",
		Issue:          "CrashLoopBackOff",
		Severity:       Critical,
		RemediationKey: "crashloopbackoff",
	}

	kb.Enrich(&p)

	if len(p.Remediation) == 0 {
		t.Fatal("Enrich() should populate Remediation commands")
	}

	firstCmd := p.Remediation[0]
	if !containsStr(firstCmd, "grep my-app") {
		t.Errorf("First remediation command should grep for pod base name, got: %q", firstCmd)
	}
	if !containsStr(firstCmd, "-n ms-card-push") {
		t.Errorf("First remediation command should use correct namespace, got: %q", firstCmd)
	}
}

func TestEnrichNoPrependForStandalonePod(t *testing.T) {
	t.Parallel()
	kb, err := NewKnowledgeBase()
	if err != nil {
		t.Fatalf("NewKnowledgeBase() error = %v", err)
	}

	p := Problem{
		Resource:       "Pod",
		Namespace:      "default",
		Name:           "my-pod",
		Issue:          "CrashLoopBackOff",
		Severity:       Critical,
		RemediationKey: "crashloopbackoff",
	}

	kb.Enrich(&p)

	if len(p.Remediation) == 0 {
		t.Fatal("Enrich() should populate Remediation commands")
	}

	firstCmd := p.Remediation[0]
	if containsStr(firstCmd, "grep") {
		t.Errorf("Should NOT prepend grep command for standalone pod, got: %q", firstCmd)
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
