package controllers

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	scalablev1alpha1 "solace.io/api/v1alpha1"
)

func tcpConfigMap(s *scalablev1alpha1.SolaceScalable, data map[string]string) *v1.ConfigMap {
	labels := labels(s)
	//fmt.Printf("%v", data)

	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name + "-tcp-ingress",
			Namespace: s.Namespace,
			Labels:    labels,
		},

		Data: data,
	}
}

// haproxy
/*func tcpConfigMap(s *scalablev1alpha1.SolaceScalable, data map[string]string) *v1.ConfigMap {
	labels := labels(s)

	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name + "-tcp-ingress",
			Namespace: s.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"balance-algorithm": "leastconn",
			"max-connections":   "10000",
			"ssl-redirect":      "true",
		},
	}
}*/

//create tcp ingress configmap
func CreateTcpIngressConfigmap(data *map[string]string, solaceScalable *scalablev1alpha1.SolaceScalable, r *SolaceScalableReconciler, ctx context.Context) (*v1.ConfigMap, *v1.ConfigMap, error) {
	log := log.FromContext(ctx)
	(*data)["balance-algorithm"] = "leastconn"
	configMap := tcpConfigMap(solaceScalable, *data)

	FoundHaproxyConfigMap := &corev1.ConfigMap{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, FoundHaproxyConfigMap); err != nil && errors.IsNotFound(err) {
		log.Info("Creating HAProxy Ingress ConfigMap", configMap.Namespace, configMap.Name)
		err = r.Create(context.TODO(), configMap)
		return nil, nil, err
	}
	return configMap, FoundHaproxyConfigMap, nil
}

func UpdateTcpIngressConfigmap(FoundHaproxyConfigMap *v1.ConfigMap, configMap *v1.ConfigMap, solaceScalable *scalablev1alpha1.SolaceScalable, r *SolaceScalableReconciler, ctx context.Context) error {
	log := log.FromContext(ctx)
	newMarshal, _ := json.Marshal(FoundHaproxyConfigMap.Data)
	datasMarshal, _ := json.Marshal(configMap.Data)

	if len(hashStore) == 0 {
		hashStore[FoundHaproxyConfigMap.Name] = asSha256(newMarshal)
	} else if asSha256(datasMarshal) != hashStore[FoundHaproxyConfigMap.Name] {
		log.Info("Updating HAProxy Ingress ConfigMap", configMap.Namespace, configMap.Name)
		FoundHaproxyConfigMap.Data = configMap.Data
		hashStore[FoundHaproxyConfigMap.Name] = asSha256(datasMarshal)
		if err := r.Update(context.TODO(), FoundHaproxyConfigMap); err != nil && errors.IsNotFound(err) {
			return err
		}
	}
	return nil
}
