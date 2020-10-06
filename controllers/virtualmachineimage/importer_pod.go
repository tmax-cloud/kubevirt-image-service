package virtualmachineimage

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// DataVolName provides a const to use for creating volumes in pod specs
	DataVolName = "data-vol"
	// WriteBlockPath provides a constant for the path where the PV is mounted.
	WriteBlockPath = "/dev/cdi-block-volume"
	// ImporterSource provides a constant to capture our env variable "IMPORTER_SOURCE"
	ImporterSource = "IMPORTER_SOURCE"
	// ImporterEndpoint provides a constant to capture our env variable "IMPORTER_ENDPOINT"
	ImporterEndpoint = "IMPORTER_ENDPOINT"
	// ImporterContentType provides a constant to capture our env variable "IMPORTER_CONTENTTYPE"
	ImporterContentType = "IMPORTER_CONTENTTYPE"
	// ImporterImageSize provides a constant to capture our env variable "IMPORTER_IMAGE_SIZE"
	ImporterImageSize = "IMPORTER_IMAGE_SIZE"
	// InsecureTLSVar provides a constant to capture our env variable "INSECURE_TLS"
	InsecureTLSVar = "INSECURE_TLS"
	// SourceHTTP is the source type HTTP
	SourceHTTP = "http"
	// SourceHostPath is the source type host path
	SourceHostPath = "hostPath"
	// ImageContentType is the content-type of the imported file
	ImageContentType = "kubevirt"
	// ImportPodImage and ImportPodVerbose should be modified to get value from vmi env
	// ImportPodImage indicates image name of the import pod
	ImportPodImage = "kubevirt/cdi-importer:v1.13.0"
	// ImportPodVerbose indicates log level of the import pod
	ImportPodVerbose = "1"
	// SourceVolumeName is used for creating source volume in pod specs
	SourceVolumeName = "source-vol"
	// SourceVolumeMountPath is a path where the source volume is mounted
	SourceVolumeMountPath = "/data/source"
)

func (r *ReconcileVirtualMachineImage) syncImporterPod() error {
	imported, found, err := r.isPvcImported()
	if err != nil {
		return err
	} else if !found {
		klog.Warningf("syncImporterPod without pvc in vmi %s", r.vmi.Name)
		return nil
	}

	importerPod := &corev1.Pod{}
	err = r.Client.Get(context.Background(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetImporterPodNameFromVmiName(r.vmi.Name)}, importerPod)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	existsImporterPod := err == nil

	if !imported && existsImporterPod && isPodCompleted(importerPod) {
		// 임포팅이 완료됐으니 애노테이션을 업데이트하고 삭제한다.
		klog.Infof("syncImporterPod finish for vmi %s, delete importerPod", r.vmi.Name)
		if err := r.updatePvcImported(true); err != nil {
			return err
		}
		if err := r.Client.Delete(context.TODO(), importerPod); err != nil && !errors.IsNotFound(err) {
			return err
		}
	} else if !imported && !existsImporterPod {
		// 임포팅을 해야 하므로 임포터파드를 만든다
		klog.Infof("syncImporterPod create new importerPod for vmi %s", r.vmi.Name)
		newPod, err := r.newImporterPod()
		if err != nil {
			return err
		}
		if err := r.Client.Create(context.TODO(), newPod); err != nil && !errors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func isPodCompleted(pod *corev1.Pod) bool {
	return len(pod.Status.ContainerStatuses) != 0 &&
		pod.Status.ContainerStatuses[0].State.Terminated != nil &&
		pod.Status.ContainerStatuses[0].State.Terminated.Reason == "Completed"
}

// GetImporterPodNameFromVmiName returns ImporterPod name from VmiName
func GetImporterPodNameFromVmiName(vmiName string) string {
	return vmiName + "-image-importer"
}

func (r *ReconcileVirtualMachineImage) newImporterPod() (*corev1.Pod, error) {
	ip := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetImporterPodNameFromVmiName(r.vmi.Name),
			Namespace: r.vmi.Namespace,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyOnFailure,
			Containers: []corev1.Container{
				{
					Name:  "importer",
					Image: ImportPodImage,
					Resources: corev1.ResourceRequirements{
						Limits: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse("0"),
							corev1.ResourceMemory: resource.MustParse("0")},
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse("0"),
							corev1.ResourceMemory: resource.MustParse("0")},
					},
					VolumeDevices: []corev1.VolumeDevice{
						{Name: DataVolName, DevicePath: WriteBlockPath},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: DataVolName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: GetPvcNameFromVmiName(r.vmi.Name),
						},
					},
				},
			},
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser: &[]int64{0}[0],
			},
		},
	}

	src, err := r.getSource()
	if err != nil {
		return nil, err
	}
	if src == SourceHTTP {
		pvcSize := r.vmi.Spec.PVC.Resources.Requests[corev1.ResourceStorage]

		ip.Spec.Containers[0].Args = []string{"-v=" + ImportPodVerbose}
		ip.Spec.Containers[0].Env = []corev1.EnvVar{
			{Name: ImporterSource, Value: SourceHTTP},
			{Name: ImporterEndpoint, Value: r.vmi.Spec.Source.HTTP},
			{Name: ImporterContentType, Value: ImageContentType},
			{Name: ImporterImageSize, Value: pvcSize.String()},
			{Name: InsecureTLSVar, Value: "true"},
		}
	} else if src == SourceHostPath {
		ip.Spec.NodeName = r.vmi.Spec.Source.HostPath.NodeName
		ip.Spec.Containers[0].Command = []string{"qemu-img", "convert", "-f", "qcow2", "-O", "raw", SourceVolumeMountPath + "/disk.img", WriteBlockPath}
		ip.Spec.Volumes = append(ip.Spec.Volumes, corev1.Volume{
			Name: SourceVolumeName,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: r.vmi.Spec.Source.HostPath.Path,
				}},
		})
		ip.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{Name: SourceVolumeName, MountPath: SourceVolumeMountPath}}
	}
	if err := controllerutil.SetControllerReference(r.vmi, ip, r.Scheme); err != nil {
		return nil, err
	}
	return ip, nil
}
