package service

import (
	"strconv"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	libs "github.com/benm-stm/solace-scalable-k8s-operator/common"

	spec "github.com/benm-stm/solace-scalable-k8s-operator/service/specs"
)

// Construct services informations
func attrSpecificDatasConstruction(
	s *scalablev1alpha1.SolaceScalable,
	pubSubSvcNames *[]string,
	cmData *map[string]string,
	svcIds *[]SvcId,
	pppo *spec.Pppo,
	oP *spec.SvcSpec,
	p int32,
	nature string,
	portsArr *[]int32,
) {
	// if no default value given to startingAvailablePorts
	// we will attribute the 1st available port offered by the system
	beginningPort := int32(1024)
	if s.Spec.NetWork.StartingAvailablePorts != 0 {
		beginningPort = s.Spec.NetWork.StartingAvailablePorts
	}
	nextAvailable := libs.NextAvailablePort(
		*portsArr,
		beginningPort,
	)

	var protocol string = "na"
	if pppo.Protocol != "" {
		protocol = pppo.Protocol
	}
	svcName := oP.MsgVpnName + "-" +
		oP.ClientUsername + "-" +
		strconv.FormatInt(int64(nextAvailable), 10) + "-" +
		protocol + "-" +
		nature

	*pubSubSvcNames = append(
		*pubSubSvcNames,
		svcName,
	)

	(*cmData)[strconv.Itoa(int(nextAvailable))] = s.Namespace + "/" +
		svcName + ":" +
		strconv.Itoa(int(p))

	*svcIds = append(
		*svcIds,
		SvcId{
			Name:           svcName,
			MsgVpnName:     oP.MsgVpnName,
			ClientUsername: oP.ClientUsername,
			Port:           nextAvailable,
			TargetPort:     int(p),
			Nature:         nature,
		},
	)
	*portsArr = append(*portsArr, nextAvailable)
}

// Take in charge the number of openings per protocol in the
// clientusername attributes (pub/sub)
func ConstructAttrSpecificDatas(
	s *scalablev1alpha1.SolaceScalable,
	pubSubSvcNames *[]string,
	cmData *map[string]string,
	svcIds *[]SvcId,
	pppo spec.Pppo,
	oP spec.SvcSpec,
	p int32,
	nature string,
	portsArr *[]int32,
) {
	if p != 0 {
		// when ppp nil, it means that no clientusername attribues (pub/sub)
		// are present, so make openings for all msgvpn protocol ports
		if (pppo == spec.Pppo{}) {
			attrSpecificDatasConstruction(
				s,
				pubSubSvcNames,
				cmData,
				svcIds,
				&pppo,
				&oP,
				p,
				nature,
				portsArr,
			)
		} else if nature == pppo.PubOrSub {
			// ex: mqtt:2, here we're gonna open 2 ports for mqtt
			for i := 0; i < int(pppo.OpeningsNumber); i++ {
				attrSpecificDatasConstruction(
					s,
					pubSubSvcNames,
					cmData,
					svcIds,
					&pppo,
					&oP,
					p,
					nature,
					portsArr,
				)
			}
		}
	}
}

func NewSvcData() *SvcData {
	return &SvcData{}
}

// pubsub SVC creation
func (sd *SvcData) Set(s *scalablev1alpha1.SolaceScalable,
	pubSubsvcSpecs *[]spec.SvcSpec,
	nature string,
) {
	var portsArr = []int32{}
	var svcIds = []SvcId{}
	var pubSubSvcNames = []string{}
	var cmData = map[string]string{}
	for _, oP := range *pubSubsvcSpecs {
		for _, ppp := range oP.Pppo {
			if nature == ppp.PubOrSub {
				ConstructAttrSpecificDatas(
					s,
					&pubSubSvcNames,
					&cmData,
					&svcIds,
					ppp,
					oP,
					ppp.Port,
					nature,
					&portsArr,
				)
			}
		}
		for _, p := range oP.AllMsgVpnPorts {
			ConstructAttrSpecificDatas(
				s,
				&pubSubSvcNames,
				&cmData,
				&svcIds,
				spec.Pppo{},
				oP,
				p,
				nature,
				&portsArr,
			)
		}
	}
	sd.CmData = cmData
	sd.SvcNames = pubSubSvcNames
	sd.SvcsId = svcIds
	//return &pubSubSvcNames, &cmData, svcIds
}
