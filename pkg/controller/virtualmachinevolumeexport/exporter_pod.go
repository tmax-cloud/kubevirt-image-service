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
	vmv "kubevirt-image-service/pkg/controller/virtualmachinevolume"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// SourceVolumeName provides a const to use for source volume in pod specs
	SourceVolumeName = "source-volume"
	// ExportVolumeName provides a const to use for export volume in pod specs
	ExportVolumeName = "export-volume"
	// SourceDevicePath provides device path for source volume in pod specs
	SourceDevicePath = "/dev/source-block-volume"
	// SourceDataDir provides mount path for source volume in pod specs
	SourceDataDir = "/source"
	// ExportDataDir provides mount path for export volume in pod specs
	ExportDataDir = "/export"
	// ExporterSourcePath  provides a constant to capture our env variable "EXPORTER_SOURCE_PATH"
	ExporterSourcePath = "EXPORTER_SOURCE_PATH"
	// ExporterExportDir  provides a constant to capture our env variable "EXPORTER_EXPORT_DIR"
	ExporterExportDir = "EXPORTER_EXPORT_DIR"
	// ExporterDestination provides a constant to capture our env variable "EXPORTER_DESTINATION"
	ExporterDestination = "EXPORTER_DESTINATION"
	// ExporterName indicates container name in exporter pod
	ExporterName = "kubevirt-image-service-exporter"
	// ExporterImage indicates exporter container image
	ExporterImage = "quay.io/tmaxanc/kubevirt-image-service-exporter:v1.2.0"
	// ExporterDestinationLocal indicates Destination to export is local
	ExporterDestinationLocal = "local"
	// ExporterDestinationS3 indicates Destination to export is s3
	ExporterDestinationS3 = "s3"
	// Endpoint is an endpoint of the external object storage where to export volume
	Endpoint = "ENDPOINT"
	// AccessKeyID is one of AWS-style credential which is needed when export volume to external object storage
	AccessKeyID = "AWS_ACCESS_KEY_ID"
	// SecretAccessKey is one of AWS-style credential which is needed when export volume to external object storage
	SecretAccessKey = "AWS_SECRET_ACCESS_KEY"
)

func (r *ReconcileVirtualMachineVolumeExport) syncExporterPod() error {
	// completed indicates if pvc export is completed
	completed, found, err := r.isPvcExportCompleted()
	if err != nil {
		return err
	} else if !found {
		klog.Warningf("syncExporterPod without pvc in vmvExport %s", r.vmvExport.Name)
		return nil
	}

	exporterPod := &corev1.Pod{}
	err = r.client.Get(context.Background(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: GetExporterPodName(r.vmvExport.Name)}, exporterPod)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	existsExporterPod := err == nil

	if !completed && existsExporterPod && isPodCompleted(exporterPod) {
		// pvc export is completed, update completed to yes and delete exporter pod
		klog.Infof("syncExporterPod finish for vmvExport %s, delete exporterPod", r.vmvExport.Name)
		if err := r.updatePvcCompleted(true); err != nil {
			return err
		}
		if err := r.client.Delete(context.TODO(), exporterPod); err != nil && !errors.IsNotFound(err) {
			return err
		}
		if destination := r.getDestination(); destination != ExporterDestinationLocal {
			if err := r.updateStateWithReadyToUse(hc.VirtualMachineVolumeExportStateCompleted, corev1.ConditionTrue, "vmvExportIsCompleted", "vmvExport is completed"); err != nil {
				return err
			}
		}
	} else if !completed && !existsExporterPod {
		// pvc export is not completed, should create exporter pod
		klog.Infof("syncExporterPod create new exporterPod for vmvExport %s", r.vmvExport.Name)
		newPod, err := r.newExporterPod(r.vmvExport, r.scheme)
		if err != nil {
			return err
		}
		if err := r.client.Create(context.Background(), newPod); err != nil && !errors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

// GetExporterPodName returns exporter pod name from vmvExport name
func GetExporterPodName(vmvExportName string) string {
	return vmvExportName + "-exporter"
}

func (r *ReconcileVirtualMachineVolumeExport) newExporterPod(vmvExport *hc.VirtualMachineVolumeExport, scheme *runtime.Scheme) (*corev1.Pod, error) {
	ep := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetExporterPodName(vmvExport.Name),
			Namespace: vmvExport.Namespace,
			Labels: map[string]string{
				"app": vmvExport.Name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            ExporterName,
					Image:           ExporterImage,
					ImagePullPolicy: corev1.PullPolicy("IfNotPresent"),
					Args:            []string{},
					Env: []corev1.EnvVar{
						{Name: ExporterDestination, Value: r.getDestination()},
						{Name: ExporterSourcePath, Value: SourceDevicePath},
						{Name: ExporterExportDir, Value: ExportDataDir},
					},
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
					VolumeDevices: []corev1.VolumeDevice{
						{Name: SourceVolumeName, DevicePath: SourceDevicePath},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: SourceVolumeName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: vmv.GetVolumePvcName(vmvExport.Spec.VirtualMachineVolume.Name),
						},
					},
				},
				{
					Name: ExportVolumeName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: GetExportPvcName(vmvExport.Name),
						},
					},
				},
			},
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser: &[]int64{0}[0],
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}

	if r.getDestination() == ExporterDestinationS3 {
		ep.Spec.Containers[0].Env = append(ep.Spec.Containers[0].Env, corev1.EnvVar{
			Name: AccessKeyID,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.vmvExport.Spec.Destination.S3.SecretRef,
					},
					Key: AccessKeyID,
				},
			},
		}, corev1.EnvVar{
			Name: SecretAccessKey,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.vmvExport.Spec.Destination.S3.SecretRef,
					},
					Key: SecretAccessKey,
				},
			},
		}, corev1.EnvVar{
			Name: Endpoint,
			Value: r.vmvExport.Spec.Destination.S3.URL,
		})
	}

	if err := controllerutil.SetControllerReference(vmvExport, ep, scheme); err != nil {
		return nil, err
	}
	return ep, nil
}

func isPodCompleted(pod *corev1.Pod) bool {
	return len(pod.Status.ContainerStatuses) != 0 &&
		pod.Status.ContainerStatuses[0].State.Terminated != nil &&
		pod.Status.ContainerStatuses[0].State.Terminated.Reason == "Completed"
}
