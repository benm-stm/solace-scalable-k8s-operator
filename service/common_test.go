package service

import (
	"context"
	"testing"

	libs "github.com/benm-stm/solace-scalable-k8s-operator/common"
	"github.com/benm-stm/solace-scalable-k8s-operator/tests"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func MockSvc() (
	*tests.SolaceScalableReconciler,
	*corev1.Service,
	error,
) {
	svcId := SvcId{
		Name:           "testName",
		ClientUsername: "testClientusername",
		MsgVpnName:     "testMsgVpn",
		Port:           1024,
		TargetPort:     1025,
		Nature:         "pub",
	}
	svc := New(
		&tests.SolaceScalable,
		svcId,
		libs.Labels(&tests.SolaceScalable),
	)

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(corev1.SchemeGroupVersion, svc)

	// Create a fake client to mock API calls.
	cl := fake.NewClientBuilder().WithScheme(s).Build()
	if err := cl.Create(context.TODO(), svc); err != nil {
		return nil, nil, err
	}

	// Create a ReconcileMemcached object with the scheme and fake client.
	return &tests.SolaceScalableReconciler{
		Client: cl,
		Scheme: s,
	}, svc, nil
}

func TestNew(t *testing.T) {
	svcId := SvcId{
		Name:           "test",
		ClientUsername: "test",
		MsgVpnName:     "test",
		Port:           1025,
		TargetPort:     1024,
		Nature:         "pub",
	}
	got := New(
		&tests.SolaceScalable,
		svcId,
		libs.Labels(&tests.SolaceScalable),
	)
	if got.Spec.Ports[0].Port != 1024 {
		t.Errorf("got %v, wanted %v", got.Spec.Ports[0].Port, svcId.TargetPort)
	}
}

func TestCreate(t *testing.T) {
	r, _, err := MockSvc()
	if err != nil {
		t.Errorf("object mock fail")
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "",
					Port: 1882,
					TargetPort: intstr.IntOrString{
						IntVal: 1883,
					},
					NodePort: 0,
				},
			},
		},
	}

	got := Create(
		svc,
		r,
		context.TODO(),
	)
	if got != nil {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}

func TestList(t *testing.T) {
	r, _, err := MockSvc()
	if err != nil {
		t.Errorf("object mock fail")
	}
	gotSvcList, gotErr := List(
		&tests.SolaceScalable,
		r,
		context.TODO(),
	)
	if len(gotSvcList.Items) != 1 {
		t.Errorf("got %v, wanted svcList, error %v", gotSvcList, gotErr)
	}

}

func TestDelete(t *testing.T) {
	r, svc, err := MockSvc()
	if err != nil {
		t.Errorf("object mock fail")
	}
	pubSubSvcNames := []string{"svc1", "svc2"}

	svcList := &corev1.ServiceList{}
	listOptions := &client.ListOptions{Namespace: svc.Namespace}
	ctx := context.TODO()
	errList := r.List(ctx, svcList, listOptions)
	var p int32 = 8080
	gotErr := Delete(
		svcList,
		&pubSubSvcNames,
		&p,
		r,
		ctx,
	)

	foundSvc := &corev1.Service{}
	errGet := r.Get(
		ctx,
		types.NamespacedName{
			Name:      svc.Name,
			Namespace: svc.Namespace,
		},
		foundSvc,
	)

	if errList == nil || gotErr == nil {
		if errGet == nil {
			t.Errorf("got %v, wanted  %v", foundSvc, nil)
		}
	}
}
