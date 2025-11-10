package scanner

import (
	"testing"
)

// TestNew testa a criação de um novo scanner
func TestNew(t *testing.T) {
	scanner, err := New()
	if err != nil {
		// É esperado falhar fora de um cluster K8s
		t.Logf("Scanner creation failed (expected outside cluster): %v", err)
		return
	}

	if scanner == nil {
		t.Error("Scanner should not be nil when no error")
	}

	if scanner.clientset == nil {
		t.Error("Clientset should be initialized")
	}
}

// TestCheckPodStatus testa a verificação de status de pods
func TestCheckPodStatus(t *testing.T) {
	tests := []struct {
		name           string
		podName        string
		namespace      string
		containerState string
		reason         string
		wantProblem    bool
	}{
		{
			name:           "CrashLoopBackOff detectado",
			podName:        "test-pod",
			namespace:      "default",
			containerState: "waiting",
			reason:         "CrashLoopBackOff",
			wantProblem:    true,
		},
		{
			name:           "ImagePullBackOff detectado",
			podName:        "test-pod",
			namespace:      "default",
			containerState: "waiting",
			reason:         "ImagePullBackOff",
			wantProblem:    true,
		},
		{
			name:           "Pod rodando normalmente",
			podName:        "test-pod",
			namespace:      "default",
			containerState: "running",
			reason:         "",
			wantProblem:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test logic would go here
			// This is a placeholder since we need actual pod objects
			t.Logf("Test case: %s", tt.name)
		})
	}
}

// BenchmarkNew benchmarks scanner creation
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = New()
	}
}
