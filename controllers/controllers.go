package controllers

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
)

const (
	mi = 1024 * 1024
)

type ContainerItem struct {
	Name                    string `json:"name,omitempty"`
	RequestCPU              int64  `json:"requestCpu"`
	RequestMem              int64  `json:"requestMem"`
	RequestEphemeralStorate int64  `json:"requestEphemeralStorate,omitempty"`
	LimitCPU                int64  `json:"limitCpu"`
	LimitMem                int64  `json:"limitMem"`
	LimitEphemeralStorate   int64  `json:"limitEphemeralStorate,omitempty"`
}

type ControllerItem struct {
	Namespace      string          `json:"namespace,omitempty"`
	ControllerType string          `json:"controllerType,omitempty"`
	Controller     string          `json:"controller,omitempty"`
	Replicas       int32           `json:"replicas,omitempty"`
	InitContainer  []ContainerItem `json:"initContainer,omitempty"`
	Container      []ContainerItem `json:"container,omitempty"`
	EmptyDir       int64           `json:"emptyDir,omitempty"`
	Storage        int             `json:"storage,omitempty"`
	StorageNoSize  bool            `json:"storageNoSize,omitempty"`
}

func ConvertResultToCsv(content []ControllerItem) [][]string {
	result := [][]string{[]string{
		"namespace", "controllerType", "controller", "replicas", "emptyDir(m)", "storage(m)", "storageNoSize",
		"containerType", "containerName", "requestCpu", "requestMem(m)", "requestEphemeralStorage(m)", "limitCpu", "limitMem(m)", "limitEphemeralStorage(m)"}}
	for _, controller := range content {
		namespace := controller.Namespace
		controllerType := controller.ControllerType
		controllerName := controller.Controller
		replicas := controller.Replicas
		emptyDir := controller.EmptyDir
		storage := controller.Storage
		storageNoSize := controller.StorageNoSize
		containerType := "initContainer"
		for _, container := range controller.InitContainer {
			result = append(result,
				[]string{
					namespace, controllerType, controllerName, strconv.Itoa(int(replicas)), strconv.FormatInt(emptyDir, 10), strconv.Itoa(storage), strconv.FormatBool(storageNoSize),
					containerType, container.Name, strconv.FormatInt(container.RequestCPU, 10), strconv.FormatInt(container.RequestMem, 10), strconv.FormatInt(container.RequestEphemeralStorate, 10),
					strconv.FormatInt(container.LimitCPU, 10), strconv.FormatInt(container.LimitMem, 10), strconv.FormatInt(container.LimitEphemeralStorate, 10),
				})
		}
		containerType = "container"
		for _, container := range controller.Container {
			result = append(result,
				[]string{
					namespace, controllerType, controllerName, strconv.Itoa(int(replicas)), strconv.FormatInt(emptyDir, 10), strconv.Itoa(storage), strconv.FormatBool(storageNoSize),
					containerType, container.Name, strconv.FormatInt(container.RequestCPU, 10), strconv.FormatInt(container.RequestMem, 10), strconv.FormatInt(container.RequestEphemeralStorate, 10),
					strconv.FormatInt(container.LimitCPU, 10), strconv.FormatInt(container.LimitMem, 10), strconv.FormatInt(container.LimitEphemeralStorate, 10),
				})
		}
	}
	return result

}

func generateVolumeResult(volumes []v1.Volume) (int64, int, bool, bool) {
	var emptyDir int64 = 0
	storage, storageNoSize, memStorage := 0, false, false
	for _, volume := range volumes {
		if volume.EmptyDir != nil {
			if volume.EmptyDir.Medium != "" {
				memStorage = true
			} else if volume.EmptyDir.SizeLimit == nil || volume.EmptyDir.SizeLimit.Value() == 0 {
				storageNoSize = true
			} else {
				emptyDir += volume.EmptyDir.SizeLimit.Value()
			}
		}
		if volume.CSI != nil {
			if volume.CSI.Size() == 0 {
				storageNoSize = true
			} else {
				storage += volume.CSI.Size()
			}
		}
	}
	return emptyDir / mi, storage / mi, storageNoSize, memStorage
}

func generateContainers(containers []v1.Container) []ContainerItem {
	var containerItems []ContainerItem
	for _, container := range containers {
		containerItems = append(containerItems, ContainerItem{
			Name:                    container.Name,
			RequestCPU:              container.Resources.Requests.Cpu().MilliValue(),
			RequestMem:              container.Resources.Requests.Memory().Value() / mi,
			RequestEphemeralStorate: container.Resources.Requests.StorageEphemeral().Value() / mi,
			LimitCPU:                container.Resources.Limits.Cpu().MilliValue(),
			LimitMem:                container.Resources.Limits.Memory().Value() / mi,
			LimitEphemeralStorate:   container.Resources.Limits.StorageEphemeral().Value() / mi,
		})
	}
	return containerItems
}

func GetNamespaces(clientset *kubernetes.Clientset) ([]string, error) {
	if namespaceList, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{}); err != nil {
		return nil, err
	} else {
		namespaces := make([]string, 0, len(namespaceList.Items))
		for _, namespace := range namespaceList.Items {
			namespaces = append(namespaces, namespace.Name)
		}
		return namespaces, nil
	}
}

func GetControllerItems(clientset *kubernetes.Clientset, namespaces []string) ([]ControllerItem, error) {
	var result []ControllerItem
	if deployments, err := getDeploymentItems(clientset, namespaces); err != nil {
		return result, err
	} else {
		result = append(result, deployments...)
	}
	if statefulsets, err := getStatefulsetItems(clientset, namespaces); err != nil {
		return result, err
	} else {
		result = append(result, statefulsets...)
	}
	if daemonsets, err := getDaemonsetItems(clientset, namespaces); err != nil {
		return result, err
	} else {
		result = append(result, daemonsets...)
	}
	return result, nil
}
