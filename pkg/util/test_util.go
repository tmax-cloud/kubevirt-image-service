package util

import (
	snapshotv1alpha1 "github.com/kubernetes-csi/external-snapshotter/pkg/apis/volumesnapshot/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// CreateFakeClientAndScheme 는 인자로 주어진 objects 를 포함한 fake 클라이언트를 리턴합니다
func CreateFakeClientAndScheme(objects ...runtime.Object) (client.Client, *runtime.Scheme, error) {
	s := scheme.Scheme
	if err := hc.SchemeBuilder.AddToScheme(s); err != nil {
		return nil, nil, err
	}
	if err := snapshotv1alpha1.AddToScheme(s); err != nil {
		return nil, nil, err
	}
	var objs []runtime.Object
	objs = append(objs, objects...)
	return fake.NewFakeClientWithScheme(s, objs...), s, nil
}
