package controllers

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func MockHaproxyReconciler() (
	*SolaceScalableReconciler,
	*corev1.Service,
	error,
) {
	haproxy := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "solacescalable",
			Namespace: "solacescalable",
		},
		Spec: corev1.ServiceSpec{},
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(corev1.SchemeGroupVersion, haproxy)

	// Create a fake client to mock API calls.
	cl := fake.NewClientBuilder().WithScheme(s).Build()
	if err := cl.Create(context.TODO(), haproxy); err != nil {
		return nil, nil, err
	}

	// Create a ReconcileMemcached object with the scheme and fake client.
	return &SolaceScalableReconciler{
		Client: cl,
		Scheme: s,
	}, haproxy, nil
}

func TestNewSvcHaproxy(t *testing.T) {
	got := NewSvcHaproxy(
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
	r, obj, err := MockHaproxyReconciler()
	if err != nil {
		t.Errorf("object mock fail")
	}
	got, err := (*r).GetExistingHaProxySvc(
		&solaceScalable,
		"solacescalable",
		context.TODO(),
	)
	if err != nil {
		t.Errorf("got %v, wanted %v", got, obj)
	}
}

func TestUpdateHAProxySvc(t *testing.T) {
	r, svc, err := MockHaproxyReconciler()
	if err != nil {
		t.Errorf("object mock fail")
	}
	var hashstore = &map[string]string{}
	err = (*r).UpdateHAProxySvc(
		hashstore,
		svc,
		context.TODO(),
	)
	if err != nil {
		t.Errorf("got %v, wanted %v", err, svc)
	}
}
