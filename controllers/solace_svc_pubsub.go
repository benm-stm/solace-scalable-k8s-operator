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
	ClientUsername string
	MsgVpnName     string
	Port           int32
	Nature         string
}

func NewSvcPubSub(
	s *scalablev1alpha1.SolaceScalable,
	msgVpnName string,
	clientUsername string,
	port int32,
	pubSub string,
	labels map[string]string,
) *corev1.Service {

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: msgVpnName + "-" +
				clientUsername + "-" +
				strconv.FormatInt(int64(port), 10) + "-" +
				pubSub,
			Namespace: s.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol: corev1.ProtocolTCP,
				Port:     port,
			}},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// pubsub SVC creation
func ConstructSvcDatas(s *scalablev1alpha1.SolaceScalable,
	pubSubOpenPorts *[]SolaceMergedResp,
	nature string,
) (*[]string, *map[string]string, []SvcId) {
	var svcIds = []SvcId{}
	var pubSubSvcNames = []string{}
	var cmData = map[string]string{}
	for _, oP := range *pubSubOpenPorts {
		for _, p := range oP.Ports {
			if p != 0 {
				//create pubsub SVC
				pubSubSvcNames = append(
					pubSubSvcNames,
					oP.MsgVpnName+"-"+
						oP.ClientUsername+"-"+
						strconv.FormatInt(int64(p), 10)+"-"+
						nature,
				)
				cmData[strconv.Itoa(int(p))] = s.Namespace + "/" +
					oP.MsgVpnName + "-" +
					oP.ClientUsername + "-" +
					strconv.Itoa(int(p)) + "-" +
					nature + ":" +
					strconv.Itoa(int(p))
				svcIds = append(
					svcIds,
					SvcId{
						MsgVpnName:     oP.MsgVpnName,
						ClientUsername: oP.ClientUsername,
						Port:           p,
						Nature:         nature,
					},
				)
			}
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
		if !StringInSlice(s.Name, *pubSubSvcNames) && s.Spec.Ports[0].Port != 8080 {
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
