package controllers

import (
	"context"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Check if the secret already exists
func (r *SolaceScalableReconciler) GetSolaceSecret(
	s *scalablev1alpha1.SolaceScalable,
	ctx context.Context,
) (*corev1.Secret, error) {
	log := log.FromContext(ctx)
	foundS := &corev1.Secret{}
	if err := r.Get(
		context.TODO(),
		types.NamespacedName{
			Name:      s.Name,
			Namespace: s.Namespace,
		},
		foundS); err != nil {
		log.Error(err, "Declared solace secret does not exist please create it!")
		return nil, err
	}
	return foundS, nil
}

func GetSecretFromKey(s *scalablev1alpha1.SolaceScalable,
	secret *corev1.Secret,
	secretName string,
	ctx context.Context,
) string {
	key := s.Spec.Container.
		Env[GetSecretEnvIndex(s, secretName, ctx)].
		ValueFrom.SecretKeyRef.Key
	return string(secret.Data[key])
}

func GetSecretEnvIndex(s *scalablev1alpha1.SolaceScalable,
	secretName string,
	ctx context.Context,
) int {
	log := log.FromContext(ctx)
	for i, v := range s.Spec.Container.Env {
		if v.Name == secretName {
			return i
		}
	}
	log.Info("Secret 'username_admin_password' not declared in container Env!")
	return -1
}
