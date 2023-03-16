package ingress

import (
	"context"

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
	Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error
	Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error
}
