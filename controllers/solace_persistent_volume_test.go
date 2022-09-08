package controllers

import (
	"context"
	"strconv"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func MockPersistentVolume() (
	*SolaceScalableReconciler,
	*corev1.PersistentVolume,
) {
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pv",
			Namespace: "test",
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceName("pbName"): resource.MustParse("50Gi"),
			},
			PersistentVolumeSource:        corev1.PersistentVolumeSource{},
			AccessModes:                   []corev1.PersistentVolumeAccessMode{},
			ClaimRef:                      &corev1.ObjectReference{},
			PersistentVolumeReclaimPolicy: "",
			StorageClassName:              "localManual",
		},
	}
	//cm := &corev1.ConfigMap{}

	// Objects to track in the fake client.
	objs := []runtime.Object{pv}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(corev1.SchemeGroupVersion, pv)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	return &SolaceScalableReconciler{
		Client: cl,
		Scheme: s,
	}, pv
}

func TestNewPersistentVolume(t *testing.T) {
	got := NewPersistentVolume(&solaceScalable,
		strconv.Itoa(int(solaceScalable.Spec.Replicas)),
		Labels(&solaceScalable),
	)
	if got == nil {
		t.Errorf("got %v, wanted *corev1.PersistentVolume", got)
	}
}

func TestCreateSolaceLocalPv(t *testing.T) {
	r, _ := MockPersistentVolume()
	success, err := r.CreateSolaceLocalPv(
		&solaceScalable,
		NewPersistentVolume(&solaceScalable, "1", Labels(&solaceScalable)),
		context.TODO(),
	)
	if !success || err != nil {
		t.Errorf("got success=%v with pvclass %v", success, solaceScalable.Spec.PvClass)
	}
}
