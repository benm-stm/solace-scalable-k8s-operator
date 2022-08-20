package controllers

import (
	"context"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	scalablev1alpha1 "solace.io/api/v1alpha1"
)

func PersistentVolume(s *scalablev1alpha1.SolaceScalable, replicaNbr string) *corev1.PersistentVolume {
	//pvcName := pv.Name + "volumeclame-node" + replicaNbr
	//FSType := "xfs"
	hostPathType := corev1.HostPathType("DirectoryOrCreate")
	prefix := "-" + replicaNbr

	return &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.Name + prefix,
			//Namespace: s.Namespace,
			//Namespace: "",
			Labels: labels(s),
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceName("storage"): resource.MustParse("50Gi"),
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{Path: "/opt/storage", Type: &hostPathType},
				//Local: &corev1.LocalVolumeSource{Path: "/opt/storage", FSType: &FSType},
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			/*ClaimRef: &corev1.ObjectReference{
				Namespace: s.Namespace,
			},*/
			PersistentVolumeReclaimPolicy: "Retain",
			//StorageClassName:              "",
			//MountOptions:                  []string{"hard", "nfsvers=4.1"},
		},
	}
}

func createSolaceLocalPv(spec *scalablev1alpha1.SolaceScalable, instanceId int, r *SolaceScalableReconciler, ctx context.Context) error {
	// create pvs if pvClass is localManual
	if spec.Spec.PvClass == "localManual" {
		log := log.FromContext(ctx)
		pv := PersistentVolume(spec, strconv.Itoa(instanceId))
		foundpv := &corev1.PersistentVolume{}
		if err := r.Get(context.TODO(), types.NamespacedName{Name: pv.Name, Namespace: pv.Namespace}, foundpv); err != nil {
			log.Info("Creating pv", pv.Namespace, pv.Name)
			if err = r.Create(context.TODO(), pv); err != nil {
				return err
			}
		}
	}
	return nil
}
