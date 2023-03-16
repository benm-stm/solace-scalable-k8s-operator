package controllers

/*
import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func MockSecret() (
	*SolaceScalableReconciler,
	*corev1.Secret,
	error,
) {
	sec := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test",
		},
		Immutable: new(bool),
		Data: map[string][]byte{
			"test": []byte("test"),
		},
		Type: "",
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(corev1.SchemeGroupVersion, sec)

	// Create a fake client to mock API calls.
	cl := fake.NewClientBuilder().WithScheme(s).Build()
	if err := cl.Create(context.TODO(), sec); err != nil {
		return nil, nil, err
	}

	// Create a ReconcileMemcached object with the scheme and fake client.
	return &SolaceScalableReconciler{
		Client: cl,
		Scheme: s,
	}, sec, nil
}

func TestGetSolaceSecret(t *testing.T) {
	r, _, err := MockSecret()
	if err != nil {
		t.Errorf("object mock fail")
	}
	got, err := r.GetSolaceSecret(&solaceScalable,
		context.TODO(),
	)
	if err == nil {
		t.Errorf("got %v, wanted *corev1.Secret", got)
	}
}

func TestGetSecretFromKey(t *testing.T) {
	got, err := GetSecretFromKey(&solaceScalable,
		&solaceScalableSecret,
		"testSecret",
	)
	if got == "" || err != nil {
		t.Errorf("got %v, wanted test", got)
	}
}

func TestGetSecretEnvIndex(t *testing.T) {
	got, err := GetSecretEnvIndex(&solaceScalable,
		"testSecret",
	)
	if got == -1 || err != nil {
		t.Errorf("got %v, wanted unsigned int index", got)
	}
}
*/
