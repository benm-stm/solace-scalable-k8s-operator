package ingress

import (
	"context"
	"testing"

	"github.com/benm-stm/solace-scalable-k8s-operator/tests"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func MockHaproxyReconciler() (
	*tests.SolaceScalableReconciler,
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
	return &tests.SolaceScalableReconciler{
		Client: cl,
		Scheme: s,
	}, haproxy, nil
}

func TestTcp(t *testing.T) {
	got := NewTcp(
		&tests.SolaceScalable,
		tests.Ports,
		map[string]string{
			"1883": "1883",
			"1884": "1884",
		},
	)
	if got == nil {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}

func TestGetDefaultTcp(t *testing.T) {
	got := GetDefaultTcp(tests.Ports)
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

func TestGetTcp(t *testing.T) {
	// serviceobject with metadata
	r, obj, err := MockHaproxyReconciler()
	if err != nil {
		t.Errorf("object mock fail")
	}
	got, err := GetTcp(
		&tests.SolaceScalable,
		"solacescalable",
		r,
		context.TODO(),
	)
	if err != nil {
		t.Errorf("got %v, wanted %v", got, obj)
	}
}

func TestUpdateTcp(t *testing.T) {
	r, svc, err := MockHaproxyReconciler()
	if err != nil {
		t.Errorf("object mock fail")
	}
	var hashstore = &map[string]string{}
	err = UpdateTcp(
		hashstore,
		svc,
		r,
		context.TODO(),
	)
	if err != nil {
		t.Errorf("got %v, wanted %v", err, svc)
	}
}
