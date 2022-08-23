# solace-scalable-arch-kubernetes-operator
a solace operator to scale horizontally solace standalone instances


# pub Haproxy ingress for pub
helm install --namespace ingress-controller --create-namespace --set controller.ingressClass='haproxy-pub',controller.ingressClassResource.name='haproxy-pub',controller.replicaCount=1,controller.extraArgs={'--configmap-tcp-services=solacescalable/solacescalable-pub-tcp-ingress'} haproxy-pub haproxytech/kubernetes-ingress

# sub Haproxy ingress for sub
helm install --namespace ingress-controller --create-namespace --set controller.ingressClass='haproxy-sub',controller.ingressClassResource.name='haproxy-sub',controller.replicaCount=1,controller.extraArgs={'--configmap-tcp-services=solacescalable/solacescalable-sub-tcp-ingress'} haproxy-sub haproxytech/kubernetes-ingress
