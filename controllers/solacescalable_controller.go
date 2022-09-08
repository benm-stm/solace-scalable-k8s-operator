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

// SolaceScalableReconciler reconciles a SolaceScalable object
type SolaceScalableReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var hashStore = make(map[string]string)
var solaceAdminPassword string

//+kubebuilder:rbac:groups=scalable.solace.io,resources=solacescalables,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=scalable.solace.io,resources=solacescalables/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=scalable.solace.io,resources=solacescalables/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
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
	solaceLabels := Labels(solaceScalable)
	if err := r.Get(
		ctx,
		request.NamespacedName,
		solaceScalable,
	); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// check secret creation and store value
	if solaceAdminPassword == "" {
		foundSecret, err := r.GetSolaceSecret(solaceScalable, ctx)
		if err != nil {
			return reconcile.Result{}, err
		}
		solaceAdminPassword = GetSecretFromKey(
			solaceScalable,
			foundSecret,
			"username_admin_password",
			ctx,
		)
	}

	// Solace statefulset CRUD
	newSs := NewStatefulset(solaceScalable, solaceLabels)
	if err := controllerutil.SetControllerReference(
		solaceScalable,
		newSs,
		r.Scheme,
	); err != nil {
		return reconcile.Result{}, err
	}
	err := r.CreateStatefulSet(newSs, ctx)
	if err != nil {
		return reconcile.Result{}, err
	}
	if err := r.UpdateStatefulSet(newSs, ctx, &hashStore); err != nil {
		return reconcile.Result{}, err
	}

	for i := 0; i < int(solaceScalable.Spec.Replicas); i++ {
		//create solace instance http console service
		newSvc := NewSvcConsole(solaceScalable, i)
		if err := controllerutil.SetControllerReference(
			solaceScalable,
			newSvc,
			r.Scheme,
		); err != nil {
			return reconcile.Result{}, err
		}
		if err := r.CreateSolaceConsoleSvc(newSvc, ctx); err != nil {
			return reconcile.Result{}, err
		}

		// create solace instances PV
		newPv := NewPersistentVolume(
			solaceScalable,
			strconv.Itoa(i),
			solaceLabels,
		)
		if _, err := r.CreateSolaceLocalPv(
			solaceScalable,
			newPv,
			ctx,
		); err != nil {
			return reconcile.Result{}, err
		}
	}

	// Delete unused console services if they exist
	if err := r.DeleteSolaceConsoleSvc(
		solaceScalable,
		ctx,
	); err != nil {
		return reconcile.Result{}, err
	}

	// Create new ingress console services
	newIngConsole := NewIngressConsole(
		solaceScalable,
		solaceLabels,
	)
	if err := r.CreateSolaceConsoleIngress(
		solaceScalable,
		newIngConsole,
		ctx,
	); err != nil {
		return reconcile.Result{}, err
	}

	// Check if solace instances are up to query SempV2
	if _, success, _ := CallSolaceSempApi(
		solaceScalable,
		"/monitor/about/api",
		ctx,
		solaceAdminPassword,
	); success {
		// Get open svc pub/sub ports
		data, _, err := CallSolaceSempApi(
			solaceScalable,
			"/config/msgVpns?select="+
				"msgVpnName,enabled,*Port"+
				"&where=enabled==true",
			ctx,
			solaceAdminPassword,
		)
		if err != nil {
			return reconcile.Result{}, err
		}

		msgVpns, err := GetEnabledSolaceMsgVpns(
			solaceScalable,
			data,
		)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Get solace clientUsernames
		clientUsernames := SolaceClientUsernamesResp{}
		for _, m := range msgVpns.Data {
			data, _, err = CallSolaceSempApi(
				solaceScalable,
				"/config/msgVpns/"+m.MsgVpnName+
					"/clientUsernames?select="+
					"clientUsername,enabled,msgVpnName"+
					"&where=clientUsername!=*client-username",
				ctx,
				solaceAdminPassword,
			)
			if err != nil {
				return reconcile.Result{}, err
			}

			clientUsernamesTemp, err := msgVpns.GetSolaceClientUsernames(
				solaceScalable,
				data,
			)

			if err != nil {
				return reconcile.Result{}, err
			}

			clientUsernames.Data = append(clientUsernames.Data, clientUsernamesTemp.Data...)
		}

		// get client usernames attributes
		var clientUsernamesAttributes = ClientUsernameAttributes{}
		for _, v := range clientUsernames.Data {
			data, _, err = CallSolaceSempApi(
				solaceScalable,
				"/config/msgVpns/"+v.MsgVpnName+
					"/clientUsernames/"+v.ClientUsername+
					"/attributes",
				ctx,
				solaceAdminPassword,
			)
			if err != nil {
				return reconcile.Result{}, err
			}
			clientUsernameAttributes, err := GetClientUsernameAttributes(
				solaceScalable,
				data,
			)
			if err != nil {
				return reconcile.Result{}, err
			}
			clientUsernamesAttributes.Data = append(clientUsernamesAttributes.Data, clientUsernameAttributes.Data...)
		}

		pubSubsvcSpecs := clientUsernames.AddClientAttributes(clientUsernamesAttributes)
		for ks := range pubSubsvcSpecs {
			for _, m := range msgVpns.Data {
				(&pubSubsvcSpecs[ks]).AddMsgVpnPorts(m)
			}
		}

		// Contruct pub svcs
		pubPorts := []int32{}
		pubSvcNames, cmDataPub, pubSvcsId := ConstructSvcDatas(
			solaceScalable,
			&pubSubsvcSpecs,
			"pub",
			&pubPorts,
		)

		for _, svc := range pubSvcsId {
			newSvcPub := NewSvcPubSub(
				solaceScalable,
				svc,
				solaceLabels,
			)
			if err := r.CreatePubSubSvc(
				solaceScalable,
				newSvcPub,
				ctx,
			); err != nil {
				return reconcile.Result{}, err
			}
		}

		// Construct sub svc
		subPorts := []int32{}
		subSvcNames, cmDataSub, subSvcsId := ConstructSvcDatas(
			solaceScalable,
			&pubSubsvcSpecs,
			"sub",
			&subPorts,
		)

		//fmt.Printf("\nrobeau %v\n", subSvcsId)
		for _, svc := range subSvcsId {
			newSvcSub := NewSvcPubSub(
				solaceScalable,
				svc,
				solaceLabels,
			)
			if err := r.CreatePubSubSvc(
				solaceScalable,
				newSvcSub,
				ctx,
			); err != nil {
				return reconcile.Result{}, err
			}
		}

		// Check HAProxy pub services
		FoundHaproxyPubSvc, err := r.GetExistingHaProxySvc(
			solaceScalable,
			solaceScalable.Spec.Haproxy.Publish.ServiceName,
			ctx,
		)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Set the new pub data in the found haproxy svc
		FoundHaproxyPubSvc.Spec.Ports = *NewSvcHaproxy(
			solaceScalable,
			FoundHaproxyPubSvc.Spec.Ports,
			*cmDataPub,
		)

		// Update the existing pub haproxy
		if err := r.UpdateHAProxySvc(
			&hashStore,
			FoundHaproxyPubSvc,
			ctx,
		); err != nil {
			return reconcile.Result{}, err
		}

		// Check HAProxy sub services
		FoundHaproxySubSvc, err := r.GetExistingHaProxySvc(
			solaceScalable,
			solaceScalable.Spec.Haproxy.Subscribe.ServiceName,
			ctx,
		)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Set the new sub data in the found svc
		FoundHaproxySubSvc.Spec.Ports = *NewSvcHaproxy(
			solaceScalable,
			FoundHaproxySubSvc.Spec.Ports,
			*cmDataSub,
		)

		// Update the existing sub haproxy
		if err := r.UpdateHAProxySvc(
			&hashStore,
			FoundHaproxySubSvc,
			ctx,
		); err != nil {
			return reconcile.Result{}, err
		}

		// Create and update haproxy pub configmap
		configMapPub, err := r.CreateSolaceTcpConfigmap(
			solaceScalable,
			cmDataPub,
			"pub",
			ctx,
		)
		if err != nil {
			return reconcile.Result{}, err
		}

		if err := r.UpdateSolaceTcpConfigmap(
			solaceScalable,
			configMapPub,
			ctx,
			&hashStore,
		); err != nil {
			return reconcile.Result{}, err
		}

		// Create and update haproxy sub configmap
		configMapSub, err := r.CreateSolaceTcpConfigmap(
			solaceScalable,
			cmDataSub,
			"sub",
			ctx,
		)
		if err != nil {
			return reconcile.Result{}, err
		}
		if err := r.UpdateSolaceTcpConfigmap(
			solaceScalable,
			configMapSub,
			ctx,
			&hashStore,
		); err != nil {
			return reconcile.Result{}, err
		}

		// Get pub/sub services list
		svcList, err := r.ListPubSubSvc(solaceScalable, ctx)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Merge pub and sub services slices
		pubSubSvcNames := append(*pubSvcNames, *subSvcNames...)

		// Delete services not returned by cluster
		if err := r.DeletePubSubSvc(svcList, &pubSubSvcNames, ctx); err != nil {
			return reconcile.Result{}, err
		}

	} else {
		log.Error(err, "Solace API call issue")
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
