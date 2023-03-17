package statefulset

import (
	"context"
	"testing"

	libs "github.com/benm-stm/solace-scalable-k8s-operator/common"
	"github.com/benm-stm/solace-scalable-k8s-operator/tests"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func MockStatefulset() (
	*tests.SolaceScalableReconciler,
	*v1.StatefulSet,
	error,
) {
	ss := New(
		&tests.SolaceScalable,
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
	return &tests.SolaceScalableReconciler{
		Client: cl,
		Scheme: s,
	}, ss, nil
}
func TestNew(t *testing.T) {
	got := New(
		&tests.SolaceScalable,
		libs.Labels(&tests.SolaceScalable),
	)
	if got == nil {
		t.Errorf("got %v, wanted *v1.StatefulSet", got)
	}
}

func TestCreate(t *testing.T) {
	r, ss, err := MockStatefulset()
	if err != nil {
		t.Errorf("object mock fail %v", err)
	}
	got := Create(
		ss,
		r,
		context.TODO(),
	)
	if got != nil {
		t.Errorf("got %v, wanted *v1.StatefulSet", got)
	}
}

func TestUpdate(t *testing.T) {
	r, ss, err := MockStatefulset()
	if err != nil {
		t.Errorf("object mock fail")
	}
	hashStore := map[string]string{}
	// Case 1: it's the fist launch of the operator
	// There is no saved hash
	err = Update(
		ss,
		r,
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
	err = Update(
		ss,
		r,
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
