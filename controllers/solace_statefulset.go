package controllers

import (
	"context"
	"encoding/json"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	libs "github.com/benm-stm/solace-scalable-k8s-operator/common"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func NewStatefulset(
	s *scalablev1alpha1.SolaceScalable,
	labels map[string]string,
) *v1.StatefulSet {
	name := s.Name
	storageClassName := s.Spec.PvClass
	if s.Spec.PvClass == "localManual" {
		storageClassName = ""
	}

	return &v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:         name,
			GenerateName: name,
			Namespace:    s.Namespace,
		},
		Spec: v1.StatefulSetSpec{
			Replicas: &s.Spec.Replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:         name,
					GenerateName: name,
					Namespace:    s.Namespace,
					Labels:       labels,
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app",
												Operator: "In",
												Values:   []string{"solacescalable"},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  s.Spec.Container.Name,
							Image: s.Spec.Container.Image,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dshm",
									MountPath: "/dev/shm",
								},
								{
									Name:      "storage",
									MountPath: "/usr/sw/internalSpool/softAdb",
									SubPath:   "softAdb",
								},
								{
									Name:      "storage",
									MountPath: "/usr/sw/adb",
									SubPath:   "adb",
								},
								{
									Name:      "storage",
									MountPath: "/usr/sw/var",
									SubPath:   "var",
								},
								{
									Name:      "storage",
									MountPath: "/usr/sw/internalSpool",
									SubPath:   "internalSpool",
								},
								{
									Name:      "storage",
									MountPath: "/var/lib/solace/diagnostics",
									SubPath:   "diagnostics",
								},
								{
									Name:      "storage",
									MountPath: "/usr/sw/jail",
									SubPath:   "jail",
								},
							},
							Env: s.Spec.Container.Env,
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dshm",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: "Memory",
								},
							},
						},
						{
							Name: s.Spec.Container.Volume.Name,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: s.Spec.Container.Volume.Name,
								},
							},
						},
					},
				},
			},
			ServiceName:     name,
			UpdateStrategy:  v1.StatefulSetUpdateStrategy{},
			MinReadySeconds: 0,
			PersistentVolumeClaimRetentionPolicy: &v1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: v1.PersistentVolumeClaimRetentionPolicyType(
					s.Spec.Container.Volume.ReclaimPolicy,
				),
				WhenScaled: v1.PersistentVolumeClaimRetentionPolicyType(
					s.Spec.Container.Volume.ReclaimPolicy,
				),
			},

			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:         s.Spec.Container.Volume.Name,
						GenerateName: name,
						Namespace:    s.Namespace,
						Labels:       labels,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceName(s.Spec.Container.Volume.Name): resource.MustParse(
									s.Spec.Container.Volume.Size,
								),
							},
						},
						StorageClassName: &storageClassName,
					},
				},
			},
		},
	}
}

// Check if the statefulset already exists
func (r *SolaceScalableReconciler) CreateStatefulSet(
	ss *v1.StatefulSet,
	ctx context.Context,
) error {
	log := log.FromContext(ctx)
	if err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      ss.Name,
			Namespace: ss.Namespace,
		},
		&v1.StatefulSet{},
	); err != nil {
		log.Info("Creating Statefulset", ss.Namespace, ss.Name)
		if err = r.Create(ctx, ss); err != nil {
			return err
		}
	}
	return nil
}

// Update the found object and write the result back if there are any changes
func (r *SolaceScalableReconciler) UpdateStatefulSet(
	ss *v1.StatefulSet,
	ctx context.Context,
	hashStore *map[string]string,
) error {
	log := log.FromContext(ctx)
	newMarshal, _ := json.Marshal(ss.Spec)
	if (*hashStore)[ss.Name] == "" ||
		libs.AsSha256(newMarshal) != (*hashStore)[ss.Name] {
		log.Info("Updating StatefulSet", ss.Namespace, ss.Name)
		if err := r.Update(ctx, ss); err != nil {
			return err
		}
		(*hashStore)[ss.Name] = libs.AsSha256(newMarshal)
	}
	return nil
}
