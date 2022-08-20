package controllers

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	scalablev1alpha1 "solace.io/api/v1alpha1"
)

func SvcConsole(s *scalablev1alpha1.SolaceScalable, counter int) *corev1.Service {
	//labels := labels(s)
	name := s.Name + "-" + strconv.Itoa(counter)
	selector := map[string]string{
		"statefulset.kubernetes.io/pod-name": name,
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: s.Namespace,
		},
		Spec: corev1.ServiceSpec{
			//Selector: labels,
			Selector: selector,
			Ports: []corev1.ServicePort{{
				Protocol: corev1.ProtocolTCP,
				Port:     8080,
				//TargetPort: intstr.FromInt(8080 + counter),
				//NodePort:   30685,

			}},
			Type: corev1.ServiceTypeClusterIP,
			//Type: corev1.ServiceTypeLoadBalancer,
		},
	}
}

func SvcPubSub(s *scalablev1alpha1.SolaceScalable, m solaceMergedResp, p int32, pubSub string) *corev1.Service {
	labels := labels(s)
	//name := s.Name

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.MsgVpnName + "-" + m.ClientUsername + "-" + strconv.FormatInt(int64(p), 10) + "-" + pubSub,
			Namespace: s.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol: corev1.ProtocolTCP,
				//TargetPort: intstr.FromInt(int(p)),
				Port: p,
			}},
			//Type: corev1.ServiceTypeLoadBalancer,
			Type: corev1.ServiceTypeClusterIP,
			//Type: corev1.ServiceTypeNodePort,
		},
	}
}

func SvcHaproxy(s *scalablev1alpha1.SolaceScalable, ports []corev1.ServicePort, d map[string]string) *[]corev1.ServicePort {
	// get default
	svcPorts := *getDefaultHaProxyConf(ports)
	var portExist bool
	var portIndex int

	for key := range d {
		portExist = false
		portIndex = 0
		svcPort := corev1.ServicePort{}

		//fmt.Println("Key:", key)
		port, err := strconv.Atoi(key)
		if err == nil {
			//check if the svc exist
			for i := 0; i < len(ports); i++ {
				if ports[i].Port == int32(port) {
					portExist = true
					portIndex = i
				}
			}
			if !portExist {
				//create new serviceport
				svcPort = corev1.ServicePort{
					Name:     "tcp-" + key,
					Protocol: "TCP",
					Port:     int32(port),
					//TargetPort:  intstr.IntOrString{},
					//NodePort: 0,
					AppProtocol: nil,
				}
				//fmt.Println(port)
			} else {
				svcPort = ports[portIndex]
			}
			svcPorts = append(svcPorts, svcPort)
		}
	}
	return &svcPorts

}

func getDefaultHaProxyConf(servicePorts []corev1.ServicePort) *[]corev1.ServicePort {
	var svcPorts = []corev1.ServicePort{}
	for i := 0; i < len(servicePorts); i++ {
		if servicePorts[i].Name == "http" || servicePorts[i].Name == "https" || servicePorts[i].Name == "stat" {
			servicePorts[i].AppProtocol = nil
			svcPorts = append(svcPorts, servicePorts[i])
		}
	}
	return &svcPorts
}

//create console service
func CreateSolaceConsoleSvc(svc *corev1.Service, r *SolaceScalableReconciler, ctx context.Context) error {
	// Check if the console svc already exists
	log := log.FromContext(ctx)
	foundSvc := &corev1.Service{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, foundSvc); err != nil {
		log.Info("Creating Solace Console Svc", svc.Namespace, svc.Name)
		if err = r.Create(context.TODO(), svc); err != nil {
			return err
		}
	}
	return nil
}

//delete unused console services
func DeleteSolaceConsoleSvc(solaceScalable *scalablev1alpha1.SolaceScalable, r *SolaceScalableReconciler, ctx context.Context) error {
	log := log.FromContext(ctx)
	i := int(solaceScalable.Spec.Replicas)
	// loop indefinitely until not finding an existi_ng console service
	for true {
		svc := SvcConsole(solaceScalable, i)
		foundExtraSvc := &corev1.Service{}
		if err := r.Get(ctx, types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, foundExtraSvc); err != nil {
			break
		} else {
			log.Info("Delete Solace Console Service", svc.Namespace, svc.Name)
			if err = r.Delete(ctx, foundExtraSvc); err != nil {
				return err
			}
			i++
		}
	}
	return nil
}

