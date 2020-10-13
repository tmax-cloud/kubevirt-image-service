module github.com/tmax-cloud/kubevirt-image-service

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/kube-object-storage/lib-bucket-provisioner v0.0.0-20200610144127-e2eec875d6d1
	github.com/kubernetes-csi/external-snapshotter/v2 v2.1.1
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.6.2
)
