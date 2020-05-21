package virtualmachineimage

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

const (
	// DataVolName provides a const to use for creating volumes in pod specs
	DataVolName = "data-vol"
	// ScratchVolName provides a const to use for creating scratch pvc volumes in pod specs
	ScratchVolName = "scratch-vol"
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

func (r *ReconcileVirtualMachineImage) syncImporterPod() error {
	imported, found, err := r.isPvcImported()
	if err != nil {
		return err
	} else if !found {
		klog.Warningf("syncImporterPod without pvc in vmi %s", r.vmi.Name)
		return nil
	}

	importerPod := &corev1.Pod{}
	err = r.client.Get(context.Background(), types.NamespacedName{Namespace: r.vmi.Namespace, Name: GetImporterPodNameFromVmiName(r.vmi.Name)}, importerPod)
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
		if err := r.client.Delete(context.TODO(), importerPod); err != nil && !errors.IsNotFound(err) {
			return err
		}
	} else if !imported && !existsImporterPod {
		// 임포팅을 해야 하므로 임포터파드를 만든다
		klog.Infof("syncImporterPod create new importerPod for vmi %s", r.vmi.Name)
		newPod, err := newImporterPod(r.vmi, r.scheme)
		if err != nil {
			return err
		}
		if err := r.client.Create(context.TODO(), newPod); err != nil && !errors.IsAlreadyExists(err) {
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

func newImporterPod(vmi *hc.VirtualMachineImage, scheme *runtime.Scheme) (*corev1.Pod, error) {
	source, endpoint, err := getSourceAndEndpoint(vmi)
	if err != nil {
		return nil, err
	}

	pvcSize, found := vmi.Spec.PVC.Resources.Requests[corev1.ResourceStorage]
	if !found {
		return nil, errors.NewBadRequest("storage request in pvc is missing")
	}

	ip := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetImporterPodNameFromVmiName(vmi.Name),
			Namespace: vmi.Namespace,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyOnFailure,
			Containers: []corev1.Container{
				{
					Name:  "importer",
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
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      ScratchVolName,
							MountPath: ScratchDataDir,
						},
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
							ClaimName: GetPvcNameFromVmiName(vmi.Name),
						},
					},
				},
				{
					Name: ScratchVolName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: getScratchPvcNameFromVmiName(vmi.Name),
						},
					},
				},
			},
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser: &[]int64{0}[0],
			},
		},
	}
	if err := controllerutil.SetControllerReference(vmi, ip, scheme); err != nil {
		return nil, err
	}
	return ip, nil
}

func getSourceAndEndpoint(vmi *hc.VirtualMachineImage) (source, endpoint string, err error) {
	if vmi.Spec.Source.HTTP == "" {
		return "", "", errors.NewBadRequest("Invalid spec.source. Must provide http source.")
	}
	return SourceHTTP, vmi.Spec.Source.HTTP, nil
}
