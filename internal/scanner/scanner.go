package scanner

import (
	"context"
	"fmt"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/jorgegabrielti/mekhanikube/pkg/types"
)

// Scanner é responsável por escanear recursos do cluster
type Scanner struct {
	clientset *kubernetes.Clientset
}

// New cria um novo Scanner
func New() (*Scanner, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar clientset: %w", err)
	}

	return &Scanner{clientset: clientset}, nil
}

// getKubeConfig tenta obter a configuração do Kubernetes
func getKubeConfig() (*rest.Config, error) {
	// Tenta config in-cluster primeiro (quando rodando dentro do cluster)
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Se não estiver in-cluster, usa kubeconfig modificado (para Docker)
	kubeconfigPath := "/root/.kube/config_mod"
	if _, err := filepath.Abs(kubeconfigPath); err != nil {
		// Fallback para kubeconfig padrão
		kubeconfigPath = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// ScanPods escaneia todos os pods e retorna problemas encontrados
func (s *Scanner) ScanPods(ctx context.Context, namespace string) ([]types.Problem, error) {
	var problems []types.Problem

	pods, err := s.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("erro ao listar pods: %w", err)
	}

	for _, pod := range pods.Items {
		// Verifica status do pod
		if problem := s.checkPodStatus(&pod); problem != nil {
			problems = append(problems, *problem)
		}

		// Verifica containers
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if problem := s.checkContainerStatus(&pod, &containerStatus); problem != nil {
				problems = append(problems, *problem)
			}
		}
	}

	return problems, nil
}

// checkPodStatus verifica o status geral do pod
func (s *Scanner) checkPodStatus(pod *corev1.Pod) *types.Problem {
	switch pod.Status.Phase {
	case corev1.PodPending:
		if len(pod.Status.Conditions) > 0 {
			for _, condition := range pod.Status.Conditions {
				if condition.Status == corev1.ConditionFalse {
					return &types.Problem{
						Kind:      "Pod",
						Namespace: pod.Namespace,
						Name:      pod.Name,
						Error:     fmt.Sprintf("%s: %s", condition.Reason, condition.Message),
						Details:   []string{string(pod.Status.Phase)},
					}
				}
			}
		}
	case corev1.PodFailed:
		return &types.Problem{
			Kind:      "Pod",
			Namespace: pod.Namespace,
			Name:      pod.Name,
			Error:     fmt.Sprintf("Pod failed: %s", pod.Status.Message),
			Details:   []string{string(pod.Status.Phase), pod.Status.Reason},
		}
	}

	return nil
}

// checkContainerStatus verifica o status dos containers
func (s *Scanner) checkContainerStatus(pod *corev1.Pod, status *corev1.ContainerStatus) *types.Problem {
	// CrashLoopBackOff
	if status.State.Waiting != nil && status.State.Waiting.Reason == "CrashLoopBackOff" {
		return &types.Problem{
			Kind:      "Pod",
			Namespace: pod.Namespace,
			Name:      pod.Name,
			Error:     fmt.Sprintf("Container %s in CrashLoopBackOff", status.Name),
			Details:   []string{status.State.Waiting.Message, fmt.Sprintf("Restart count: %d", status.RestartCount)},
		}
	}

	// ImagePullBackOff
	if status.State.Waiting != nil && (status.State.Waiting.Reason == "ImagePullBackOff" || status.State.Waiting.Reason == "ErrImagePull") {
		return &types.Problem{
			Kind:      "Pod",
			Namespace: pod.Namespace,
			Name:      pod.Name,
			Error:     fmt.Sprintf("Container %s cannot pull image: %s", status.Name, status.Image),
			Details:   []string{status.State.Waiting.Message},
		}
	}

	// Container terminated with error
	if status.State.Terminated != nil && status.State.Terminated.ExitCode != 0 {
		return &types.Problem{
			Kind:      "Pod",
			Namespace: pod.Namespace,
			Name:      pod.Name,
			Error:     fmt.Sprintf("Container %s terminated with exit code %d", status.Name, status.State.Terminated.ExitCode),
			Details:   []string{status.State.Terminated.Reason, status.State.Terminated.Message},
		}
	}

	return nil
}

// ScanConfigMaps escaneia ConfigMaps não utilizados
func (s *Scanner) ScanConfigMaps(ctx context.Context, namespace string) ([]types.Problem, error) {
	var problems []types.Problem

	configMaps, err := s.clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("erro ao listar configmaps: %w", err)
	}

	pods, err := s.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("erro ao listar pods: %w", err)
	}

	// Cria mapa de ConfigMaps usados
	usedConfigMaps := make(map[string]bool)
	for _, pod := range pods.Items {
		// Verifica volumes
		for _, volume := range pod.Spec.Volumes {
			if volume.ConfigMap != nil {
				key := fmt.Sprintf("%s/%s", pod.Namespace, volume.ConfigMap.Name)
				usedConfigMaps[key] = true
			}
		}

		// Verifica envFrom
		for _, container := range pod.Spec.Containers {
			for _, envFrom := range container.EnvFrom {
				if envFrom.ConfigMapRef != nil {
					key := fmt.Sprintf("%s/%s", pod.Namespace, envFrom.ConfigMapRef.Name)
					usedConfigMaps[key] = true
				}
			}
		}
	}

	// Verifica ConfigMaps não utilizados
	for _, cm := range configMaps.Items {
		key := fmt.Sprintf("%s/%s", cm.Namespace, cm.Name)
		if !usedConfigMaps[key] {
			problems = append(problems, types.Problem{
				Kind:      "ConfigMap",
				Namespace: cm.Namespace,
				Name:      cm.Name,
				Error:     fmt.Sprintf("ConfigMap %s is not used by any pods in the namespace", cm.Name),
				Details:   []string{},
			})
		}
	}

	return problems, nil
}
