/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"strconv"
	"time"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	libs "github.com/benm-stm/solace-scalable-k8s-operator/common"
	"github.com/benm-stm/solace-scalable-k8s-operator/configmap"
	"github.com/benm-stm/solace-scalable-k8s-operator/handler/solace"
	"github.com/benm-stm/solace-scalable-k8s-operator/ingress"
	pv "github.com/benm-stm/solace-scalable-k8s-operator/persistentvolume"
	"github.com/benm-stm/solace-scalable-k8s-operator/secret"
	"github.com/benm-stm/solace-scalable-k8s-operator/service"
	specs "github.com/benm-stm/solace-scalable-k8s-operator/service/specs"
	"github.com/benm-stm/solace-scalable-k8s-operator/statefulset"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const adminPwdEnvName = "username_admin_password"

var consolePort int32 = 8080

// SolaceScalableReconciler reconciles a SolaceScalable object
type SolaceScalableReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var hashStore = make(map[string]string)
var solaceAdminPassword string

var singleton bool = false

//+kubebuilder:rbac:groups=scalable.solace.io,resources=solacescalables,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=scalable.solace.io,resources=solacescalables/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=scalable.solace.io,resources=solacescalables/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;delete;update;patch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;delete;update;patch
//+kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;list;watch;create
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Modify the Reconcile function to compare the state specified by
// the SolaceScalable object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile

