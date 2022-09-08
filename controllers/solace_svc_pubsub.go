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

func ConstructAttrSpecificDatas(
	s *scalablev1alpha1.SolaceScalable,
	pubSubSvcNames *[]string,
	cmData *map[string]string,
	svcIds *[]SvcId,
	ppp *Ppp,
	oP *SolaceSvcSpec,
	p int32,
	nature string,
	portsArr *[]int32,
) {
	if p != 0 {
		// when ppp nil, it means that no clientusername attribues (pub/sub)
		// are present, so make openings for all msgvpn protocol ports
		if ppp == nil ||
			(ppp != nil && nature == ppp.PubOrSub) {

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
			if ppp != nil {
				svcName = oP.MsgVpnName + "-" +
					oP.ClientUsername + "-" +
					strconv.FormatInt(int64(nextAvailable), 10) + "-" +
					ppp.Protocol + "-" +
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
		for _, ppp := range oP.Ppp {
			for _, p := range ppp.Port {
				if nature == ppp.PubOrSub {
					ConstructAttrSpecificDatas(
						s,
						&pubSubSvcNames,
						&cmData,
						&svcIds,
						&ppp,
						&oP,
						p,
						nature,
						portsArr,
					)
				}
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
