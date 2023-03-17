package secret

import (
	"context"
	"errors"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type k8sClient interface {
	Get(
		ctx context.Context,
		key types.NamespacedName,
		obj client.Object,
		opts ...client.GetOption,
	) error
}

// Check if the secret already exists
func Get(
	s *scalablev1alpha1.SolaceScalable,
	k k8sClient,
	ctx context.Context,
) (*corev1.Secret, error) {
	foundS := &corev1.Secret{}
	if err := k.Get(
		ctx,
		types.NamespacedName{
			Name:      s.Name,
			Namespace: s.Namespace,
		},
		foundS); err != nil {
		return nil, err
	}
	return foundS, nil
}

func GetSecretFromKey(s *scalablev1alpha1.SolaceScalable,
	secret *corev1.Secret,
	secretName string,
) (string, error) {
	i, err := GetSecretEnvIndex(s, secretName)
	if err != nil {
		return "", err
	}
	key := s.Spec.Container.
		Env[i].
		ValueFrom.SecretKeyRef.Key
	return string(secret.Data[key]), nil
}

// Gets the index of the secret in the env array
func GetSecretEnvIndex(s *scalablev1alpha1.SolaceScalable,
	secretName string,
) (int, error) {
	for i, v := range s.Spec.Container.Env {
		if v.Name == secretName {
			return i, nil
		}
	}
	err := errors.New("secret not found")
	return -1, err
}
