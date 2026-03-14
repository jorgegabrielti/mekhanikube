package scanner

import (
	"context"
	"testing"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPodScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewPodScanner()
	if got := s.Name(); got != "Pod" {
		t.Errorf("Name() = %q, want %q", got, "Pod")
	}
}

func TestPodScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		pods         []corev1.Pod
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "CrashLoopBackOff detected",
			pods: []corev1.Pod{
				podWithWaiting("crash-pod", "default", "app", "CrashLoopBackOff"),
			},
			wantCount:    1,
			wantSeverity: diagnosis.Critical,
			wantRemedKey: "crashloopbackoff",
			wantIssueHas: "CrashLoopBackOff",
		},
		{
			name: "ImagePullBackOff detected",
			pods: []corev1.Pod{
				podWithWaiting("pull-pod", "default", "app", "ImagePullBackOff"),
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "imagepullbackoff",
			wantIssueHas: "ImagePullBackOff",
		},
		{
			name: "ErrImagePull detected",
			pods: []corev1.Pod{
				podWithWaiting("pull-pod", "default", "app", "ErrImagePull"),
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "imagepullbackoff",
			wantIssueHas: "ErrImagePull",
		},
		{
			name: "OOMKilled detected",
			pods: []corev1.Pod{
				podWithTerminated("oom-pod", "production", "app", "OOMKilled"),
			},
			wantCount:    1,
			wantSeverity: diagnosis.Critical,
			wantRemedKey: "oomkilled",
			wantIssueHas: "OOMKilled",
		},
		{
			name: "high restart count over 50 is critical",
			pods: []corev1.Pod{
				podWithRestarts("restart-pod", "default", "app", 51),
			},
			wantCount:    1,
			wantSeverity: diagnosis.Critical,
			wantRemedKey: "high-restarts",
			wantIssueHas: "restarted",
		},
		{
			name: "high restart count over 20 is high",
			pods: []corev1.Pod{
				podWithRestarts("restart-pod", "default", "app", 21),
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "high-restarts",
		},
		{
			name: "high restart count over 5 is medium",
			pods: []corev1.Pod{
				podWithRestarts("restart-pod", "default", "app", 6),
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
			wantRemedKey: "high-restarts",
		},
		{
			name: "restart count 5 or below not reported",
			pods: []corev1.Pod{
				podWithRestarts("restart-pod", "default", "app", 5),
			},
			wantCount: 0,
		},
		{
			name: "pending pod detected",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pending-pod", Namespace: "default"},
					Status:     corev1.PodStatus{Phase: corev1.PodPending},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "pod-pending",
			wantIssueHas: "Pending",
		},
		{
			name: "container not ready detected",
			pods: []corev1.Pod{
				podNotReady("unready-pod", "default", "app"),
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
			wantRemedKey: "container-not-ready",
			wantIssueHas: "not ready",
		},
		{
			name: "succeeded pod skipped",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "job-pod", Namespace: "default"},
					Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
				},
			},
			wantCount: 0,
		},
		{
			name: "healthy running pod produces zero problems",
			pods: []corev1.Pod{
				healthyPod("healthy-pod", "default", "app"),
			},
			wantCount: 0,
		},
		{
			name:      "no pods produces zero problems",
			pods:      nil,
			wantCount: 0,
		},
		{
			name: "init container CrashLoopBackOff detected",
			pods: []corev1.Pod{
				podWithInitContainerWaiting("init-pod", "default", "init-app", "CrashLoopBackOff"),
			},
			wantCount:    1,
			wantSeverity: diagnosis.Critical,
			wantRemedKey: "crashloopbackoff",
		},
		{
			name: "multiple containers with different issues reported separately",
			pods: []corev1.Pod{
				podWithMultipleIssues("multi-pod", "default"),
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewSimpleClientset(podsToObjects(tt.pods)...)
			s := NewPodScanner()

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

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && findStr(s, substr)
}

func findStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func podsToObjects(pods []corev1.Pod) []runtime.Object {
	objs := make([]runtime.Object, len(pods))
	for i := range pods {
		objs[i] = &pods[i]
	}
	return objs
}

func podWithWaiting(name, ns, container, reason string) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  container,
					State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: reason}},
				},
			},
		},
	}
}

func podWithTerminated(name, ns, container, reason string) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  container,
					State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: reason}},
				},
			},
		},
	}
}

func podWithRestarts(name, ns, container string, restarts int32) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         container,
					Ready:        true,
					RestartCount: restarts,
					State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
				},
			},
		},
	}
}

func podNotReady(name, ns, container string) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  container,
					Ready: false,
					State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
				},
			},
		},
	}
}

func healthyPod(name, ns, container string) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  container,
					Ready: true,
					State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
				},
			},
		},
	}
}

func podWithInitContainerWaiting(name, ns, container, reason string) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			InitContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  container,
					State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: reason}},
				},
			},
		},
	}
}

func podWithMultipleIssues(name, ns string) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "container-a",
					State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}},
				},
				{
					Name:  "container-b",
					State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff"}},
				},
			},
		},
	}
}
