package controllers

import (
	"context"
	"strconv"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	netv1 "k8s.io/api/networking/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func IngressConsole(s *scalablev1alpha1.SolaceScalable, labels map[string]string) *v1.Ingress {
	//labels := labels(s)
	icn := "haproxy-sub"
	annotations := map[string]string{
		"ingress.kubernetes.io/add-base-url": "true",
	}

	return &v1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        s.Name + "-console",
			Namespace:   s.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1.IngressSpec{
			IngressClassName: &icn,
			Rules:            CreateIngressConsoleRules(s),
		},
	}
}

func CreateIngressConsoleRules(s *scalablev1alpha1.SolaceScalable) []v1.IngressRule {
	prefix := v1.PathTypePrefix
	var rules = []v1.IngressRule{}
	for i := 0; i < int(s.Spec.Replicas); i++ {
		rule := v1.IngressRule{
			Host: "n" + strconv.Itoa(i) + "." + s.Spec.ClusterUrl,
			IngressRuleValue: v1.IngressRuleValue{
				HTTP: &v1.HTTPIngressRuleValue{
					Paths: []v1.HTTPIngressPath{
						{
							Path:     "/",
							PathType: &prefix,
							Backend: v1.IngressBackend{
								Service: &v1.IngressServiceBackend{
									Name: s.Namespace + "-" + strconv.Itoa(i),
									Port: v1.ServiceBackendPort{
										Number: 8080,
									},
								},
							},
						},
					},
				},
			},
		}
		rules = append(rules, rule)
	}

	return rules
}

func CreateSolaceConsoleIngress(solaceScalable *scalablev1alpha1.SolaceScalable, r *SolaceScalableReconciler, ctx context.Context) error {
	//create ingress console services
	log := log.FromContext(ctx)
	foundIngress := &netv1.Ingress{}
	ingConsole := IngressConsole(solaceScalable, Labels(solaceScalable))
	if err := r.Get(context.TODO(), types.NamespacedName{Name: ingConsole.Name, Namespace: ingConsole.Namespace}, foundIngress); err != nil && errors.IsNotFound(err) {
		log.Info("Creating Solace Console Ingress", ingConsole.Namespace, ingConsole.Name)
		if err = r.Create(context.TODO(), ingConsole); err != nil && errors.IsNotFound(err) {
			return err
		}
	}
	return nil
}
