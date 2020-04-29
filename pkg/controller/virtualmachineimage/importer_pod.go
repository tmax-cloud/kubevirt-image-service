package virtualmachineimage

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// DataVolName provides a const to use for creating volumes in pod specs
	DataVolName = "data-vol"
	// ScratchVolName provides a const to use for creating scratch pvc volumes in pod specs
	ScratchVolName = "scratch-vol"
	// ImporterDataDir provides a constant for the controller pkg to use as a hardcoded path to where content is transferred to/from (controller only)
	ImporterDataDir = "/data"
	// ScratchDataDir provides a constant for the controller pkg to use as a hardcoded path to where scratch space is located.
	ScratchDataDir = "/scratch"
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
	// SourceHTTP is the source type HTTP, if unspecified or invalid, it defaults to SourceHTTP
	SourceHTTP = "http"
	// ImageContentType is the content-type of the imported file
	ImageContentType = "kubevirt"

	// ImportPodImage and ImportPodVerbose should be modified to get value from vmi env
	// ImportPodImage indicates image name of the import pod
	ImportPodImage = "kubevirt/cdi-importer:v1.13.0"
	// ImportPodVerbose indicates log level of the import pod
	ImportPodVerbose = "1"
)

func (r *ReconcileVirtualMachineImage) getImporterPod() (*corev1.Pod, error) {
	ipName := GetImporterPodName(r.vmi.Name)
	ipNamespace := r.getNamespace()
	pod := &corev1.Pod{}
	if err := r.client.Get(context.Background(), types.NamespacedName{Namespace: ipNamespace, Name: ipName}, pod); err != nil {
		return nil, err
	}
	r.log.Info("Get Pod", "pod", pod)
	return pod, nil
}

func (r *ReconcileVirtualMachineImage) createImporterPod() (*corev1.Pod, error) {
	r.log.Info("Create new importerPod", "name", GetImporterPodName(r.vmi.Name), "namespace", r.getNamespace())
	ip, err := r.newImporterPod()
	if err != nil {
		return nil, err
	}
	if err := r.client.Create(context.Background(), ip); err != nil {
		return nil, err
	}
	return ip, nil
}

func (r *ReconcileVirtualMachineImage) deleteImporterPod() error {
	ipName := GetImporterPodName(r.vmi.Name)
	ipNamespace := r.getNamespace()
	r.log.Info("Delete importerPod", "name", ipName, "namespace", ipNamespace)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ipName,
			Namespace: ipNamespace,
		},
	}
	return r.client.Delete(context.Background(), pod)
}

// GetImporterPodName is called to create importer pod name
func GetImporterPodName(vmiName string) string {
	return vmiName + "-image-importer"
}

func (r *ReconcileVirtualMachineImage) newImporterPod() (*corev1.Pod, error) {
	ipName := GetImporterPodName(r.vmi.Name)
	ipNamespace := r.getNamespace()
	source, endpoint, err := r.getSourceAndEndpoint()
	if err != nil {
		return nil, err
	}
	pvcSize, found := r.vmi.Spec.PVC.Resources.Requests[corev1.ResourceStorage]
	if !found {
		return nil, errors.NewBadRequest("storage request in pvc is missing")
	}
	ip := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ipName,
			Namespace: ipNamespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  ipName,
					Image: ImportPodImage,
					Args:  []string{"-v=" + ImportPodVerbose},
					Env: []corev1.EnvVar{
						{Name: ImporterSource, Value: source},
						{Name: ImporterEndpoint, Value: endpoint},
						{Name: ImporterContentType, Value: ImageContentType},
						{Name: ImporterImageSize, Value: pvcSize.String()},
						{Name: InsecureTLSVar, Value: "true"},
					},
					Resources: corev1.ResourceRequirements{
						Limits: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse("0"),
							corev1.ResourceMemory: resource.MustParse("0")},
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse("0"),
							corev1.ResourceMemory: resource.MustParse("0")},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: DataVolName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: GetPvcName(r.vmi.Name, false),
						},
					},
				},
				{
					Name: ScratchVolName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: GetPvcName(r.vmi.Name, true),
						},
					},
				},
			},
		},
	}

	r.addVolumeMounts(ip)

	if err := controllerutil.SetControllerReference(r.vmi, ip, r.scheme); err != nil {
		return nil, err
	}
	return ip, nil
}

func (r *ReconcileVirtualMachineImage) getSourceAndEndpoint() (source, endpoint string, err error) {
	if r.vmi.Spec.Source.HTTP == "" {
		return "", "", errors.NewBadRequest("Invalid spec.source. Must provide http source.")
	}
	return SourceHTTP, r.vmi.Spec.Source.HTTP, nil
}

func (r *ReconcileVirtualMachineImage) isCompleted(pod *corev1.Pod) bool {
	return len(pod.Status.ContainerStatuses) != 0 &&
		pod.Status.ContainerStatuses[0].State.Terminated != nil &&
		pod.Status.ContainerStatuses[0].State.Terminated.Reason == "Completed"
}

func (r *ReconcileVirtualMachineImage) addVolumeMounts(pod *corev1.Pod) {
	volumeMode := getVolumeMode(&r.vmi.Spec.PVC)
	if volumeMode == corev1.PersistentVolumeBlock {
		pod.Spec.Containers[0].VolumeDevices = []corev1.VolumeDevice{
			{Name: DataVolName, DevicePath: WriteBlockPath},
		}
		pod.Spec.SecurityContext = &corev1.PodSecurityContext{
			RunAsUser: &[]int64{0}[0],
		}
	} else {
		pod.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{Name: DataVolName, MountPath: ImporterDataDir},
		}
	}

	pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:      ScratchVolName,
		MountPath: ScratchDataDir,
	})
}

func getVolumeMode(pvcSpec *corev1.PersistentVolumeClaimSpec) corev1.PersistentVolumeMode {
	if pvcSpec.VolumeMode != nil {
		return *pvcSpec.VolumeMode
	}
	return corev1.PersistentVolumeBlock
}