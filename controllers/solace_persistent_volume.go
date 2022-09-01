package controllers

import (
	"context"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func NewPersistentVolume(s *scalablev1alpha1.SolaceScalable,
	instanceId string,
	labels map[string]string,
) *corev1.PersistentVolume {

	hostPathType := corev1.HostPathType("DirectoryOrCreate")
	prefix := "-" + instanceId
	spec := s.Spec.Container

	return &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:   s.Name + prefix,
			Labels: labels,
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceName(spec.Volume.Name): resource.MustParse(spec.Volume.Size),
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: spec.Volume.HostPath, Type: &hostPathType,
				},
			},
			AccessModes:                   []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimPolicy(spec.Volume.ReclaimPolicy),
		},
	}
}

func (r *SolaceScalableReconciler) CreateSolaceLocalPv(
	s *scalablev1alpha1.SolaceScalable,
	pv *corev1.PersistentVolume,
	ctx context.Context,
) (bool, error) {
	// create pvs if pvClass is localManual
	if s.Spec.PvClass == "localManual" {
		log := log.FromContext(ctx)
		foundpv := &corev1.PersistentVolume{}
		if err := r.Get(ctx, types.NamespacedName{Name: pv.Name, Namespace: pv.Namespace}, foundpv); err != nil {
			log.Info("Creating pv", pv.Namespace, pv.Name)
			if err = r.Create(ctx, pv); err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return true, nil
}
