package service

import (
	"strconv"
	"testing"

	specs "github.com/benm-stm/solace-scalable-k8s-operator/service/specs"
	"github.com/benm-stm/solace-scalable-k8s-operator/tests"
)

func TestAttrSpecificDatasConstruction(t *testing.T) {
	var pubSubSvcNames = []string{}
	var cmData = map[string]string{}
	var svcIds = []SvcId{}
	oP := []specs.SvcSpec{
		{
			MsgVpnName:     "test",
			ClientUsername: "test",
			Pppo: []specs.Pppo{
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

	attrSpecificDatasConstruction(
		&tests.SolaceScalable,
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
	wantedCmData := tests.SolaceScalable.Namespace + "/" +
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
	oP := []specs.SvcSpec{
		{
			MsgVpnName:     "test",
			ClientUsername: "test",
			Pppo: []specs.Pppo{
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
		&tests.SolaceScalable,
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

/*
func TestSet(t *testing.T) {
	oP := specs.SvcSpec{
		MsgVpnName:     "test",
		ClientUsername: "test",
		Pppo: []specs.Pppo{
			{
				Protocol:       "mqtt",
				Port:           int32(1026),
				PubOrSub:       "pub",
				OpeningsNumber: 1,
			},
		},
		AllMsgVpnPorts: []int32{},
	}

	sD := &SvcData{
		SvcNames: []string{"test1", "test2", "test3", "test4", "test5"},
		CmData:   map[string]string{"8080": "testA", "443": "testB"},
		SvcsId:   []SvcId{},
	}
	nature := "pub"
	//ports := []int32{1024, 1027}

	pubSubSvcNames, cmData, svcIds := &sD.Set(
		&tests.SolaceScalable,
		&oP,
		nature,
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
	wantedCmData := tests.SolaceScalable.Namespace + "/" +
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
*/
