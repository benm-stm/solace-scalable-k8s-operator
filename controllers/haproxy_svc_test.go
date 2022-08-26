package controllers

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func MockHaproxyReconciler() (*SolaceScalableReconciler, []runtime.Object, *corev1.Service) {
	haproxy := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "solacescalable",
			Namespace: "solacescalable",
		},
		Spec: corev1.ServiceSpec{},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{haproxy}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(corev1.SchemeGroupVersion, haproxy)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	return &SolaceScalableReconciler{
		Client: cl,
		Scheme: s,
	}, objs, haproxy
}

func TestSvcHaproxy(t *testing.T) {
	got := SvcHaproxy(
		&solaceScalable,
		ports,
		map[string]string{
			"1883": "1883",
			"1884": "1884",
		},
	)
	if got == nil {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}

func TestGetDefaultHaProxyConf(t *testing.T) {
	got := GetDefaultHaProxyConf(ports)
	want := &[]corev1.ServicePort{
		{
			Name:        "port1",
			Protocol:    "http",
			AppProtocol: nil,
			Port:        1884,
		},
	}
	if (*got)[0].AppProtocol != (*want)[0].AppProtocol {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}

func TestGetExistingHaProxySvc(t *testing.T) {
	// serviceobject with metadata
	r, objs, _ := MockHaproxyReconciler()
	got, err := (*r).GetExistingHaProxySvc(
		&solaceScalable,
		"solacescalable",
		context.TODO(),
	)
	if err != nil {
		t.Errorf("got %v, wanted %v", got, objs)
	}
}

func TestUpdateHAProxySvc(t *testing.T) {
	r, objs, svc := MockHaproxyReconciler()
	var hashstore = &map[string]string{}
	err := (*r).UpdateHAProxySvc(
		hashstore,
		svc,
		context.TODO(),
	)
	if err != nil {
		t.Errorf("got %v, wanted %v", err, objs)
	}
}
