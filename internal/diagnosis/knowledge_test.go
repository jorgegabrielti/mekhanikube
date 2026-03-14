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

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
