```
# Kubevirt 설치
export KUBEVIRT_VERSION=v0.27.0
kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-operator.yaml
kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-cr.yaml

# CDI 설치
export VERSION=v1.13.2
kubectl apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/$VERSION/cdi-operator.yaml
kubectl apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/$VERSION/cdi-cr.yaml

# rook-ceph 설치
kubectl apply -f common.yaml
kubectl apply -f operator.yaml
kubectl apply -f cluster-test.yaml
kubectl apply -f storageclass-test.yaml
```
