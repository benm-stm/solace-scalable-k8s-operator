package controllers

import (
	"context"
	"testing"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func MockStatefulset() (
	*SolaceScalableReconciler,
	*v1.StatefulSet,
	error,
) {
	ss := NewStatefulset(
		&solaceScalable,
		map[string]string{
			"app": "test",
		},
	)

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(v1.SchemeGroupVersion, ss)

	// Create a fake client to mock API calls.
	cl := fake.NewClientBuilder().WithScheme(s).Build()
	if err := cl.Create(context.TODO(), ss); err != nil {
		return nil, nil, err
	}

	// Create a ReconcileMemcached object with the scheme and fake client.
	return &SolaceScalableReconciler{
		Client: cl,
		Scheme: s,
	}, ss, nil
}
func TestNewStatefulset(t *testing.T) {
	got := NewStatefulset(
		&solaceScalable,
		Labels(&solaceScalable),
	)
	if got == nil {
		t.Errorf("got %v, wanted *v1.StatefulSet", got)
	}
}

func TestCreateStatefulSet(t *testing.T) {
	r, ss, err := MockStatefulset()
	if err != nil {
		t.Errorf("object mock fail %v", err)
	}
	got := (*r).CreateStatefulSet(
		ss,
		context.TODO(),
	)
	if got != nil {
		t.Errorf("got %v, wanted *v1.StatefulSet", got)
	}
}

func TestUpdateStatefulSet(t *testing.T) {
	r, ss, err := MockStatefulset()
	if err != nil {
		t.Errorf("object mock fail")
	}
	hashStore := map[string]string{}
	// Case 1: it's the fist launch of the operator
	// There is no saved hash
	err = (*r).UpdateStatefulSet(
		ss,
		context.TODO(),
		&hashStore,
	)
	if err != nil || hashStore[ss.Name] == "" {
		t.Errorf("got %v, wanted nil, error %v",
			hashStore[ss.Name],
			err,
		)
	}
	// Case 2: hash exists, su update statefulset
	oldHash := hashStore[ss.Name]
	var replicas int32 = 10
	ss.Spec.Replicas = &replicas
	err = (*r).UpdateStatefulSet(
		ss,
		context.TODO(),
		&hashStore,
	)
	if err != nil || hashStore[ss.Name] == oldHash {
		t.Errorf("got %v, wanted not %v, error %v",
			hashStore[ss.Name],
			oldHash,
			err,
		)
	}
}
