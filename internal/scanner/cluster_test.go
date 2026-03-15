package scanner

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestClusterScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewClusterScanner()
	if got := s.Name(); got != "Cluster" {
		t.Errorf("Name() = %q, want %q", got, "Cluster")
	}
}

func TestClusterScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		objects      []runtime.Object
		wantCount    int
		wantRemedKey string
	}{
		{
			name: "healthy cluster",
			objects: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "coredns-1",
						Namespace: "kube-system",
						Labels:    map[string]string{"k8s-app": "kube-dns"},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						Conditions: []corev1.PodCondition{
							{Type: corev1.PodReady, Status: corev1.ConditionTrue},
						},
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
					Status:     corev1.NodeStatus{NodeInfo: corev1.NodeSystemInfo{KubeletVersion: "v1.28.0"}},
				},
			},
			wantCount: 2, // CoreDNS Info + Node Info (API check skipped in fake client)
		},
		{
			name: "coredns missing",
			objects: []runtime.Object{
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
					Status:     corev1.NodeStatus{NodeInfo: corev1.NodeSystemInfo{KubeletVersion: "v1.28.0"}},
				},
			},
			wantCount:    2, // CoreDNS missing (Critical) + Node Info
			wantRemedKey: "cluster-coredns-missing",
		},
		{
			name: "node version skew",
			objects: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "coredns-1",
						Namespace: "kube-system",
						Labels:    map[string]string{"k8s-app": "kube-dns"},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						Conditions: []corev1.PodCondition{
							{Type: corev1.PodReady, Status: corev1.ConditionTrue},
						},
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
					Status:     corev1.NodeStatus{NodeInfo: corev1.NodeSystemInfo{KubeletVersion: "v1.28.0"}},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
					Status:     corev1.NodeStatus{NodeInfo: corev1.NodeSystemInfo{KubeletVersion: "v1.27.0"}},
				},
			},
			wantCount:    2, // CoreDNS Info + Node skew (Medium)
			wantRemedKey: "cluster-node-skew",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset(tt.objects...)
			s := NewClusterScanner()

			problems, err := s.Scan(context.Background(), client, "")
			if err != nil {
				t.Fatalf("Scan() error = %v", err)
			}

			if tt.wantCount >= 0 && len(problems) != tt.wantCount {
				t.Errorf("Scan() returned %d problems, want %d", len(problems), tt.wantCount)
			}

			if tt.wantRemedKey != "" {
				found := false
				for _, p := range problems {
					if p.RemediationKey == tt.wantRemedKey {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("RemediationKey %q not found in problems", tt.wantRemedKey)
				}
			}
		})
	}
}
