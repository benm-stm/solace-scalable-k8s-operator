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
	log := log.FromContext(ctx)
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
	ss := Statefulset(solaceScalable, Labels(solaceScalable))
	if err := controllerutil.SetControllerReference(solaceScalable, ss, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}
	foundSs, err := r.CreateStatefulSet(ss, ctx)
	if err != nil {
		return reconcile.Result{}, err
	}
	if err := r.UpdateStatefulSet(ss, foundSs, ctx, &hashStore); err != nil {
		return reconcile.Result{}, err
	}

	for i := 0; i < int(solaceScalable.Spec.Replicas); i++ {
		//create solace instance http console service
		svc := SvcConsole(solaceScalable, i)
		if err := controllerutil.SetControllerReference(solaceScalable, svc, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		if err := r.CreateSolaceConsoleSvc(svc, ctx); err != nil {
			return reconcile.Result{}, err
		}

		// create solace instances PV
		if err := r.CreateSolaceLocalPv(solaceScalable, i, ctx); err != nil {
			return reconcile.Result{}, err
		}
	}

	if err := r.DeleteSolaceConsoleSvc(solaceScalable, ctx); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.CreateSolaceConsoleIngress(solaceScalable, ctx); err != nil {
		return reconcile.Result{}, err
	}

	if _, success, _ := CallSolaceSempApi(solaceScalable, "/monitor/about/api", ctx); success == true {
		// get open svc pub/sub ports
		m, err := GetEnabledSolaceMsgVpns(solaceScalable, ctx)
		if err != nil {
			return reconcile.Result{}, err
		}
		c, err := m.GetSolaceClientUsernames(solaceScalable, ctx)
		if err != nil {
			return reconcile.Result{}, err
		}
		pubSubOpenPorts := c.MergeSolaceResponses(m).Data

		pubSvcNames, dataPub, err := r.CreatePubSubSvc(solaceScalable, &pubSubOpenPorts, &m, "pub", ctx)
		if err != nil {
			return reconcile.Result{}, err
		}

		subSvcNames, dataSub, err := r.CreatePubSubSvc(solaceScalable, &pubSubOpenPorts, &m, "sub", ctx)
		if err != nil {
			return reconcile.Result{}, err
		}

		// check HAProxy pub service
		FoundHaproxyPubSvc, err := r.GetExistingHaProxySvc(solaceScalable, solaceScalable.Spec.Haproxy.Publish.ServiceName, ctx)
		if err != nil {
			return reconcile.Result{}, err
		}
		//set the new pub data in the found svc
		FoundHaproxyPubSvc.Spec.Ports = *SvcHaproxy(solaceScalable, FoundHaproxyPubSvc.Spec.Ports, *dataPub)

		if err := r.UpdateHAProxySvc(&hashStore, FoundHaproxyPubSvc, ctx); err != nil {
			return reconcile.Result{}, err
		}

		// check HAProxy pub service
		FoundHaproxySubSvc, err := r.GetExistingHaProxySvc(solaceScalable, solaceScalable.Spec.Haproxy.Subscribe.ServiceName, ctx)
		if err != nil {
			return reconcile.Result{}, err
		}
		//set the new sub data in the found svc
		FoundHaproxySubSvc.Spec.Ports = *SvcHaproxy(solaceScalable, FoundHaproxySubSvc.Spec.Ports, *dataSub)

		if err := r.UpdateHAProxySvc(&hashStore, FoundHaproxySubSvc, ctx); err != nil {
			return reconcile.Result{}, err
		}

		// create and update haproxy pub configmap
		configMapPub, FoundHaproxyConfigMap, err := r.CreateSolaceTcpConfigmap(dataPub, "pub", solaceScalable, ctx)
		if err != nil {
			return reconcile.Result{}, err
		}
		if err := r.UpdateSolaceTcpConfigmap(FoundHaproxyConfigMap, configMapPub, solaceScalable, ctx, &hashStore); err != nil {
			return reconcile.Result{}, err
		}
		// create and update haproxy pub configmap
		configMapSub, FoundHaproxyConfigMap, err := r.CreateSolaceTcpConfigmap(dataSub, "sub", solaceScalable, ctx)
		if err != nil {
			return reconcile.Result{}, err
		}
		if err := r.UpdateSolaceTcpConfigmap(FoundHaproxyConfigMap, configMapSub, solaceScalable, ctx, &hashStore); err != nil {
			return reconcile.Result{}, err
		}

		svcList, foundExtraPubSubSvc, err := ListPubSubSvc(solaceScalable, r)
		if err != nil {
			return reconcile.Result{}, err
		}

		//merge pub and sub svc slices
		pubSubSvcNames := append(*pubSvcNames, *subSvcNames...)
		if err := DeletePubSubSvc(svcList, foundExtraPubSubSvc, &pubSubSvcNames, r, ctx); err != nil {
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
		Owns(&corev1.Service{}).
		Complete(r)
}
