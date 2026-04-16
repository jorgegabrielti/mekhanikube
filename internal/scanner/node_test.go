package scanner

import (
	"context"
	"testing"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNodeScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewNodeScanner()
	if got := s.Name(); got != "Node" {
		t.Errorf("Name() = %q, want %q", got, "Node")
	}
}

func TestNodeScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		node         corev1.Node
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name:         "node not ready (False)",
			node:         nodeWithCondition("bad-node", corev1.NodeReady, corev1.ConditionFalse),
			wantCount:    1,
			wantSeverity: diagnosis.Critical,
			wantRemedKey: "node-not-ready",
			wantIssueHas: "not Ready",
		},
		{
			name:         "node not ready (Unknown)",
			node:         nodeWithCondition("unknown-node", corev1.NodeReady, corev1.ConditionUnknown),
			wantCount:    1,
			wantSeverity: diagnosis.Critical,
			wantRemedKey: "node-not-ready",
		},
		{
			name:         "memory pressure",
			node:         nodeWithCondition("mem-node", corev1.NodeMemoryPressure, corev1.ConditionTrue),
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "node-memory-pressure",
		},
		{
			name:         "disk pressure",
			node:         nodeWithCondition("disk-node", corev1.NodeDiskPressure, corev1.ConditionTrue),
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "node-disk-pressure",
		},
		{
			name:         "PID pressure",
			node:         nodeWithCondition("pid-node", corev1.NodePIDPressure, corev1.ConditionTrue),
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "node-pid-pressure",
		},
		{
			name: "unschedulable (cordoned)",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "cordoned-node"},
				Spec:       corev1.NodeSpec{Unschedulable: true},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Low,
			wantRemedKey: "node-unschedulable",
		},
		{
			name:      "healthy node produces zero problems",
			node:      healthyNode("good-node"),
			wantCount: 0,
		},
		{
			name: "multiple pressure conditions reported separately",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "stressed-node"},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
						{Type: corev1.NodeMemoryPressure, Status: corev1.ConditionTrue},
						{Type: corev1.NodeDiskPressure, Status: corev1.ConditionTrue},
					},
				},
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewSimpleClientset(&tt.node)
			s := NewNodeScanner()

			problems, err := s.Scan(context.Background(), client, "")
			if err != nil {
				t.Fatalf("Scan() error = %v", err)
			}

			if len(problems) != tt.wantCount {
				t.Fatalf("Scan() returned %d problems, want %d", len(problems), tt.wantCount)
			}

			if tt.wantCount == 0 {
				return
			}

			p := problems[0]
			if tt.wantSeverity != "" && p.Severity != tt.wantSeverity {
				t.Errorf("Severity = %q, want %q", p.Severity, tt.wantSeverity)
			}
			if tt.wantRemedKey != "" && p.RemediationKey != tt.wantRemedKey {
				t.Errorf("RemediationKey = %q, want %q", p.RemediationKey, tt.wantRemedKey)
			}
			if tt.wantIssueHas != "" && !containsStr(p.Issue, tt.wantIssueHas) {
				t.Errorf("Issue = %q, want it to contain %q", p.Issue, tt.wantIssueHas)
			}
		})
	}
}

// --- test helpers ---

func nodeWithCondition(name string, condType corev1.NodeConditionType, status corev1.ConditionStatus) corev1.Node {
	conditions := []corev1.NodeCondition{
		{Type: condType, Status: status},
	}
	// Add a healthy Ready condition if testing a pressure condition.
	if condType != corev1.NodeReady {
		conditions = append(conditions, corev1.NodeCondition{
			Type: corev1.NodeReady, Status: corev1.ConditionTrue,
		})
	}
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status:     corev1.NodeStatus{Conditions: conditions},
	}
}

func healthyNode(name string) corev1.Node {
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
				{Type: corev1.NodeMemoryPressure, Status: corev1.ConditionFalse},
				{Type: corev1.NodeDiskPressure, Status: corev1.ConditionFalse},
				{Type: corev1.NodePIDPressure, Status: corev1.ConditionFalse},
			},
		},
	}
}
