package controllers

import (
	"context"
	"encoding/json"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func NewtcpConfigmap(
	s *scalablev1alpha1.SolaceScalable,
	data map[string]string,
	nature string,
	labels map[string]string,
) *v1.ConfigMap {

	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name + "-" + nature + "-tcp-ingress",
			Namespace: s.Namespace,
			Labels:    labels,
		},

		Data: data,
	}
}

//create tcp ingress configmap
func (r *SolaceScalableReconciler) CreateSolaceTcpConfigmap(
	data *map[string]string,
	nature string,
	solaceScalable *scalablev1alpha1.SolaceScalable,
	ctx context.Context,
) (*v1.ConfigMap, *v1.ConfigMap, error) {
	log := log.FromContext(ctx)
	configMap := NewtcpConfigmap(solaceScalable, *data, nature, Labels(solaceScalable))

	FoundHaproxyConfigMap := &corev1.ConfigMap{}
	if err := r.Get(context.TODO(),
		types.NamespacedName{
			Name:      configMap.Name,
			Namespace: configMap.Namespace,
		}, FoundHaproxyConfigMap,
	); err != nil {
		log.Info("Creating HAProxy Ingress ConfigMap", configMap.Namespace, configMap.Name)
		err = r.Create(context.TODO(), configMap)
		return nil, nil, err
	}
	return configMap, FoundHaproxyConfigMap, nil
}

//update tcp ingress configmap
func (r *SolaceScalableReconciler) UpdateSolaceTcpConfigmap(
	f *v1.ConfigMap,
	configMap *v1.ConfigMap,
	solaceScalable *scalablev1alpha1.SolaceScalable,
	ctx context.Context,
	hashStore *map[string]string,
) error {
	log := log.FromContext(ctx)
	newMarshal, _ := json.Marshal(f.Data)
	datasMarshal, _ := json.Marshal(configMap.Data)

	if len(*hashStore) == 0 {
		(*hashStore)[f.Name] = AsSha256(newMarshal)
	} else if AsSha256(datasMarshal) != (*hashStore)[f.Name] {
		log.Info("Updating HAProxy Ingress ConfigMap", configMap.Namespace, configMap.Name)
		f.Data = configMap.Data
		(*hashStore)[f.Name] = AsSha256(datasMarshal)
		if err := r.Update(context.TODO(), f); err != nil {
			return err
		}
	}
	return nil
}

// update default haproxy configmap
func (r *SolaceScalableReconciler) UpdateDefaultHaproxyConfigmap(
	FoundHaproxyConfigMap *v1.ConfigMap,
	configMap *v1.ConfigMap,
	solaceScalable *scalablev1alpha1.SolaceScalable,
	ctx context.Context,
	hashStore *map[string]string,
) error {
	log := log.FromContext(ctx)
	newMarshal, _ := json.Marshal(FoundHaproxyConfigMap.Data)
	datasMarshal, _ := json.Marshal(configMap.Data)

	if len(*hashStore) == 0 {
		(*hashStore)[FoundHaproxyConfigMap.Name] = AsSha256(newMarshal)
	} else if AsSha256(datasMarshal) != (*hashStore)[FoundHaproxyConfigMap.Name] {
		log.Info("Updating HAProxy default ConfigMap", configMap.Namespace, configMap.Name)
		FoundHaproxyConfigMap.Data = configMap.Data
		(*hashStore)[FoundHaproxyConfigMap.Name] = AsSha256(datasMarshal)
		if err := r.Update(context.TODO(), FoundHaproxyConfigMap); err != nil {
			return err
		}
	}
	return nil
}
