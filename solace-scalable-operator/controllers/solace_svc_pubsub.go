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

func SvcPubSub(s *scalablev1alpha1.SolaceScalable,
	m SolaceMergedResp,
	p int32,
	pubSub string,
	labels map[string]string,
) *corev1.Service {

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: m.MsgVpnName + "-" +
				m.ClientUsername + "-" +
				strconv.FormatInt(int64(p), 10) + "-" +
				pubSub,
			Namespace: s.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol: corev1.ProtocolTCP,
				Port:     p,
			}},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// pubsub SVC creation
func (r *SolaceScalableReconciler) CreatePubSubSvc(s *scalablev1alpha1.SolaceScalable,
	pubSubOpenPorts *[]SolaceMergedResp,
	enabledMsgVpns *SolaceMsgVpnsResp,
	nature string,
	ctx context.Context,
) (*[]string, *map[string]string, error) {
	log := log.FromContext(ctx)
	var pubSubSvcNames = []string{}
	var data = map[string]string{}
	for _, i := range *pubSubOpenPorts {
		for _, j := range i.Ports {
			if j != 0 {
				//create pubsub SVC
				pubSubSvcNames = append(
					pubSubSvcNames,
					i.MsgVpnName+"-"+
						i.ClientUsername+"-"+
						strconv.FormatInt(int64(j), 10)+"-"+
						nature,
				)
				pbs := SvcPubSub(s, i, j, nature, Labels(s))
				foundPubSubSvc := &corev1.Service{}
				if err := r.Get(
					context.TODO(),
					types.NamespacedName{
						Name:      pbs.Name,
						Namespace: pbs.Namespace,
					},
					foundPubSubSvc,
				); err != nil {
					log.Info("Creating pubSub SVC", pbs.Namespace, pbs.Name)
					if err = r.Create(context.TODO(), pbs); err != nil {
						return nil, nil, err
					}
				}
				data[strconv.Itoa(int(j))] = s.Namespace + "/" +
					i.MsgVpnName + "-" +
					i.ClientUsername + "-" +
					strconv.Itoa(int(j)) + "-" +
					nature + ":" +
					strconv.Itoa(int(j))
			}
		}
	}
	return &pubSubSvcNames, &data, nil
}

func ListPubSubSvc(solaceScalable *scalablev1alpha1.SolaceScalable, r *SolaceScalableReconciler) (*corev1.ServiceList, *corev1.Service, error) {
	// get existing svc list
	foundExtraPubSubSvc := &corev1.Service{}
	svcList := &corev1.ServiceList{}
	listOptions := &client.ListOptions{Namespace: solaceScalable.Namespace}

	if err := r.List(context.TODO(), svcList, listOptions); err != nil {
		return nil, nil, err
	}
	return svcList, foundExtraPubSubSvc, nil
}

func DeletePubSubSvc(svcList *corev1.ServiceList, foundExtraPubSubSvc *corev1.Service, pubSubSvcNames *[]string, r *SolaceScalableReconciler, ctx context.Context) error {
	log := log.FromContext(ctx)
	for _, s := range svcList.Items {
		if !StringInSlice(s.Name, *pubSubSvcNames) && s.Spec.Ports[0].Port != 8080 {
			if err := r.Get(context.TODO(), types.NamespacedName{Namespace: s.Namespace, Name: s.Name}, foundExtraPubSubSvc); err != nil {
				break
			} else {
				log.Info("Delete PubSubSvc", s.Namespace, s.Name)
				if err = r.Delete(context.TODO(), foundExtraPubSubSvc); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
