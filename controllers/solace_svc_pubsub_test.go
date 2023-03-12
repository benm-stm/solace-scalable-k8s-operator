package controllers

/*
import (
	"context"
	"strconv"
	"testing"

	libs "github.com/benm-stm/solace-scalable-k8s-operator/common"
	handler "github.com/benm-stm/solace-scalable-k8s-operator/handler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func MockSvc() (
	*SolaceScalableReconciler,
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
	svc := NewSvcPubSub(
		&solaceScalable,
		svcId,
		libs.Labels(&solaceScalable),
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
	return &SolaceScalableReconciler{
		Client: cl,
		Scheme: s,
	}, svc, nil
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
		libs.Labels(&solaceScalable),
	)
	if got.Spec.Ports[0].Port != 1024 {
		t.Errorf("got %v, wanted %v", got.Spec.Ports[0].Port, svcId.TargetPort)
	}
}

func TestCreatePubSubSvc(t *testing.T) {
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

	got := (*r).CreatePubSubSvc(
		&solaceScalable,
		svc,
		context.TODO(),
	)
	if got != nil {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}
func TestAttrSpecificDatasConstruction(t *testing.T) {
	var pubSubSvcNames = []string{}
	var cmData = map[string]string{}
	var svcIds = []SvcId{}
	oP := []handler.SolaceSvcSpec{
		{
			MsgVpnName:     "test",
			ClientUsername: "test",
			Pppo: []handler.Pppo{
				{
					Protocol: "mqtt",
					Port:     int32(1026),
					PubOrSub: "pub",
				},
			},
			AllMsgVpnPorts: []int32{},
		},
	}
	nature := "pub"
	ports := []int32{1024, 1027}
	p := int32(1026)

	AttrSpecificDatasConstruction(
		&solaceScalable,
		&pubSubSvcNames,
		&cmData,
		&svcIds,
		&oP[0].Pppo[0],
		&oP[0],
		p,
		nature,
		&ports,
	)

	//check service name
	wantedSvcName := oP[0].MsgVpnName + "-" +
		oP[0].ClientUsername + "-" +
		"1025-" +
		oP[0].Pppo[0].Protocol + "-" +
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

func TestConstructAttrSpecificDatas(t *testing.T) {
	var pubSubSvcNames = []string{}
	var cmData = map[string]string{}
	var svcIds = []SvcId{}
	oP := []handler.SolaceSvcSpec{
		{
			MsgVpnName:     "test",
			ClientUsername: "test",
			Pppo: []handler.Pppo{
				{
					Protocol:       "mqtt",
					Port:           int32(1026),
					PubOrSub:       "pub",
					OpeningsNumber: 2,
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
		oP[0].Pppo[0],
		oP[0],
		p,
		nature,
		&ports,
	)

	//check services number
	if len(pubSubSvcNames) != int(oP[0].Pppo[0].OpeningsNumber) {
		t.Errorf("got %v, wanted %v", len(pubSubSvcNames), int(oP[0].Pppo[0].OpeningsNumber))
	}

	//check configmaps data number
	if len(cmData) != int(oP[0].Pppo[0].OpeningsNumber) {
		t.Errorf("got %v, wanted %v", len(cmData), int(oP[0].Pppo[0].OpeningsNumber))
	}

	//check svcId
	if len(svcIds) != int(oP[0].Pppo[0].OpeningsNumber) {
		t.Errorf("got %v, wanted %v", len(svcIds), int(oP[0].Pppo[0].OpeningsNumber))
	}
}

func TestConstructSvcDatas(t *testing.T) {
	oP := []handler.SolaceSvcSpec{
		{
			MsgVpnName:     "test",
			ClientUsername: "test",
			Pppo: []handler.Pppo{
				{
					Protocol:       "mqtt",
					Port:           int32(1026),
					PubOrSub:       "pub",
					OpeningsNumber: 1,
				},
			},
			AllMsgVpnPorts: []int32{},
		},
	}
	nature := "pub"
	ports := []int32{1024, 1027}

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
		oP[0].Pppo[0].Protocol + "-" +
		nature

	if (*pubSubSvcNames)[0] != wantedSvcName {
		t.Errorf("got %v, wanted %v", (*pubSubSvcNames)[0], wantedSvcName)
	}

	//check configmap Datas
	wantedCmData := solaceScalable.Namespace + "/" +
		wantedSvcName + ":" +
		strconv.Itoa(int(oP[0].Pppo[0].Port))

	if (*cmData)["1025"] != wantedCmData {
		t.Errorf("got %v, wanted %v", (*cmData)["1025"], wantedCmData)
	}

	//check svcId
	if svcIds[0].TargetPort != int(oP[0].Pppo[0].Port) {
		t.Errorf("got %v, wanted %v", svcIds[0].TargetPort, int(oP[0].Pppo[0].Port))
	}
}

func TestListPubSubSvc(t *testing.T) {
	r, _, err := MockSvc()
	if err != nil {
		t.Errorf("object mock fail")
	}
	gotSvcList, gotErr := (*r).ListPubSubSvc(
		&solaceScalable,
		context.TODO(),
	)
	if len(gotSvcList.Items) != 1 {
		t.Errorf("got %v, wanted svcList, error %v", gotSvcList, gotErr)
	}

}

func TestDeletePubSubSvc(t *testing.T) {
	r, svc, err := MockSvc()
	if err != nil {
		t.Errorf("object mock fail")
	}
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

	if errList == nil || gotErr == nil {
		if errGet == nil {
			t.Errorf("got %v, wanted  %v", foundSvc, nil)
		}
	}
}
*/
