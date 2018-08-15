kubectl create namespace custom-metrics

cat <<EOF | cfssl genkey - | cfssljson -bare serving
{
  "hosts": [
    "team1-system.prometheus-team1.svc.cluster.local"
  ],
  "CN": "team1-system.prometheus-team1.pod.cluster.local",
  "key": {
    "algo": "ecdsa",
    "size": 256
  }
}
EOF

mv serving-key.pem serving.key

kubectl delete csr custom-metrics.prometheus-team1

cat <<EOF | kubectl apply -f -
apiVersion: certificates.k8s.io/v1beta1
kind: CertificateSigningRequest
metadata:
  name: custom-metrics.prometheus-team1
spec:
  groups:
  - system:authenticated
  request: $(cat serving.csr | base64 | tr -d '\n')
  usages:
  - digital signature
  - key encipherment
  - server auth
EOF

kubectl certificate approve custom-metrics.prometheus-team1

kubectl get csr custom-metrics.prometheus-team1 -o jsonpath='{.status.certificate}' | base64 --decode > serving.crt

kubectl -n custom-metrics create secret generic cm-adapter-serving-certs --from-file=serving.key --from-file=serving.crt --dry-run -o yaml | kubectl apply -f -

rm serving.key serving.crt serving.csr