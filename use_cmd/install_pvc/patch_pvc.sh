kubectl patch configmap config-features -n knative-serving -p '{"data":{"kubernetes.podspec-persistent-volume-claim":"enabled"}}'
kubectl patch configmap config-features -n knative-serving -p '{"data": {"kubernetes.podspec-persistent-volume-write":"enabled"}}'