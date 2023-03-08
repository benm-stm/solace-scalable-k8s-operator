## Contributing
in order to contribute you open a pull request and we will discuss it :)

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

### Important Note
When testing the operator from your local machine, the operator is interacting with kubernetes cluster from its API
This way the dns feature won't resolve the svc ip
You have to forward the port and set it in the CR definition in the ClusterUrl field
This way, solace will be visible to the operator
```sh
$ make run
```
open a new terminal and run the following command
````sh
$ kubectl port-forward -n solacescalable svc/solacescalable-0 27017:8080
````


you can check your solace interface using **localhost:27017**
change you CR definition field **ClusterUrl** and set **localhost:27017**