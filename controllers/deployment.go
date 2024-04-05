package controllers

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

func getDeploymentItems(clientset *kubernetes.Clientset, namespaces []string) ([]ControllerItem, error) {
	var result []ControllerItem
	for _, namespace := range namespaces {
		controllerClient := clientset.AppsV1().Deployments(namespace)
		controllers, err := controllerClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, nil
		}
		for _, controller := range controllers.Items {
			controllerItem := ControllerItem{
				Namespace:      controller.Namespace,
				ControllerType: "Deployment",
				Controller:     controller.Name,
				Replicas:       *controller.Spec.Replicas,
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
