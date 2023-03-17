package service

import (
	"context"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	libs "github.com/benm-stm/solace-scalable-k8s-operator/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func New(
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

// create console service
func Create(
	svc *corev1.Service,
	k k8sClient,
	ctx context.Context,
) error {
	// Check if the console svc already exists
	log := log.FromContext(ctx)
	foundSvc := &corev1.Service{}
	if err := k.Get(
		ctx,
		types.NamespacedName{
			Name:      svc.Name,
			Namespace: svc.Namespace,
		},
		foundSvc,
	); err != nil {
		log.Info("Create Svc", svc.Namespace, svc.Name)
		if err = k.Create(ctx, svc); err != nil {
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
	svcNames *[]string,
	exceptionPort *int32,
	k k8sClient,
	ctx context.Context,
) error {
	log := log.FromContext(ctx)
	for _, s := range svcList.Items {
		if !libs.IsItInSlice(s.Name, *svcNames) && s.Spec.Ports[0].Port != *exceptionPort {
			foundExtraPubSubSvc := &corev1.Service{}
			if err := k.Get(
				ctx,
				types.NamespacedName{
					Namespace: s.Namespace,
					Name:      s.Name,
				}, foundExtraPubSubSvc,
			); err == nil {
				log.Info("Delete Svc", s.Namespace, s.Name)
				if err = k.Delete(ctx, foundExtraPubSubSvc); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
