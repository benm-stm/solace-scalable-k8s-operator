package controllers

import (
	"context"
	"strconv"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func MockSvc() (
	*SolaceScalableReconciler,
	*corev1.Service,
) {
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-svc",
			Namespace: "test-namespace",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{},
			Ports: []corev1.ServicePort{
				{
					Protocol: "http",
					Port:     8082,
				},
				{
					Protocol: "tcp",
					Port:     8081,
				},
			},
			Type: "ClusterIP",
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{svc}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(corev1.SchemeGroupVersion, svc)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	return &SolaceScalableReconciler{
		Client: cl,
		Scheme: s,
	}, svc
}

func TestSvcPubSub(t *testing.T) {
	var wantedPort int32 = 8080
	got := NewSvcPubSub(
		&solaceScalable,
		"testMsgVpn",
		"testClientUsername",
		wantedPort,
		"pub",
		Labels(&solaceScalable),
	)
	if got.Spec.Ports[0].Port != 8080 {
		t.Errorf("got %v, wanted %v", got, wantedPort)
	}
}

func TestCreatePubSubSvc(t *testing.T) {
	r, _ := MockSvc()
	svcId := []SvcId{
		{
			ClientUsername: "testClientUsername",
			MsgVpnName:     "testMsgVpn",
			Port:           8080,
			Nature:         "testPub",
		},
		{
			ClientUsername: "testClientUsername2",
			MsgVpnName:     "testMsgVpn2",
			Port:           8081,
			Nature:         "testPub",
		},
	}
	got := (*r).CreatePubSubSvc(
		&solaceScalable,
		NewSvcPubSub(
			&solaceScalable,
			svcId[0].MsgVpnName,
			svcId[0].ClientUsername,
			svcId[0].Port,
			svcId[0].Nature,
			Labels(&solaceScalable)),
		context.TODO(),
	)
	if got != nil {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}

func TestConstructSvcDatas(t *testing.T) {
	smr := []SolaceMergedResp{
		{
			MsgVpnName:     "testMsgVpn1",
			ClientUsername: "testClientUsername1",
			Ports:          []int32{8000, 8001},
		},
		{
			MsgVpnName:     "testMsgVpn2",
			ClientUsername: "testClientUsername2",
			Ports:          []int32{8003, 8004},
		},
	}
	nature := "pub"

	gotPubSubSvcNames, gotCmData, gotSvcIds := ConstructSvcDatas(
		&solaceScalable,
		&smr,
		nature,
	)
	wantedSmr := smr[0].MsgVpnName + "-" +
		smr[0].ClientUsername + "-" +
		strconv.Itoa(int(smr[0].Ports[0])) + "-" +
		nature

	if (*gotPubSubSvcNames)[0] != wantedSmr {
		t.Errorf("got %v, wanted %v", (*gotPubSubSvcNames)[0], wantedSmr)
	}

	//	8000:test/testMsgVpn1-testClientUsername1-8000-pub:8000
	wantedDataCm := solaceScalable.Namespace + "/" +
		smr[0].MsgVpnName + "-" +
		smr[0].ClientUsername + "-" +
		strconv.Itoa(int(smr[0].Ports[0])) + "-" +
		nature + ":" +
		strconv.Itoa(int(smr[0].Ports[0]))

	if (*gotCmData)[strconv.Itoa(int(smr[0].Ports[0]))] != wantedDataCm {
		t.Errorf("got %v, wanted %v",
			(*gotCmData)[strconv.Itoa(int(smr[0].Ports[0]))],
			wantedDataCm,
		)
	}

	// {testClientUsername1 testMsgVpn1 8000 pub}
	wantedSvcId := SvcId{
		ClientUsername: smr[0].ClientUsername,
		MsgVpnName:     smr[0].MsgVpnName,
		Port:           smr[0].Ports[0],
		Nature:         nature,
	}
	if gotSvcIds[0].ClientUsername != wantedSvcId.ClientUsername ||
		gotSvcIds[0].MsgVpnName != wantedSvcId.MsgVpnName ||
		gotSvcIds[0].Port != wantedSvcId.Port ||
		gotSvcIds[0].Nature != wantedSvcId.Nature {
		t.Errorf("got %v, wanted %v", gotSvcIds[0], wantedSvcId)
	}
}

func TestListPubSubSvc(t *testing.T) {
	r, _ := MockSvc()
	gotSvcList, gotErr := (*r).ListPubSubSvc(
		&solaceScalable,
		context.TODO(),
	)
	if len(gotSvcList.Items) != 1 {
		t.Errorf("got %v, wanted svcList, error %v", gotSvcList, gotErr)
	}

}

func TestDeletePubSubSvc(t *testing.T) {
	r, svc := MockSvc()
	pubSubSvcNames := []string{"svc1", "svc2"}

	svcList := &corev1.ServiceList{}
	listOptions := &client.ListOptions{Namespace: svc.Namespace}
	ctx := context.TODO()
	errList := r.List(ctx, svcList, listOptions)

	gotErr := (*r).DeletePubSubSvc(
		svcList,
		&pubSubSvcNames,
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

	//t.Errorf("list %v, \n\nget  %v\n\n\n\n", svcList.Items, foundSvc)
	if errList == nil || gotErr == nil {
		if errGet == nil {
			t.Errorf("got %v, wanted  %v", foundSvc, nil)
		}
	}
}
