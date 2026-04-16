package scanner

import (
	"context"
	"testing"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPodDisruptionBudgetScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewPodDisruptionBudgetScanner()
	if got := s.Name(); got != "PodDisruptionBudget" {
		t.Errorf("Name() = %q, want %q", got, "PodDisruptionBudget")
	}
}

func TestPodDisruptionBudgetScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		pdbs         []policyv1.PodDisruptionBudget
		wantCount    int
		wantSeverity diagnosis.Severity
		wantRemedKey string
		wantIssueHas string
	}{
		{
			name: "PDB with zero DisruptionsAllowed detected",
			pdbs: []policyv1.PodDisruptionBudget{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "zero-pdb", Namespace: "default"},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					},
					Status: policyv1.PodDisruptionBudgetStatus{
						DisruptionsAllowed: 0,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "pdb_no_disruptions",
			wantIssueHas: "No disruptions",
		},
		{
			name: "PDB with positive DisruptionsAllowed not reported",
			pdbs: []policyv1.PodDisruptionBudget{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-pdb", Namespace: "default"},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					},
					Status: policyv1.PodDisruptionBudgetStatus{
						DisruptionsAllowed: 2,
					},
				},
			},
			wantCount: 0,
		},
		{
			name:      "no PodDisruptionBudgets produces zero problems",
			pdbs:      nil,
			wantCount: 0,
		},
		{
			name: "multiple PDBs with zero DisruptionsAllowed all detected",
			pdbs: []policyv1.PodDisruptionBudget{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "zero-pdb-1", Namespace: "default"},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					},
					Status: policyv1.PodDisruptionBudgetStatus{
						DisruptionsAllowed: 0,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "zero-pdb-2", Namespace: "production"},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 2},
					},
					Status: policyv1.PodDisruptionBudgetStatus{
						DisruptionsAllowed: 0,
					},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.High,
		},
		{
			name: "mixed PDB DisruptionAllowed states",
			pdbs: []policyv1.PodDisruptionBudget{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "healthy-pdb", Namespace: "default"},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					},
					Status: policyv1.PodDisruptionBudgetStatus{
						DisruptionsAllowed: 3,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "zero-pdb", Namespace: "default"},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					},
					Status: policyv1.PodDisruptionBudgetStatus{
						DisruptionsAllowed: 0,
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
		},
		{
			name: "PDB with high DisruptionAllowed not reported",
			pdbs: []policyv1.PodDisruptionBudget{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "many-disruptions-pdb", Namespace: "default"},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					},
					Status: policyv1.PodDisruptionBudgetStatus{
						DisruptionsAllowed: 10,
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "PDB with one DisruptionAllowed not reported",
			pdbs: []policyv1.PodDisruptionBudget{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "one-disruption-pdb", Namespace: "default"},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					},
					Status: policyv1.PodDisruptionBudgetStatus{
						DisruptionsAllowed: 1,
					},
				},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewSimpleClientset(pdbsToObjects(tt.pdbs)...)
			s := NewPodDisruptionBudgetScanner()

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

func pdbsToObjects(pdbs []policyv1.PodDisruptionBudget) []runtime.Object {
	objs := make([]runtime.Object, len(pdbs))
	for i := range pdbs {
		objs[i] = &pdbs[i]
	}
	return objs
}
