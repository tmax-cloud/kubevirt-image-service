package util

import (
	snapshotv1beta1 "github.com/kubernetes-csi/external-snapshotter/v2/pkg/apis/volumesnapshot/v1beta1"
	hc "github.com/tmax-cloud/kubevirt-image-service/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// CreateFakeClientAndScheme returns fake client that includes objects given by the parameter
func CreateFakeClientAndScheme(objects ...runtime.Object) (client.Client, *runtime.Scheme, error) {
	s := scheme.Scheme
	if err := hc.SchemeBuilder.AddToScheme(s); err != nil {
		return nil, nil, err
	}
	if err := snapshotv1beta1.AddToScheme(s); err != nil {
		return nil, nil, err
	}
	var objs []runtime.Object
	objs = append(objs, objects...)
	return fake.NewFakeClientWithScheme(s, objs...), s, nil
}
