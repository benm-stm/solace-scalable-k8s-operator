package controllers

import (
	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var solaceScalable = scalablev1alpha1.SolaceScalable{
	TypeMeta: metav1.TypeMeta{},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test",
		Namespace: "test",
		Labels:    map[string]string{},
	},
	Spec: scalablev1alpha1.SolaceScalableSpec{
		Replicas: 1,
		Haproxy: scalablev1alpha1.Haproxy{
			Namespace: "solacescalable",
		},
	},
	Status: scalablev1alpha1.SolaceScalableStatus{},
}

var appProtocol = "http"
var ports = []corev1.ServicePort{
	{
		Name:     "port1",
		Protocol: "tcp",
		Port:     1883,
	},
	{
		Name:        "http",
		Protocol:    "http",
		Port:        1884,
		AppProtocol: &appProtocol,
	},
}
