module kubevirt-image-service

go 1.13

require (
	github.com/golang/protobuf v1.4.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/kube-object-storage/lib-bucket-provisioner v0.0.0-20200610144127-e2eec875d6d1
	github.com/kubernetes-csi/external-snapshotter/v2 v2.1.1
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/operator-framework/operator-sdk v0.18.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20200425230154-ff2c4b7c35a0 // indirect
	golang.org/x/sys v0.0.0-20200420163511-1957bb5e6d1f // indirect
	golang.org/x/tools v0.0.0-20200423205358-59e73619c742 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
