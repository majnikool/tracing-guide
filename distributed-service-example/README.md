# distributed-service-example
An example deployment and instrumentation of a distributed service that
is automatically instrumented by Beyla.

Requirements:
* Local Kubernetes (Kind or K3d are fine)

## Run the test environment and services

With Kind:

```
kind create cluster
make build-all push-all-kind
kubectl apply -f all-services.yml
```

With K3d:

```
k3d cluster create
make build-all push-all-k3d
kubectl apply -f all-services.yml
```

## Instrument all the services as a Daemonset

First, make a copy of the `grafana-credentials.template.yml` file and
fill the gaps with your grafana endpoint and credentials information:

```
cp grafana-credentials.template.yml grafana-credentials.yml
vi grafana-credentials.yml
# after editing...
kubectl apply -f grafana-credentials.yml
```

Then you can deploy Grafana Beyla along with the grafana agent. Maybe you are interested to provide
a finer-grained configuration by editing the `beyla-config` configmap
that is inside the `beyla-daemonset.yml` file:

```
kubectl apply -f beyla-daemonset.yml
```


