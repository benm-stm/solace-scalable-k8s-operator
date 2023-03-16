package service

import (
	"context"
	"strconv"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	libs "github.com/benm-stm/solace-scalable-k8s-operator/common"

	spec "github.com/benm-stm/solace-scalable-k8s-operator/service/specs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func NewSvc(
	s *scalablev1alpha1.SolaceScalable,
	svc SvcId,
	labels map[string]string,
) *corev1.Service {

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: s.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol: corev1.ProtocolTCP,
				Port:     int32(svc.TargetPort),
			}},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// Construct services informations
func AttrSpecificDatasConstruction(
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
			AttrSpecificDatasConstruction(
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
				AttrSpecificDatasConstruction(
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
	//portsArr *[]int32,
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

func Create(
	s *scalablev1alpha1.SolaceScalable,
	newSvcPubSub *corev1.Service,
	k k8sClient,
	ctx context.Context,
) error {
	log := log.FromContext(ctx)
	foundPubSubSvc := &corev1.Service{}
	if err := k.Get(
		ctx,
		types.NamespacedName{
			Name:      newSvcPubSub.Name,
			Namespace: newSvcPubSub.Namespace,
		},
		foundPubSubSvc,
	); err != nil {
		log.Info("Creating pubSub SVC", newSvcPubSub.Namespace, newSvcPubSub.Name)
		if err = k.Create(ctx, newSvcPubSub); err != nil {
			return err
		}

	}
	return nil
}

func List(
	solaceScalable *scalablev1alpha1.SolaceScalable,
	k k8sClient,
	ctx context.Context,
) (*corev1.ServiceList, error) {
	// get existing svc list
	svcList := &corev1.ServiceList{}
	listOptions := &client.ListOptions{Namespace: solaceScalable.Namespace}

	if err := k.List(ctx, svcList, listOptions); err != nil {
		return nil, err
	}
	return svcList, nil
}

func Delete(
	svcList *corev1.ServiceList,
	pubSubSvcNames *[]string,
	k k8sClient,
	ctx context.Context,
) error {
	log := log.FromContext(ctx)
	for _, s := range svcList.Items {
		if !libs.IsItInSlice(s.Name, *pubSubSvcNames) && s.Spec.Ports[0].Port != 8080 {
			foundExtraPubSubSvc := &corev1.Service{}
			if err := k.Get(
				ctx,
				types.NamespacedName{
					Namespace: s.Namespace,
					Name:      s.Name,
				}, foundExtraPubSubSvc,
			); err == nil {
				log.Info("Delete PubSubSvc", s.Namespace, s.Name)
				if err = k.Delete(ctx, foundExtraPubSubSvc); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
