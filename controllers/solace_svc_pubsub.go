package controllers

import (
	"context"
	"strconv"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type SvcId struct {
	Name           string
	ClientUsername string
	MsgVpnName     string
	Port           int32
	TargetPort     int
	Nature         string
}

func NewSvcPubSub(
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

func ConstructDatas(
	s *scalablev1alpha1.SolaceScalable,
	pubSubSvcNames *[]string,
	cmData *map[string]string,
	svcIds *[]SvcId,
	pppo *Pppo,
	oP *SolaceSvcSpec,
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
	nextAvailable := NextAvailablePort(
		*portsArr,
		beginningPort,
	)

	svcName := oP.MsgVpnName + "-" +
		oP.ClientUsername + "-" +
		strconv.FormatInt(int64(nextAvailable), 10) + "-" +
		"na" + "-" +
		nature
	if pppo != nil {
		svcName = oP.MsgVpnName + "-" +
			oP.ClientUsername + "-" +
			strconv.FormatInt(int64(nextAvailable), 10) + "-" +
			pppo.Protocol + "-" +
			nature
	}

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

func ConstructAttrSpecificDatas(
	s *scalablev1alpha1.SolaceScalable,
	pubSubSvcNames *[]string,
	cmData *map[string]string,
	svcIds *[]SvcId,
	pppo *Pppo,
	oP *SolaceSvcSpec,
	p int32,
	nature string,
	portsArr *[]int32,
) {
	if p != 0 {
		// when ppp nil, it means that no clientusername attribues (pub/sub)
		// are present, so make openings for all msgvpn protocol ports
		if pppo == nil {
			ConstructDatas(s, pubSubSvcNames, cmData, svcIds, pppo, oP, p, nature, portsArr)
		} else if nature == pppo.PubOrSub {
			for i := 0; i < int(pppo.OpeningsNumber); i++ {
				ConstructDatas(s, pubSubSvcNames, cmData, svcIds, pppo, oP, p, nature, portsArr)
			}
		}
	}
}

// pubsub SVC creation
func ConstructSvcDatas(s *scalablev1alpha1.SolaceScalable,
	pubSubsvcSpecs *[]SolaceSvcSpec,
	nature string,
	portsArr *[]int32,
) (*[]string, *map[string]string, []SvcId) {
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
					&ppp,
					&oP,
					ppp.Port,
					nature,
					portsArr,
				)
			}
		}
		for _, p := range oP.AllMsgVpnPorts {
			ConstructAttrSpecificDatas(
				s,
				&pubSubSvcNames,
				&cmData,
				&svcIds,
				nil,
				&oP,
				p,
				nature,
				portsArr,
			)
		}
	}
	return &pubSubSvcNames, &cmData, svcIds
}

func (r *SolaceScalableReconciler) CreatePubSubSvc(
	s *scalablev1alpha1.SolaceScalable,
	newSvcPubSub *corev1.Service,
	ctx context.Context,
) error {
	log := log.FromContext(ctx)
	foundPubSubSvc := &corev1.Service{}
	if err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      newSvcPubSub.Name,
			Namespace: newSvcPubSub.Namespace,
		},
		foundPubSubSvc,
	); err != nil {
		log.Info("Creating pubSub SVC", newSvcPubSub.Namespace, newSvcPubSub.Name)
		if err = r.Create(ctx, newSvcPubSub); err != nil {
			return err
		}

	}
	return nil
}

func (r *SolaceScalableReconciler) ListPubSubSvc(
	solaceScalable *scalablev1alpha1.SolaceScalable,
	ctx context.Context,
) (*corev1.ServiceList, error) {
	// get existing svc list
	svcList := &corev1.ServiceList{}
	listOptions := &client.ListOptions{Namespace: solaceScalable.Namespace}

	if err := r.List(ctx, svcList, listOptions); err != nil {
		return nil, err
	}
	return svcList, nil
}

func (r *SolaceScalableReconciler) DeletePubSubSvc(
	svcList *corev1.ServiceList,
	pubSubSvcNames *[]string,
	ctx context.Context,
) error {
	log := log.FromContext(ctx)
	for _, s := range svcList.Items {
		if !IsItInSlice(s.Name, *pubSubSvcNames) && s.Spec.Ports[0].Port != 8080 {
			foundExtraPubSubSvc := &corev1.Service{}
			if err := r.Get(
				ctx,
				types.NamespacedName{
					Namespace: s.Namespace,
					Name:      s.Name,
				}, foundExtraPubSubSvc,
			); err == nil {
				log.Info("Delete PubSubSvc", s.Namespace, s.Name)
				if err = r.Delete(ctx, foundExtraPubSubSvc); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

/*
NAME                       TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
solacescalable-0           ClusterIP   10.97.227.171    <none>        8080/TCP   2d2h
solacescalable-1           ClusterIP   10.108.4.216     <none>        8080/TCP   2d2h
solacescalable-2           ClusterIP   10.103.204.8     <none>        8080/TCP   2d2h
test-botti-1025-amqp-sub   ClusterIP   10.106.107.93    <none>        1100/TCP   2d1h
test-botti-1025-mqtt-pub   ClusterIP   10.100.124.186   <none>        1050/TCP   2d1h
test-botti-1026-amqp-pub   ClusterIP   10.103.2.39      <none>        1100/TCP   2d1h
test-default-1026-na-sub   ClusterIP   10.97.8.194      <none>        1100/TCP   2d1h
test-default-1027-na-pub   ClusterIP   10.96.209.139    <none>        1100/TCP   2d1h
test-default-1027-na-sub   ClusterIP   10.98.35.201     <none>        1050/TCP   2d1h
test-default-1028-na-pub   ClusterIP   10.107.172.1     <none>        1050/TCP   2d1h
*/
