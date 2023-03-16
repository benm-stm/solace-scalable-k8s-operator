package service

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
	Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error
	Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error
	List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
}

type SvcId struct {
	Name           string
	ClientUsername string
	MsgVpnName     string
	Port           int32
	TargetPort     int
	Nature         string
}

type SvcData struct {
	SvcNames []string
	CmData   map[string]string
	SvcsId   []SvcId
}
