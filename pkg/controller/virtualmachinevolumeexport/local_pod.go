package virtualmachinevolumeexport

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileVirtualMachineVolumeExport) syncLocalPod() error {
	// completed indicates if pvc export is completed
	completed, found, err := r.isPvcExportCompleted()
	if err != nil {
		return err
	} else if !found {
		klog.Warningf("syncLocalPod without pvc in vmvExport %s", r.vmvExport.Name)
		return nil
	}

	localPod := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: getLocalPodName(r.vmvExport.Name)}, localPod)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	existsLocalPod := err == nil

	if completed && !existsLocalPod {
		// pvc export is completed but there is no local pod, so create a local pod and update readytouse to true
		klog.Infof("syncLocalPod create new localPod for vmvExport %s", r.vmvExport.Name)
		newPod, err := newLocalPod(r.vmvExport, r.scheme)
		if err != nil {
			return err
		}
		if err := r.client.Create(context.TODO(), newPod); err != nil && !errors.IsAlreadyExists(err) {
			return err
		}
		if err := r.updateStateWithReadyToUse(r.vmvExport, true, hc.VirtualMachineVolumeExportStateCompleted); err != nil {
			return err
		}
	} else if !completed && existsLocalPod {
		// pvc export is not completed but there is a local pod, so delete the local pod
		if err := r.client.Delete(context.TODO(), localPod); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func getLocalPodName(vmvExportName string) string {
	return vmvExportName + "-exporter-local"
}

func newLocalPod(vmvExport *hc.VirtualMachineVolumeExport, scheme *runtime.Scheme) (*corev1.Pod, error) {
	lp := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getLocalPodName(vmvExport.Name),
			Namespace: vmvExport.Namespace,
			Labels: map[string]string{
				"app": vmvExport.Name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "busybox",
					Image:           "busybox",
					Command:         []string{"sleep", "600"},
					ImagePullPolicy: corev1.PullPolicy("IfNotPresent"),
					Resources: corev1.ResourceRequirements{
						Limits: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse("0"),
							corev1.ResourceMemory: resource.MustParse("0")},
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse("0"),
							corev1.ResourceMemory: resource.MustParse("0")},
					},
					VolumeMounts: []corev1.VolumeMount{
						{Name: ExportVolumeName, MountPath: ExportDataDir},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: ExportVolumeName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: GetExportPvcName(vmvExport.Name),
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyAlways,
		},
	}
	if err := controllerutil.SetControllerReference(vmvExport, lp, scheme); err != nil {
		return nil, err
	}
	return lp, nil
}
