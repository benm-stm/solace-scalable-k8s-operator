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
	"time"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	scalablev1alpha1 "solace.io/api/v1alpha1"
)

// SolaceScalableReconciler reconciles a SolaceScalable object
type SolaceScalableReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//var blacklistedClientUsernames = [1]string{"#client-username"}

var hashStore = make(map[string]string)

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

func (r *SolaceScalableReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	// TODO: Check existance of CRD
	//log := log.FromContext(ctx)
	solaceScalable := &scalablev1alpha1.SolaceScalable{}
	if err := r.Get(context.TODO(), request.NamespacedName, solaceScalable); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// TODO: Solace statefulset creation
	ss := Statefulset(solaceScalable)
	if err := controllerutil.SetControllerReference(solaceScalable, ss, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}
	foundSs, err := CreateStatefulSet(ss, r, ctx)
	if err != nil {
		return reconcile.Result{}, err
	}
	if err := UpdateStatefulSet(ss, foundSs, r, ctx, &hashStore); err != nil {
		return reconcile.Result{}, err
	}

	for i := 0; i < int(solaceScalable.Spec.Replicas); i++ {
		//create solace instance http console service
		svc := SvcConsole(solaceScalable, i)
		if err := controllerutil.SetControllerReference(solaceScalable, svc, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		if err := CreateSolaceConsoleSvc(svc, r, ctx); err != nil {
			return reconcile.Result{}, err
		}

		// create solace instances PV
		if err := createSolaceLocalPv(solaceScalable, i, r, ctx); err != nil {
			return reconcile.Result{}, err
		}
	}

	if err := DeleteSolaceConsoleSvc(solaceScalable, r, ctx); err != nil {
		return reconcile.Result{}, err
	}

	if err := CreateSolaceConsoleIngress(solaceScalable, r, ctx); err != nil {
		return reconcile.Result{}, err
	}

	// Open svc pub/sub ports
	enabledMsgVpns := getEnabledSolaceMsgVpns(solaceScalable)
	pubSubOpenPorts := mergeSolaceResponses(enabledMsgVpns, getSolaceClientUsernames(solaceScalable, enabledMsgVpns)).Data
	pubSubSvcNames, data, err := CreatePubSubSvc(solaceScalable, &pubSubOpenPorts, &enabledMsgVpns, r, ctx)
	if err != nil {
		return reconcile.Result{}, err
	}

	// check HAProxy service
	FoundHaproxySvc, err := GetExistingHaProxySvc(solaceScalable, r, ctx)
	if err != nil {
		return reconcile.Result{}, err
	}
	//set the new data in the found svc
	FoundHaproxySvc.Spec.Ports = *SvcHaproxy(solaceScalable, FoundHaproxySvc.Spec.Ports, *data)

	if err := UpdateHAProxySvc(&hashStore, FoundHaproxySvc, r, ctx); err != nil {
		return reconcile.Result{}, err
	}

	// create and update haproxy configmap
	configMap, FoundHaproxyConfigMap, err := CreateTcpIngressConfigmap(data, solaceScalable, r, ctx)
	if err != nil {
		return reconcile.Result{}, err
	}
	if err := UpdateTcpIngressConfigmap(FoundHaproxyConfigMap, configMap, solaceScalable, r, ctx); err != nil {
		return reconcile.Result{}, err
	}

	svcList, foundExtraPubSubSvc, err := ListPubSubSvc(solaceScalable, r)
	if err != nil {
		return reconcile.Result{}, err
	}
	if err := DeletePubSubSvc(svcList, foundExtraPubSubSvc, pubSubSvcNames, r, ctx); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{RequeueAfter: time.Second * 10}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SolaceScalableReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scalablev1alpha1.SolaceScalable{}).
		Owns(&v1.StatefulSet{}).
		Owns(&corev1.Service{}).
		//WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Complete(r)
}
