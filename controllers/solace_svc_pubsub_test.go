package controllers

import (
	"context"
	"strconv"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func MockSvc() (
	*SolaceScalableReconciler,
	*corev1.Service,
) {
	svcId := SvcId{
		Name:           "testName",
		ClientUsername: "testClientusername",
		MsgVpnName:     "testMsgVpn",
		Port:           1024,
		TargetPort:     1025,
		Nature:         "pub",
	}
	svc := NewSvcPubSub(
		&solaceScalable,
		svcId,
		Labels(&solaceScalable),
	)

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

func TestNewSvcPubSub(t *testing.T) {
	svcId := SvcId{
		Name:           "test",
		ClientUsername: "test",
		MsgVpnName:     "test",
		Port:           1025,
		TargetPort:     1024,
		Nature:         "pub",
	}
	got := NewSvcPubSub(
		&solaceScalable,
		svcId,
		Labels(&solaceScalable),
	)
	if got.Spec.Ports[0].Port != 1024 {
		t.Errorf("got %v, wanted %v", got.Spec.Ports[0].Port, svcId.TargetPort)
	}
}

func TestCreatePubSubSvc(t *testing.T) {
	r, _ := MockSvc()
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

	got := (*r).CreatePubSubSvc(
		&solaceScalable,
		svc,
		context.TODO(),
	)
	if got != nil {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}
func TestConstructAttrSpecificDatas(t *testing.T) {
	var pubSubSvcNames = []string{}
	var cmData = map[string]string{}
	var svcIds = []SvcId{}
	oP := []SolaceSvcSpec{
		{
			MsgVpnName:     "test",
			ClientUsername: "test",
			Ppp: []Ppp{
				{
					Protocol: "mqtt",
					Port: []int32{
						int32(1026),
					},
					PubOrSub: "pub",
				},
			},
			AllMsgVpnPorts: []int32{},
		},
	}
	nature := "pub"
	ports := []int32{1024, 1027}
	p := int32(1026)

	ConstructAttrSpecificDatas(
		&solaceScalable,
		&pubSubSvcNames,
		&cmData,
		&svcIds,
		&oP[0].Ppp[0],
		&oP[0],
		p,
		"pub",
		&ports,
	)

	//check service name
	wantedSvcName := oP[0].MsgVpnName + "-" +
		oP[0].ClientUsername + "-" +
		"1025-" +
		oP[0].Ppp[0].Protocol + "-" +
		nature

	if pubSubSvcNames[0] != wantedSvcName {
		t.Errorf("got %v, wanted %v", pubSubSvcNames[0], wantedSvcName)
	}

	//check configmap Datas
	wantedCmData := solaceScalable.Namespace + "/" +
		wantedSvcName + ":" +
		strconv.Itoa(int(p))

	if cmData["1025"] != wantedCmData {
		t.Errorf("got %v, wanted %v", cmData["1025"], wantedCmData)
	}

	//check svcId
	if svcIds[0].TargetPort != int(p) {
		t.Errorf("got %v, wanted %v", svcIds[0].TargetPort, p)
	}
}

func TestConstructSvcDatas(t *testing.T) {
	oP := []SolaceSvcSpec{
		{
			MsgVpnName:     "test",
			ClientUsername: "test",
			Ppp: []Ppp{
				{
					Protocol: "mqtt",
					Port: []int32{
						int32(1026),
					},
					PubOrSub: "pub",
				},
			},
			AllMsgVpnPorts: []int32{},
		},
	}
	nature := "pub"
	ports := []int32{1024, 1027}
	//p := int32(1026)

	pubSubSvcNames, cmData, svcIds := ConstructSvcDatas(
		&solaceScalable,
		&oP,
		nature,
		&ports,
	)

	//check service name
	wantedSvcName := oP[0].MsgVpnName + "-" +
		oP[0].ClientUsername + "-" +
		"1025-" +
		oP[0].Ppp[0].Protocol + "-" +
		nature

	if (*pubSubSvcNames)[0] != wantedSvcName {
		t.Errorf("got %v, wanted %v", (*pubSubSvcNames)[0], wantedSvcName)
	}

	//check configmap Datas
	wantedCmData := solaceScalable.Namespace + "/" +
		wantedSvcName + ":" +
		strconv.Itoa(int(oP[0].Ppp[0].Port[0]))

	if (*cmData)["1025"] != wantedCmData {
		t.Errorf("got %v, wanted %v", (*cmData)["1025"], wantedCmData)
	}

	//check svcId
	if svcIds[0].TargetPort != int(oP[0].Ppp[0].Port[0]) {
		t.Errorf("got %v, wanted %v", svcIds[0].TargetPort, int(oP[0].Ppp[0].Port[0]))
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