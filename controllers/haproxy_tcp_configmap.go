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

/*type TcpConfigmap struct {
	Port string
	Svc  string
}*/
func NewTcpConfigmap(
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
	solaceScalable *scalablev1alpha1.SolaceScalable,
	data *map[string]string,
	nature string,
	ctx context.Context,
) (*corev1.ConfigMap, error) {
	log := log.FromContext(ctx)
	configMap := NewTcpConfigmap(solaceScalable, *data, nature, Labels(solaceScalable))

	// check if the configmap exists
	if err := r.Get(ctx,
		types.NamespacedName{
			Name:      configMap.Name,
			Namespace: configMap.Namespace,
		}, &corev1.ConfigMap{},
	); err != nil {
		log.Info("Creating HAProxy Ingress ConfigMap", configMap.Namespace, configMap.Name)
		err = r.Create(ctx, configMap)
		return nil, err
	}
	return configMap, nil
}

//update tcp ingress configmap
func (r *SolaceScalableReconciler) UpdateSolaceTcpConfigmap(
	solaceScalable *scalablev1alpha1.SolaceScalable,
	configMap *corev1.ConfigMap,
	ctx context.Context,
	hashStore *map[string]string,
) error {
	// when i delete the configmap, a nil pointer will trig
	if configMap != nil {
		log := log.FromContext(ctx)
		datasMarshal, _ := json.Marshal(configMap.Data)

		/*if (*hashStore)[configMap.Name] == "" {
			(*hashStore)[configMap.Name] = AsSha256(datasMarshal)
		} else */
		if (*hashStore)[configMap.Name] == "" ||
			AsSha256(datasMarshal) != (*hashStore)[configMap.Name] {
			log.Info("Updating HAProxy Ingress ConfigMap", configMap.Namespace, configMap.Name)
			if err := r.Update(ctx, configMap); err != nil {
				return err
			}
			//update hash to not trig update if conf has not changed
			(*hashStore)[configMap.Name] = AsSha256(datasMarshal)
		}
	}
	return nil
}
