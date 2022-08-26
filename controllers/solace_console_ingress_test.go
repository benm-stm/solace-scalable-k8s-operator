package controllers

import (
	"context"
	"strconv"
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func MockSolaceConsoleReconciler() *SolaceScalableReconciler {
	solaceIngress := &v1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "solacescalable",
			Namespace: "solacescalable",
		},
		Spec: v1.IngressSpec{},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{solaceIngress}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(corev1.SchemeGroupVersion, solaceIngress)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	return &SolaceScalableReconciler{
		Client: cl,
		Scheme: s,
	}
}

func TestIngressConsole(t *testing.T) {
	got := IngressConsole(
		&solaceScalable,
		Labels(&solaceScalable),
	)
	if got == nil {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}

func TestCreateIngressConsoleRules(t *testing.T) {
	got := CreateIngressConsoleRules(
		&solaceScalable,
	)
	for i := 0; i < int(solaceScalable.Spec.Replicas); i++ {
		if got[0].IngressRuleValue.HTTP.Paths[i].Backend.Service.Name != solaceScalable.ObjectMeta.Namespace+"-"+strconv.Itoa(i) {
			t.Errorf("got %v, wanted %v", got, nil)
		}
	}
}

func TestCreateSolaceConsoleIngress(t *testing.T) {
	r, _, _ := MockHaproxyReconciler()
	err := (*r).CreateSolaceConsoleIngress(
		&solaceScalable,
		context.TODO(),
	)
	if err != nil {
		t.Errorf("got %v, wanted %v", err, nil)
	}

}
