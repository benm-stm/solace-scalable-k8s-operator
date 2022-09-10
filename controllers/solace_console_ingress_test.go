package controllers

import (
	"context"
	"strconv"
	"testing"
)

/*
func MockSolaceConsoleReconciler() (*SolaceScalableReconciler, error) {
	solaceIngress := &v1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "solacescalable",
			Namespace: "solacescalable",
		},
		Spec: v1.IngressSpec{},
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(corev1.SchemeGroupVersion, solaceIngress)

	// Create a fake client to mock API calls.
	cl := fake.NewClientBuilder().WithScheme(s).Build()
	if err := cl.Create(context.TODO(), solaceIngress); err != nil {
		return nil, err
	}

	// Create a ReconcileMemcached object with the scheme and fake client.
	return &SolaceScalableReconciler{
		Client: cl,
		Scheme: s,
	}, nil
}*/

func TestNewIngressConsole(t *testing.T) {
	got := NewIngressConsole(
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
	r, _, err := MockHaproxyReconciler()
	if err != nil {
		t.Errorf("object mock fail")
	}

	err = (*r).CreateSolaceConsoleIngress(
		&solaceScalable,
		NewIngressConsole(&solaceScalable, Labels(&solaceScalable)),
		context.TODO(),
	)
	if err != nil {
		t.Errorf("got %v, wanted %v", err, nil)
	}

}
