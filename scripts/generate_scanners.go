package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

type ScannerInfo struct {
	Name      string
	CamelName string
	Type      string
	APIClient string
	ListFunc  string
	Import    string
}

const scannerTemplate = `package scanner

import (
	"context"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type {{.CamelName}}Scanner struct{}

func New{{.CamelName}}Scanner() *{{.CamelName}}Scanner {
	return &{{.CamelName}}Scanner{}
}

func (s *{{.CamelName}}Scanner) Name() string {
	return "{{.Name}}"
}

func (s *{{.CamelName}}Scanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// This is a basic scaffold. It will be expanded with real checks.
	items, err := client.{{.APIClient}}(namespace).{{.ListFunc}}(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		_ = item // implement checks here
	}
	return problems, nil
}
`

const getScannerTemplate = `package scanner

import (
	"context"

	"github.com/jorgegabrielti/nautikube/internal/diagnosis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type {{.CamelName}}Scanner struct{}

func New{{.CamelName}}Scanner() *{{.CamelName}}Scanner {
	return &{{.CamelName}}Scanner{}
}

func (s *{{.CamelName}}Scanner) Name() string {
	return "{{.Name}}"
}

func (s *{{.CamelName}}Scanner) Scan(ctx context.Context, client kubernetes.Interface, namespace string) ([]diagnosis.Problem, error) {
	// Cluster-scoped resource
	items, err := client.{{.APIClient}}().{{.ListFunc}}(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []diagnosis.Problem
	for _, item := range items.Items {
		_ = item // implement checks here
	}
	return problems, nil
}
`

func main() {
	scanners := []ScannerInfo{
		{"ConfigMap", "ConfigMap", "v1.ConfigMap", "CoreV1().ConfigMaps", "List", ""},
		{"Secret", "Secret", "v1.Secret", "CoreV1().Secrets", "List", ""},
		{"PersistentVolumeClaim", "PersistentVolumeClaim", "v1.PersistentVolumeClaim", "CoreV1().PersistentVolumeClaims", "List", ""},
		{"PersistentVolume", "PersistentVolume", "v1.PersistentVolume", "CoreV1().PersistentVolumes", "List", "cluster"},
		{"ServiceAccount", "ServiceAccount", "v1.ServiceAccount", "CoreV1().ServiceAccounts", "List", ""},
		{"StatefulSet", "StatefulSet", "v1.StatefulSet", "AppsV1().StatefulSets", "List", ""},
		{"DaemonSet", "DaemonSet", "v1.DaemonSet", "AppsV1().DaemonSets", "List", ""},
		{"ReplicaSet", "ReplicaSet", "v1.ReplicaSet", "AppsV1().ReplicaSets", "List", ""},
		{"Job", "Job", "v1.Job", "BatchV1().Jobs", "List", ""},
		{"CronJob", "CronJob", "v1.CronJob", "BatchV1().CronJobs", "List", ""},
		{"Ingress", "Ingress", "v1.Ingress", "NetworkingV1().Ingresses", "List", ""},
		{"NetworkPolicy", "NetworkPolicy", "v1.NetworkPolicy", "NetworkingV1().NetworkPolicies", "List", ""},
		{"HorizontalPodAutoscaler", "HorizontalPodAutoscaler", "v1.HorizontalPodAutoscaler", "AutoscalingV1().HorizontalPodAutoscalers", "List", ""},
		{"Role", "Role", "v1.Role", "RbacV1().Roles", "List", ""},
		{"RoleBinding", "RoleBinding", "v1.RoleBinding", "RbacV1().RoleBindings", "List", ""},
		{"ClusterRole", "ClusterRole", "v1.ClusterRole", "RbacV1().ClusterRoles", "List", "cluster"},
		{"ClusterRoleBinding", "ClusterRoleBinding", "v1.ClusterRoleBinding", "RbacV1().ClusterRoleBindings", "List", "cluster"},
	}

	tmpl, _ := template.New("scanner").Parse(scannerTemplate)
	clusterTmpl, _ := template.New("cluster").Parse(getScannerTemplate)

	for _, s := range scanners {
		filename := "internal/scanner/" + strings.ToLower(s.Name) + ".go"
		if _, err := os.Stat(filename); err == nil {
			fmt.Printf("Skipping %s, already exists\n", filename)
			continue
		}

		f, _ := os.Create(filename)
		if s.Import == "cluster" {
			clusterTmpl.Execute(f, s)
		} else {
			tmpl.Execute(f, s)
		}
		f.Close()
		fmt.Printf("Created %s\n", filename)
	}
}
