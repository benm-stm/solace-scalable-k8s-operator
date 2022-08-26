package controllers

import (
	"context"
	"encoding/json"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func NewtcpConfigmap(
	s *scalablev1alpha1.SolaceScalable,
	data map[string]string,
	nature string,
	labels map[string]string,
) *corev1.ConfigMap {

	return &corev1.ConfigMap{
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
) (*corev1.ConfigMap, *corev1.ConfigMap, error) {
	log := log.FromContext(ctx)
	configMap := NewtcpConfigmap(solaceScalable, *data, nature, Labels(solaceScalable))

	FoundHaproxyConfigMap := &corev1.ConfigMap{}
	if err := r.Get(ctx,
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
	c *corev1.ConfigMap,
	configMap *corev1.ConfigMap,
	solaceScalable *scalablev1alpha1.SolaceScalable,
	ctx context.Context,
	hashStore *map[string]string,
) error {
	log := log.FromContext(ctx)
	newMarshal, _ := json.Marshal(c.Data)
	datasMarshal, _ := json.Marshal(configMap.Data)

	if len(*hashStore) == 0 {
		(*hashStore)[c.Name] = AsSha256(newMarshal)
	} else if AsSha256(datasMarshal) != (*hashStore)[c.Name] {
		log.Info("Updating HAProxy Ingress ConfigMap", configMap.Namespace, configMap.Name)
		c.Data = configMap.Data
		(*hashStore)[c.Name] = AsSha256(datasMarshal)
		if err := r.Update(context.TODO(), c); err != nil {
			return err
		}
	}
	return nil
}