func (r *SolaceScalableReconciler) Reconcile(
	ctx context.Context,
	request ctrl.Request,
) (ctrl.Result, error) {
	// Check existance of CRD
	log := log.FromContext(ctx)
	solaceScalable := &scalablev1alpha1.SolaceScalable{}
	solaceLabels := libs.Labels(solaceScalable)
	if err := r.Get(
		ctx,
		request.NamespacedName,
		solaceScalable,
	); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// check secret creation and store value
	if solaceAdminPassword == "" {
		foundSecret, err := secret.Get(solaceScalable, r, ctx)
		if err != nil {
			return reconcile.Result{}, err
		}
		solaceAdminPassword, err = secret.GetSecretFromKey(
			solaceScalable,
			foundSecret,
			adminPwdEnvName,
		)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Solace statefulset CRUD
	newSs := statefulset.New(solaceScalable, solaceLabels)
	if err := controllerutil.SetControllerReference(
		solaceScalable,
		newSs,
		r.Scheme,
	); err != nil {
		return reconcile.Result{}, err
	}
	err := statefulset.Create(newSs, r, ctx)
	if err != nil {
		return reconcile.Result{}, err
	}
	if err := statefulset.Update(newSs, r, ctx, &hashStore); err != nil {
		return reconcile.Result{}, err
	}

	for i := 0; i < int(solaceScalable.Spec.Replicas); i++ {
		//create solace http console service
		newSvc := service.NewConsole(solaceScalable, i)
		if err := controllerutil.SetControllerReference(
			solaceScalable,
			newSvc,
			r.Scheme,
		); err != nil {
			return reconcile.Result{}, err
		}

		if err := service.Create(newSvc, r, ctx); err != nil {
			return reconcile.Result{}, err
		}

		// create solace localPV if storageclass is localmanual
		newPv := pv.New(
			solaceScalable,
			strconv.Itoa(i),
			solaceLabels,
		)
		if _, err := pv.Create(
			solaceScalable,
			newPv,
			r,
			ctx,
		); err != nil {
			return reconcile.Result{}, err
		}
	}

	// Create new ingress console services
	newIngConsole := ingress.NewConsole(
		solaceScalable,
		solaceLabels,
	)
	if err := ingress.CreateConsole(
		solaceScalable,
		newIngConsole,
		r,
		ctx,
	); err != nil {
		return reconcile.Result{}, err
	}

	// Check if solace instances are up to query SempV2
	for i := 0; i < int((*solaceScalable).Spec.Replicas); i++ {
		// show solace info only once
		if !singleton {
			aboutApi := solace.NewAboutApi()
			err := aboutApi.GetInfos(
				i,
				solaceScalable,
				ctx,
				solaceAdminPassword,
			)
			if err != nil {
				return reconcile.Result{}, err
			}
			log.Info("Solace Api Informations",
				aboutApi.GetPlatform(),
				aboutApi.GetSempVersion(),
			)
			singleton = true
		}

		msgVpns := solace.NewMsgVpns()
		if err := msgVpns.GetEnabledMsgVpns(
			i,
			solaceScalable,
			ctx,
			solaceAdminPassword,
		); err != nil {
			return reconcile.Result{}, err
		}

		// Get solace clientUsernames
		cu := solace.NewClientUsernames()
		for _, m := range msgVpns.Data {
			if err := cu.Add(
				solaceScalable,
				i,
				&m,
				solaceAdminPassword,
				ctx,
			); err != nil {
				return reconcile.Result{}, err
			}
		}

		// get client usernames attributes
		var attr = solace.NewClientUsernameAttrs()
		for _, v := range cu.Data {
			if err := attr.Add(
				solaceScalable,
				i,
				&v,
				solaceAdminPassword,
				ctx,
			); err != nil {
				return reconcile.Result{}, err
			}
		}

		// build pub and sub services spec
		svcsSpec := specs.NewSvcsSpec()
		if err := svcsSpec.WithClientAttributes(
			attr,
			cu,
		); err != nil {
			return reconcile.Result{}, err
		}
		// Merge message vpn ports in solace spec
		for _, ss := range svcsSpec.Data {
			for _, m := range msgVpns.Data {
				ss.WithMsgVpnPorts(&m)
			}
		}

		// Create pub svcs
		pubSvcData := service.NewSvcData()
		pubSvcData.Set(
			solaceScalable,
			&svcsSpec.Data,
			"pub",
		)

		for _, svc := range pubSvcData.SvcsId {
			newSvcPub := service.New(
				solaceScalable,
				svc,
				solaceLabels,
			)
			if err := service.Create(
				//solaceScalable,
				newSvcPub,
				r,
				ctx,
			); err != nil {
				return reconcile.Result{}, err
			}
		}

		// Create sub svcs
		subSvcData := service.NewSvcData()
		subSvcData.Set(
			solaceScalable,
			&svcsSpec.Data,
			"sub",
		)

		for _, svc := range subSvcData.SvcsId {
			newSvcSub := service.New(
				solaceScalable,
				svc,
				solaceLabels,
			)
			if err := service.Create(
				//solaceScalable,
				newSvcSub,
				r,
				ctx,
			); err != nil {
				return reconcile.Result{}, err
			}
		}

		// Check HAProxy pub services
		FoundHaproxyPubSvc, err := ingress.GetTcp(
			solaceScalable,
			solaceScalable.Spec.Haproxy.Publish.ServiceName,
			r,
			ctx,
		)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Set the new pub data in the found haproxy svc
		FoundHaproxyPubSvc.Spec.Ports = *ingress.NewTcp(
			solaceScalable,
			FoundHaproxyPubSvc.Spec.Ports,
			pubSvcData.CmData,
		)

		// Update the existing pub haproxy
		if err := ingress.UpdateTcp(
			&hashStore,
			FoundHaproxyPubSvc,
			r,
			ctx,
		); err != nil {
			return reconcile.Result{}, err
		}

		// Check HAProxy sub services
		FoundHaproxySubSvc, err := ingress.GetTcp(
			solaceScalable,
			solaceScalable.Spec.Haproxy.Subscribe.ServiceName,
			r,
			ctx,
		)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Set the new sub data in the found svc
		FoundHaproxySubSvc.Spec.Ports = *ingress.NewTcp(
			solaceScalable,
			FoundHaproxySubSvc.Spec.Ports,
			//*cmDataSub,
			subSvcData.CmData,
		)

		// Update the existing sub haproxy
		if err := ingress.UpdateTcp(
			&hashStore,
			FoundHaproxySubSvc,
			r,
			ctx,
		); err != nil {
			return reconcile.Result{}, err
		}

		// Create and update haproxy pub configmap
		configMapPub, err := configmap.Create(
			solaceScalable,
			&pubSvcData.CmData,
			"pub",
			r,
			ctx,
		)
		if err != nil {
			return reconcile.Result{}, err
		}

		if err := configmap.Update(
			solaceScalable,
			configMapPub,
			r,
			ctx,
			&hashStore,
		); err != nil {
			return reconcile.Result{}, err
		}

		// Create and update haproxy sub configmap
		configMapSub, err := configmap.Create(
			solaceScalable,
			&subSvcData.CmData,
			"sub",
			r,
			ctx,
		)
		if err != nil {
			return reconcile.Result{}, err
		}
		if err := configmap.Update(
			solaceScalable,
			configMapSub,
			r,
			ctx,
			&hashStore,
		); err != nil {
			return reconcile.Result{}, err
		}

		// Get pub/sub services list
		svcList, err := service.List(solaceScalable, r, ctx)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Merge pub and sub services slices
		pubSubSvcNames := append(pubSvcData.SvcNames, subSvcData.SvcNames...)

		// Delete unused console services if they exist
		if err := service.DeleteConsole(
			solaceScalable,
			r,
			ctx,
		); err != nil {
			return reconcile.Result{}, err
		}

		// Delete services not returned by cluster
		if err := service.Delete(svcList, &pubSubSvcNames, &consolePort, r, ctx); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{RequeueAfter: time.Second * 10}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SolaceScalableReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scalablev1alpha1.SolaceScalable{}).
		Owns(&v1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolume{}).
		Complete(r)
}
