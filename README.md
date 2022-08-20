# solace-scalable-arch-kubernetes-operator
a solace operator to scale horizontally solace standalone instances


# pub Haproxy ingress as Daemonset
helm install --namespace ingress --create-namespace --set controller.ingressClass='haproxy-pub',controller.ingressClassResource.name='haproxy-pub',controller.kind='DaemonSet',controller.extraArgs={'--configmap-tcp-services=solacescalable/solacescalable-pub-tcp-ingress'}  haproxy-pub haproxytech/kubernetes-ingress

# sub Haproxy ingress as Daemonset
helm install --namespace ingress --create-namespace --set controller.ingressClass='haproxy-sub',controller.ingressClassResource.name='haproxy-sub',controller.kind='DaemonSet',controller.extraArgs={'--configmap-tcp-services=solacescalable/solacescalable-sub-tcp-ingress'}  haproxy-sub haproxytech/kubernetes-ingress