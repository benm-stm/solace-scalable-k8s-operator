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
		Container: scalablev1alpha1.Container{
			Image: "solacescalable:test",
			Name:  "solacescalable",
			Env: []corev1.EnvVar{
				{
					Name: "testSecret",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "testSecret",
							},
							Key: "test",
						},
					},
				},
			},
			Volume: scalablev1alpha1.Volume{
				Name:          "volume",
				Size:          "50Gi",
				HostPath:      "/opt/storage",
				ReclaimPolicy: "Retain",
			},
		},
		Replicas:   1,
		ClusterUrl: "scalable.solace.io",
		Haproxy: scalablev1alpha1.Haproxy{
			Namespace: "solacescalable",
		},
		PvClass: "localManual",
		NetWork: scalablev1alpha1.Network{
			StartingAvailablePorts: 1025,
		},
	},
}
var solaceScalableSecret = corev1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "testSecret",
		Namespace: "test",
	},
	Data: map[string][]byte{
		"test": []byte("test"),
	},
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
