package service

import (
	"context"
	"strconv"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func NewConsole(
	s *scalablev1alpha1.SolaceScalable,
	counter int,
) *corev1.Service {
	name := s.Name + "-" +
		strconv.Itoa(counter)
	selector := map[string]string{
		"statefulset.kubernetes.io/pod-name": name,
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: s.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports: []corev1.ServicePort{{
				Protocol: corev1.ProtocolTCP,
				Port:     8080,
			}},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// create console service
func CreateConsole(
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
		log.Info("Creating Solace Console Svc", svc.Namespace, svc.Name)
		if err = k.Create(ctx, svc); err != nil {
			return err
		}
	}
	return nil
}

// delete unused console services
func DeleteConsole(
	solaceScalable *scalablev1alpha1.SolaceScalable,
	k k8sClient,
	ctx context.Context,
) error {
	log := log.FromContext(ctx)
	counter := int(solaceScalable.Spec.Replicas)
	nbSvcToCheck := 5 + counter
	// loop indefinitely until not finding 5 existing console service
	for {
		svc := NewConsole(solaceScalable, counter)
		foundExtraSvc := &corev1.Service{}
		if err := k.Get(
			ctx,
			types.NamespacedName{
				Name:      svc.Name,
				Namespace: svc.Namespace,
			},
			foundExtraSvc,
		); err != nil {
			counter++
		} else {
			log.Info("Delete Solace Console Service", svc.Namespace, svc.Name)
			if err = k.Delete(ctx, foundExtraSvc); err != nil {
				return err
			}
		}
		if counter == nbSvcToCheck {
			break
		}
	}
	return nil
}
