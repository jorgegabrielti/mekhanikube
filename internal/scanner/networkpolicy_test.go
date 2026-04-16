package scanner

import (
	"context"
	"testing"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNetworkPolicyScanner_Name(t *testing.T) {
	t.Parallel()
	s := NewNetworkPolicyScanner()
	if got := s.Name(); got != "NetworkPolicy" {
		t.Errorf("Name() = %q, want %q", got, "NetworkPolicy")
	}
}

func TestNetworkPolicyScanner_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		networkpolicies []networkingv1.NetworkPolicy
		wantCount       int
		wantSeverity    diagnosis.Severity
		wantRemedKey    string
		wantIssueHas    string
	}{
		{
			name:            "no NetworkPolicies produces zero problems",
			networkpolicies: nil,
			wantCount:       0,
		},
		{
			name: "allow-all ingress rule (empty rule) detected",
			networkpolicies: []networkingv1.NetworkPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "allow-all", Namespace: "default"},
					Spec: networkingv1.NetworkPolicySpec{
						Ingress: []networkingv1.NetworkPolicyIngressRule{
							{}, // empty rule = allow all
						},
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
			wantRemedKey: "netpol_allow_all",
			wantIssueHas: "allow-all",
		},
		{
			name: "NetworkPolicy with specific From selector not reported",
			networkpolicies: []networkingv1.NetworkPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "restricted", Namespace: "default"},
					Spec: networkingv1.NetworkPolicySpec{
						Ingress: []networkingv1.NetworkPolicyIngressRule{
							{
								From: []networkingv1.NetworkPolicyPeer{
									{
										PodSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{"app": "frontend"},
										},
									},
								},
							},
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "NetworkPolicy with specific Ports only not reported",
			networkpolicies: []networkingv1.NetworkPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "port-restricted", Namespace: "default"},
					Spec: networkingv1.NetworkPolicySpec{
						Ingress: []networkingv1.NetworkPolicyIngressRule{
							{
								Ports: []networkingv1.NetworkPolicyPort{
									{}, // port restriction present
								},
							},
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "NetworkPolicy with no ingress rules (deny-all) not reported",
			networkpolicies: []networkingv1.NetworkPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "deny-all", Namespace: "default"},
					Spec: networkingv1.NetworkPolicySpec{
						Ingress: []networkingv1.NetworkPolicyIngressRule{},
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "multiple allow-all NetworkPolicies all detected",
			networkpolicies: []networkingv1.NetworkPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "allow-all-1", Namespace: "default"},
					Spec: networkingv1.NetworkPolicySpec{
						Ingress: []networkingv1.NetworkPolicyIngressRule{{}},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "allow-all-2", Namespace: "production"},
					Spec: networkingv1.NetworkPolicySpec{
						Ingress: []networkingv1.NetworkPolicyIngressRule{{}},
					},
				},
			},
			wantCount:    2,
			wantSeverity: diagnosis.High,
		},
		{
			name: "mixed policies: allow-all and restricted",
			networkpolicies: []networkingv1.NetworkPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "restricted", Namespace: "default"},
					Spec: networkingv1.NetworkPolicySpec{
						Ingress: []networkingv1.NetworkPolicyIngressRule{
							{
								From: []networkingv1.NetworkPolicyPeer{
									{
										PodSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{"role": "backend"},
										},
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "allow-all", Namespace: "default"},
					Spec: networkingv1.NetworkPolicySpec{
						Ingress: []networkingv1.NetworkPolicyIngressRule{{}},
					},
				},
			},
			wantCount:    1,
			wantSeverity: diagnosis.High,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := fake.NewSimpleClientset(networkpoliciesToObjects(tt.networkpolicies)...)
			s := NewNetworkPolicyScanner()

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

func networkpoliciesToObjects(policies []networkingv1.NetworkPolicy) []runtime.Object {
	objs := make([]runtime.Object, len(policies))
	for i := range policies {
		objs[i] = &policies[i]
	}
	return objs
}
