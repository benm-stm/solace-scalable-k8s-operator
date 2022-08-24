package controllers

import (
	"context"
	"encoding/json"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func Statefulset(s *scalablev1alpha1.SolaceScalable, labels map[string]string) *v1.StatefulSet {
	name := s.Name
	storageClassName := s.Spec.PvClass
	if s.Spec.PvClass == "localManual" {
		storageClassName = ""
	}
	storageSize := "50Gi"

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
							Env: EnvVars(&s.Spec),
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
							Name: "storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "storage",
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
				WhenDeleted: "Retain",
				WhenScaled:  "Retain",
			},

			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:         "storage",
						GenerateName: name,
						Namespace:    s.Namespace,
						Labels:       labels,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"storage": resource.MustParse(storageSize),
							},
						},
						StorageClassName: &storageClassName,
					},
				},
			},
		},
		Status: v1.StatefulSetStatus{
			Replicas: s.Spec.Replicas,
		},
	}
}

// Check if the statefulset already exists
func (r *SolaceScalableReconciler) CreateStatefulSet(
	ss *v1.StatefulSet,
	ctx context.Context,
) (*v1.StatefulSet, error) {
	log := log.FromContext(ctx)
	foundSs := &v1.StatefulSet{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: ss.Name, Namespace: ss.Namespace}, foundSs); err != nil {
		log.Info("Creating Statefulset", ss.Namespace, ss.Name)
		if err = r.Create(context.TODO(), ss); err != nil {
			return nil, err
		}
	}
	return foundSs, nil
}

// Update the found object and write the result back if there are any changes
func (r *SolaceScalableReconciler) UpdateStatefulSet(
	ss *v1.StatefulSet,
	foundSs *v1.StatefulSet,
	ctx context.Context,
	hashStore *map[string]string,
) error {
	log := log.FromContext(ctx)
	newMarshal, _ := json.Marshal(foundSs.Spec)
	if len(*hashStore) == 0 {
		(*hashStore)[ss.Name] = AsSha256(newMarshal)
	} else if AsSha256(newMarshal) != (*hashStore)[ss.Name] {
		log.Info("Updating StatefulSet", ss.Namespace, ss.Name)
		if err := r.Update(ctx, ss); err != nil {
			return err
		}
		(*hashStore)[ss.Name] = AsSha256(newMarshal)
	}
	return nil
}
