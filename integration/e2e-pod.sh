# kubectl apply -f ./manifests/executor.yaml
# kubectl wait --for=condition=ready --timeout=60s deploy/integration
pod=$(kubectl get po -l run=integration -o jsonpath="{.items[0].metadata.name}")
for f in *.js; do
    kubectl cp "./${f}" "${pod}:/app/"
done
kubectl cp ./use-cases "${pod}:/app/"
kubectl exec -it "$pod" -- /bin/sh -c "INTEGRATION_ENTRYPOINT=http://kubeapps.kubeapps yarn start"