// TODO: pubsub SVC creation
func CreatePubSubSvc(solaceScalable *scalablev1alpha1.SolaceScalable, pubSubOpenPorts *[]solaceMergedResp, enabledMsgVpns *solaceMsgVpnsResp, nature string, r *SolaceScalableReconciler, ctx context.Context) (*[]string, *map[string]string, error) {
	log := log.FromContext(ctx)
	var pubSubSvcNames = []string{}
	var data = map[string]string{}
	for i := 0; i < len(*pubSubOpenPorts); i++ {
		for j := 0; j < len((*pubSubOpenPorts)[i].Ports); j++ {
			if (*pubSubOpenPorts)[i].Ports[j] != 0 {
				//create pubsub SVC
				pubSubSvcNames = append(pubSubSvcNames, (*pubSubOpenPorts)[i].MsgVpnName+"-"+(*pubSubOpenPorts)[i].ClientUsername+"-"+strconv.FormatInt(int64((*pubSubOpenPorts)[i].Ports[j]), 10)+"-"+nature)
				pbs := SvcPubSub(solaceScalable, (*pubSubOpenPorts)[i], (*pubSubOpenPorts)[i].Ports[j], nature)
				foundPubSubSvc := &corev1.Service{}
				if err := r.Get(context.TODO(), types.NamespacedName{Name: pbs.Name, Namespace: pbs.Namespace}, foundPubSubSvc); err != nil {
					log.Info("Creating pubSub SVC", pbs.Namespace, pbs.Name)
					if err = r.Create(context.TODO(), pbs); err != nil {
						return nil, nil, err
					}
				}
				data[strconv.Itoa(int((*pubSubOpenPorts)[i].Ports[j]))] = solaceScalable.Namespace + "/" + (*pubSubOpenPorts)[i].MsgVpnName + "-" + (*pubSubOpenPorts)[i].ClientUsername + "-" + strconv.Itoa(int((*pubSubOpenPorts)[i].Ports[j])) + "-" + nature + ":" + strconv.Itoa(int((*pubSubOpenPorts)[i].Ports[j]))
			}
		}
	}
	return &pubSubSvcNames, &data, nil
}

func GetExistingHaProxySvc(solaceScalable *scalablev1alpha1.SolaceScalable, serviceName string, r *SolaceScalableReconciler, ctx context.Context) (*corev1.Service, error) {
	log := log.FromContext(ctx)
	FoundHaproxySvc := &corev1.Service{}
	//newMarshal := []byte{}
	if err := r.Get(context.TODO(), types.NamespacedName{Namespace: solaceScalable.Spec.Haproxy.Namespace, Name: serviceName}, FoundHaproxySvc); err != nil {
		//newMarshal, _ = json.Marshal(FoundHaproxySvc.Spec.Ports)
		log.Info("HAProxy service is not found", FoundHaproxySvc.Namespace, FoundHaproxySvc.Name)
		return nil, err
	}
	return FoundHaproxySvc, nil
}

func UpdateHAProxySvc(hashStore *map[string]string, FoundHaproxySvc *corev1.Service, r *SolaceScalableReconciler, ctx context.Context) error {
	log := log.FromContext(ctx)
	// sort the data (ports cause marshall to fail)
	sort.Slice(FoundHaproxySvc.Spec.Ports, func(i, j int) bool {
		return FoundHaproxySvc.Spec.Ports[i].Name < FoundHaproxySvc.Spec.Ports[j].Name
	})
	portsMarshal, _ := json.Marshal(FoundHaproxySvc.Spec.Ports)
	if len(*hashStore) == 0 {
		(*hashStore)[FoundHaproxySvc.Name] = asSha256(portsMarshal)
	} else if asSha256(portsMarshal) != (*hashStore)[FoundHaproxySvc.Name] {
		log.Info("Updating Haproxy Svc", FoundHaproxySvc.Namespace, FoundHaproxySvc.Name)
		(*hashStore)[FoundHaproxySvc.Name] = asSha256(portsMarshal)
		if err := r.Update(context.TODO(), FoundHaproxySvc); err != nil && errors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func ListPubSubSvc(solaceScalable *scalablev1alpha1.SolaceScalable, r *SolaceScalableReconciler) (*corev1.ServiceList, *corev1.Service, error) {
	// get existing svc list
	foundExtraPubSubSvc := &corev1.Service{}
	svcList := &corev1.ServiceList{}
	listOptions := &client.ListOptions{Namespace: solaceScalable.Namespace}

	if err := r.List(context.TODO(), svcList, listOptions); err != nil && errors.IsNotFound(err) {
		return nil, nil, err
	}
	return svcList, foundExtraPubSubSvc, nil
}
func DeletePubSubSvc(svcList *corev1.ServiceList, foundExtraPubSubSvc *corev1.Service, pubSubSvcNames *[]string, r *SolaceScalableReconciler, ctx context.Context) error {
	log := log.FromContext(ctx)
	for i := 0; i < len(svcList.Items); i++ {
		if !stringInSlice(svcList.Items[i].Name, *pubSubSvcNames) && svcList.Items[i].Spec.Ports[0].Port != 8080 {
			if err := r.Get(context.TODO(), types.NamespacedName{Namespace: svcList.Items[i].Namespace, Name: svcList.Items[i].Name}, foundExtraPubSubSvc); err != nil && errors.IsNotFound(err) {
				break
			} else {
				log.Info("Delete PubSubSvc", svcList.Items[i].Namespace, svcList.Items[i].Name)
				if err = r.Delete(context.TODO(), foundExtraPubSubSvc); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
