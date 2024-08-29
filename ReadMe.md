# Distributed Tracing Setup Guide

This guide provides step-by-step instructions for setting up distributed tracing in a Kubernetes environment using Tempo, OpenTelemetry Operator, Istio, and Grafana.

## 1. Install Tempo

1.1. Search for available versions of the Tempo Helm chart:
```
helm search repo grafana/tempo-distributed --versions
```

1.2. Generate a values file for the desired version:
```
helm show values grafana/tempo-distributed --version 1.9.3 > tempo-overrides.yaml
```

1.3. Create a namespace for tracing:
```
kubectl create namespace tracing
```

1.4. Install Tempo using the generated values file:
```
helm upgrade --install tempo grafana/tempo-distributed --values tempo-values.yaml -n tracing
```

The values file used for this installation can be found at:
https://github.com/grafana/tempo/blob/main/example/helm/microservices-tempo-values.yaml

## 2. Deploy OpenTelemetry Operator

2.1. Install cert-manager (required by OpenTelemetry Operator):
```
helm repo add jetstack https://charts.jetstack.io
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
helm repo update
helm install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --version v1.14.4 --set installCRDs=true
```

2.2. Install OpenTelemetry Operator:
```
helm install opentelemetry-operator open-telemetry/opentelemetry-operator \
-n tracing \
--set manager.collectorImage.repository=otel/opentelemetry-collector-k8s


or otel/opentelemetry-collector-contrib

helm upgrade opentelemetry-operator open-telemetry/opentelemetry-operator \
-n tracing \
--set manager.collectorImage.repository=otel/opentelemetry-collector-contrib

more info at : https://github.com/open-telemetry/opentelemetry-helm-charts/blob/main/charts/opentelemetry-operator/UPGRADING.md#0553-to-0560

```

2.3. Fetch the OpenTelemetry Operator chart:
```
helm search repo open-telemetry/opentelemetry-operator --versions
helm show values open-telemetry/opentelemetry-operator --version 0.55.3 > otl-operator-overrides.yaml
```

2.4. Enable debug logging for the OpenTelemetry Operator:
```
kubectl patch -n tracing $(kubectl get -n tracing -l app.kubernetes.io/name=opentelemetry-operator deployment -o name) --type=json -p '[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--zap-log-level=debug"}]'
```

2.5. Create an OpenTelemetry Collector (configure the Tempo endpoint here):
```
kubectl apply -f otl-collector.yaml -n tracing
```

## 3. Install Istio with Demo Profile and Kernel Modules

3.1. Run the `setup_istio_modules.sh` script on all nodes to install the required kernel modules before setting up Istio.

3.2. Ensure the demo profile is added to all Istio charts. An example can be found at:
https://git.xx.com/devops/mj-rancher-appcluster-flux/-/blob/master/system/infrastructure/controllers/istio/base.yaml?ref_type=heads

## 4. Deploy Beyla (Optional)

If you need auto-instrumentation and are not using OpenTelemetry auto-instrumentation, deploy Beyla:
```
kubectl apply -f beyla-daemonset.yml
```

Update the watched namespace in the YAML file:
```yaml
discovery:
 services:
  - k8s_namespace: demo
```

Update the OpenTelemetry endpoint in the YAML file:
```yaml
value: "http://opentelemetry-collector.tracing.svc.cluster.local:4317"
```

## 5. Enable Remote Write Receiver in Rancher Monitoring

Enable the following configuration in the Rancher Monitoring chart values, under `prometheusSpec.prometheus`:
```yaml
enableRemoteWriteReceiver: true
```

To test the configuration:
```
kubectl run -it --rm --restart=Never --image=curlimages/curl test-net -- curl -X POST -d '{}' http://rancher-monitoring-prometheus.cattle-monitoring-system.svc.cluster.local:9090/api/v1/write
```

## 6. Apply Telemetry Object in Istio Namespace

Apply the Telemetry object in the Istio namespace (the `randomSamplingPercentage` can be set here):
```
kubectl apply -f telemetry.yaml -n istio-system
```

## 7. Update Mesh Config and Define Telemetry

Update the Istio ConfigMap and define telemetry in the mesh config:
```
kubectl get configmap istio -n istio-system -o yaml
```

Example config:
```yaml
meshConfig:
 accessLogFile: /dev/stdout
 accessLogFormat: |
  [%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %RESPONSE_CODE_DETAILS% %CONNECTION_TERMINATION_DETAILS% "%UPSTREAM_TRANSPORT_FAILURE_REASON%" %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%" %UPSTREAM_CLUSTER% %UPSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_LOCAL_ADDRESS% %DOWNSTREAM_REMOTE_ADDRESS% %REQUESTED_SERVER_NAME% %ROUTE_NAME% traceID=%REQ(TRACEPARENT)%
 enableAutoMtls: true
extensionProviders:
- opentelemetry:
  port: 4317
  service: otel-collector-collector.tracing.svc.cluster.local
  name: otel-tracing
```

## 8. Update Ingress ConfigMap

Update the Ingress ConfigMap by applying the `update-ingress-configmap.yaml` file (the OpenTelemetry Collector endpoint is defined here).

## 9. Create OpenTelemetry Auto-Instrumentation Object

Create an OpenTelemetry auto-instrumentation object based on the programming language used in your application. Examples can be found in this repo

Refer to the README for a list of supported languages and their templates. Be careful about the collector endpoint port used in each config, as it can be gRPC or HTTP, and they are different but mentioned in the README for each language.

https://opentelemetry.io/docs/kubernetes/operator/automatic/

## 10. Create Telemetry Sidecar Object

Create a Telemetry sidecar object in your application namespace. Be careful about the ports. Examples can be found in `test-app-sidecar.yaml` and `drm-ns-sidecar.yaml`.

Annotate the namespace and deployment with the appropriate instrumentation annotation:
```
kubectl annotate namespace test-app instrumentation.opentelemetry.io/inject-python="true"
kubectl annotate deployment test-app-0-1713262482 instrumentation.opentelemetry.io/inject-python="true" -n test-app
```

## 11. Redeploy Application

Redeploy your application, and you should see Istio and OpenTelemetry containers in your pod's container list.

## 12. Push Application Logs to Loki

Follow this guide to push application logs to Loki:
https://git.xx.com/devops/loki-logging

## 13. Configure Grafana Data Sources

In Grafana, add the following data sources:
```yaml
datasources:
  datasources.yaml:
   apiVersion: 1
   datasources:
   - access: proxy
    isDefault: true
    name: Loki
    type: loki
    url: http://logging-loki-gateway.logging.svc.cluster.local
    jsonData:
     derivedFields:
      - datasourceName: Tempo
       matcherRegex: "traceID=00-([^\\-]+)-"
       name: traceID
       url: "$${__value.raw}"
       datasourceUid: tempo
   - name: Tempo
    type: tempo
    uid: tempo
    access: proxy
    url: http://tempo-query-frontend.tracing.svc.cluster.local:3100
    jsonData:
     nodeGraph:
      enabled: true
      serviceGraph:
       enabled: true
       datasourceUid: 'prometheus'
   - name: Prometheus
    type: prometheus
    uid: prometheus # Ensure this UID is referenced in the Tempo serviceGraph settings
    access: proxy
    url: http://rancher-monitoring-prometheus.cattle-monitoring-system.svc.cluster.local:9090
```

Also, enable the "Service graph data source" in the Additional settings of the Tempo data source to connect it to Prometheus.

With these steps completed, you should have a fully functional distributed tracing setup using Tempo, OpenTelemetry, Istio, and Grafana in your Kubernetes environment.
