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

func TestConfigMapScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewConfigMapScanner()
	if got := s.Name(); got != "ConfigMap" {
		t.Errorf("Name() = %q, want %q", got, "ConfigMap")
	}
}

func TestConfigMapScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		configmaps   []corev1.ConfigMap
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "ConfigMap oversized (>500KB) detected",
			configmaps: []corev1.ConfigMap{
				oversizedConfigMap("large-config", "default", 600*1024), // 600KB
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
			wantRemedKey: "configmap_large",
			wantIssueHas: "dangerously large",
		},
		{
			name: "ConfigMap at threshold (500KB) detected",
			configmaps: []corev1.ConfigMap{
				oversizedConfigMap("at-threshold", "default", 500*1024), // Exactly 500KB
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
			wantRemedKey: "configmap_large",
		},
		{
			name: "ConfigMap just under threshold not reported",
			configmaps: []corev1.ConfigMap{
				oversizedConfigMap("small-config", "default", 499*1024), // Just under 500KB
			},
			wantCount: 0,
		},
		{
			name: "Empty ConfigMap not reported",
			configmaps: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "empty-config", Namespace: "default"},
					Data:       map[string]string{},
					BinaryData: map[string][]byte{},
				},
			},
			wantCount: 0,
		},
		{
			name: "healthy small ConfigMap produces zero problems",
			configmaps: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-config", Namespace: "default"},
					Data: map[string]string{
						"config.yaml": "app: myapp\nversion: 1.0\n",
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "multiple oversized ConfigMaps all detected",
			configmaps: []corev1.ConfigMap{
				oversizedConfigMap("large1", "default", 600*1024),
				oversizedConfigMap("large2", "kube-system", 700*1024),
			},
			wantCount:    2,
			wantSeverity: diagnosis.Medium,
		},
		{
			name:       "no ConfigMaps produces zero problems",
			configmaps: nil,
			wantCount:  0,
		},
		{
			name: "mixed oversized and normal ConfigMaps",
			configmaps: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "normal1", Namespace: "default"},
					Data: map[string]string{
						"app.conf": "setting: value\n",
					},
				},
				oversizedConfigMap("huge", "production", 800*1024),
				{
					ObjectMeta: metav1.ObjectMeta{Name: "normal2", Namespace: "staging"},
					Data: map[string]string{
						"db.conf": "host: localhost\nport: 5432\n",
					},
				},
			},
			wantCount:    1, // Only the huge one
			wantSeverity: diagnosis.Medium,
		},
		{
			name: "ConfigMap with binary data exceeding threshold",
			configmaps: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "binary-config", Namespace: "default"},
					BinaryData: map[string][]byte{
						"data.bin": make([]byte, 600*1024), // 600KB binary
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
		},
		{
			name: "ConfigMap with mixed text and binary data oversized",
			configmaps: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "mixed-config", Namespace: "default"},
					Data: map[string]string{
						"config.txt": "x" + string(make([]byte, 300*1024)), // 300KB text
					},
					BinaryData: map[string][]byte{
						"data.bin": make([]byte, 300*1024), // 300KB binary = 600KB total
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.Medium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewSimpleClientset(configmapsToObjects(tt.configmaps)...)
			s := NewConfigMapScanner()

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

func oversizedConfigMap(name, ns string, sizeBytes int) corev1.ConfigMap {
	// Create a large data entry to reach the desired size
	data := make(map[string]string)
	data["large-data"] = string(make([]byte, sizeBytes))
	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Data:       data,
	}
}

func configmapsToObjects(cms []corev1.ConfigMap) []runtime.Object {
	objs := make([]runtime.Object, len(cms))
	for i := range cms {
		objs[i] = &cms[i]
	}
	return objs
}
