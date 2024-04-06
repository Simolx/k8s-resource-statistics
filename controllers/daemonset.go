package controllers

import (
	"context"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

func getDaemonsetItems(clientset *kubernetes.Clientset, namespaces []string, debugInfo bool) ([]ControllerItem, error) {
	var result []ControllerItem
	for _, namespace := range namespaces {
		controllerClient := clientset.AppsV1().DaemonSets(namespace)
		controllers, err := controllerClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, nil
		}
		if debugInfo {
			if controllerJson, err := json.Marshal(controllers); err != nil {
				return nil, nil
			} else {
				klog.Infof("namespaces %s statefulset:\n%s", namespace, controllerJson)
			}
		}
		for _, controller := range controllers.Items {
			controllerItem := ControllerItem{
				Namespace:      controller.Namespace,
				ControllerType: "Daemonset",
				Controller:     controller.Name,
				Replicas:       1,
			}
			emptyDir, storage, storageNoSize, memStorage := generateVolumeResult(controller.Spec.Template.Spec.Volumes)
			if memStorage {
				klog.Infof("memory EmptyDir, namespace: %q, %s: %q", controllerItem.Namespace, controllerItem.ControllerType, controllerItem.Controller)
			}
			controllerItem.EmptyDir = emptyDir
			controllerItem.Storage = storage
			controllerItem.StorageNoSize = storageNoSize

			controllerItem.Container = generateContainers(controller.Spec.Template.Spec.Containers)
			controllerItem.InitContainer = generateContainers(controller.Spec.Template.Spec.InitContainers)
			result = append(result, controllerItem)
		}
	}
	return result, nil
}
