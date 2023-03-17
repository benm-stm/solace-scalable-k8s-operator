package configmap

import (
	"context"
	"testing"

	libs "github.com/benm-stm/solace-scalable-k8s-operator/common"
	"github.com/benm-stm/solace-scalable-k8s-operator/tests"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func MockConfigmap() (
	*tests.SolaceScalableReconciler,
	*corev1.ConfigMap,
	error,
) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-test-tcp-ingress",
			Namespace: "test",
		},
		Data: map[string]string{
			"1880": "solasescalable/test-svc-init:1880",
		},
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(corev1.SchemeGroupVersion, cm)

	// Create a fake client to mock API calls.
	cl := fake.NewClientBuilder().WithScheme(s).Build()
	if err := cl.Create(context.TODO(), cm); err != nil {
		return nil, nil, err
	}

	// Create a ReconcileMemcached object with the scheme and fake client.
	return &tests.SolaceScalableReconciler{
		Client: cl,
		Scheme: s,
	}, cm, nil
}

func TestNewtcpConfigmap(t *testing.T) {
	nature := "pub"
	data := map[string]string{
		"1883": "solasescalable/test-svc:1883",
	}
	got := New(
		&tests.SolaceScalable,
		data,
		nature,
		libs.Labels(&tests.SolaceScalable),
	)
	if got.ObjectMeta.Name != tests.SolaceScalable.ObjectMeta.Name+"-"+
		nature+"-tcp-ingress" ||
		got.Data["1883"] != data["1883"] {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}

func TestCreateSolaceTcpConfigmap(t *testing.T) {
	r, _, _ := MockConfigmap()
	//when cm exist
	nature := "test"
	data := map[string]string{
		"1880": "test-srv:1880",
	}

	cm, err := Create(
		&tests.SolaceScalable,
		&data,
		nature,
		r,
		context.TODO(),
	)
	if cm == nil || err != nil {
		t.Errorf("wantedCm %v error %v", cm, err)
	}

	//when cm doesn't exist
	nature = "testNotFound"
	cm, err = Create(
		&tests.SolaceScalable,
		&data,
		nature,
		r,
		context.TODO(),
	)
	if cm != nil || err != nil {
		t.Errorf("wantedCm %v error %v", cm, err)
	}
}

func TestUpdateSolaceTcpConfigmap(t *testing.T) {
	r, cm, err := MockConfigmap()
	if err != nil {
		t.Errorf("object mock fail")
	}
	// when does not exist
	hashStore := map[string]string{}
	err = Update(
		&tests.SolaceScalable,
		cm,
		r,
		context.TODO(),
		&hashStore,
	)
	if hashStore[cm.Name] == "648a4a777504b4e69a1e63ebce71340aeb0d18667f87c88556f618279aaf40d1" {
		t.Errorf(
			"when does not exist : got %v, wanted %v error %v",
			"test",
			hashStore[cm.Name],
			err,
		)
	}

	// when configmap have changed
	hashStore = map[string]string{
		cm.Name: "test",
	}
	err = Update(
		&tests.SolaceScalable,
		cm,
		r,
		context.TODO(),
		&hashStore,
	)
	if hashStore[cm.Name] == "test" {
		t.Errorf("got %v, wanted %v error %v", "test", hashStore[cm.Name], err)
	}
}
