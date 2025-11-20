package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jorgegabrielti/nautikube/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Scanner analisa o cluster Kubernetes
type Scanner struct {
	Client *kubernetes.Clientset
}

// New cria uma nova instância do Scanner com detecção agnóstica
func New() (*Scanner, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente Kubernetes: %v", err)
	}

	return &Scanner{
		Client: clientset,
	}, nil
}

// getKubeConfig tenta obter a configuração do cluster de várias formas (Agnostic)
func getKubeConfig() (*rest.Config, error) {
	// 1. Tenta configuração in-cluster (se estiver rodando dentro de um Pod)
	config, err := rest.InClusterConfig()
	if err == nil {
		fmt.Println("🔌 Usando configuração In-Cluster")
		return config, nil
	}

	// 2. Tenta usar o arquivo modificado pelo entrypoint (config_mod)
	modPath := "/root/.kube/config_mod"
	if _, err := os.Stat(modPath); err == nil {
		config, err := clientcmd.BuildConfigFromFlags("", modPath)
		if err == nil {
			fmt.Println("🔌 Usando configuração modificada (config_mod)")
			return config, nil
		}
	}

	// 3. Tenta usar o arquivo padrão ~/.kube/config
	if home := homedir.HomeDir(); home != "" {
		kubeconfig := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(kubeconfig); err == nil {
			config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err == nil {
				fmt.Println("🔌 Usando configuração padrão (~/.kube/config)")
				return config, nil
			}
		}
	}

	// 4. Tenta ler da variável de ambiente KUBECONFIG
	if envConfig := os.Getenv("KUBECONFIG"); envConfig != "" {
		config, err := clientcmd.BuildConfigFromFlags("", envConfig)
		if err == nil {
			fmt.Println("🔌 Usando configuração da variável KUBECONFIG")
			return config, nil
		}
	}

	return nil, fmt.Errorf("nenhuma configuração Kubernetes encontrada (tentado: in-cluster, config_mod, home, env)")
}

// ScanPods analisa pods em busca de problemas
func (s *Scanner) ScanPods(ctx context.Context, namespace string) ([]types.Problem, error) {
	pods, err := s.Client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []types.Problem
	for _, pod := range pods.Items {
		// Verifica pods que não estão rodando
		if pod.Status.Phase != "Running" && pod.Status.Phase != "Succeeded" {
			problems = append(problems, types.Problem{
				Kind:      "Pod",
				Namespace: pod.Namespace,
				Name:      pod.Name,
				Error:     fmt.Sprintf("Pod is in %s state", pod.Status.Phase),
			})
		}

		// Verifica restarts
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.RestartCount > 5 {
				problems = append(problems, types.Problem{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      pod.Name,
					Error:     fmt.Sprintf("Container %s has high restart count: %d", containerStatus.Name, containerStatus.RestartCount),
				})
			}
		}
	}
	return problems, nil
}

// ScanConfigMaps analisa ConfigMaps não utilizados (simplificado)
func (s *Scanner) ScanConfigMaps(ctx context.Context, namespace string) ([]types.Problem, error) {
	// Implementação simplificada: lista ConfigMaps e verifica se existem
	// Em uma implementação real, verificaríamos referências em Pods
	cms, err := s.Client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var problems []types.Problem

	// Exemplo: Detecta ConfigMaps "órfãos" (lógica simplificada para demonstração)
	// Na prática, precisaríamos listar todos os pods e checar volumes/envFrom

	// Vamos apenas reportar se encontrarmos ConfigMaps suspeitos de não uso (ex: sufixo .bak)
	for _, cm := range cms.Items {
		if len(cm.Data) == 0 {
			problems = append(problems, types.Problem{
				Kind:      "ConfigMap",
				Namespace: cm.Namespace,
				Name:      cm.Name,
				Error:     "ConfigMap is empty (no data)",
			})
		}
	}

	return problems, nil
}
