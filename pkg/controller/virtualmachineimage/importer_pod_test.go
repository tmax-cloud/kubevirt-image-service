package virtualmachineimage

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("getImporterPod", func() {
	It("Should get the importer pod, if vmi has a importer pod", func() {
		r, importerPod := createFakeReconcileVmiWithImporterPod()
		importerPodFound, err := r.getImporterPod()

		Expect(err).ToNot(HaveOccurred())
		Expect(importerPodFound.Spec).To(Equal(importerPod.Spec))
	})

	It("Should not get the importer pod, if vmi has no importer pod", func() {
		r := createFakeReconcileVmi()
		_, err := r.getImporterPod()

		Expect(errors.IsNotFound(err)).To(BeTrue())
	})
})

var _ = Describe("createImporterPod", func() {
	It("Should success to create the importer pod", func() {
		expectedImporterPodSpec := *createImporterPodSpec()

		r := createFakeReconcileVmi()
		importerPodCreated, err := r.createImporterPod()
		Expect(err).ToNot(HaveOccurred())

		importerPodFound := &corev1.Pod{}
		err = r.client.Get(context.Background(), types.NamespacedName{Name: "testvmi-importer", Namespace: "default"}, importerPodFound)
		Expect(errors.IsNotFound(err)).To(BeFalse())

		Expect(importerPodCreated.Spec).To(Equal(expectedImporterPodSpec))
		Expect(importerPodFound.Spec).To(Equal(expectedImporterPodSpec))
	})
})

var _ = Describe("deleteImporterPod", func() {
	It("Should delete the importer pod, if vmi has a importer pod", func() {
		r, _ := createFakeReconcileVmiWithImporterPod()

		err := r.deleteImporterPod()
		Expect(err).ToNot(HaveOccurred())

		importerPodFound := &corev1.Pod{}
		err = r.client.Get(context.Background(), types.NamespacedName{Name: "testvmi-importer", Namespace: "default"}, importerPodFound)
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})

	It("Should not delete the importer pod, if vmi has no importer pod", func() {
		r := createFakeReconcileVmi()
		err := r.deleteImporterPod()

		Expect(errors.IsNotFound(err)).To(BeTrue())
	})
})

var _ = Describe("GetImporterPodName", func() {
	It("Should get the importerPodName", func() {
		expectedImporterPodName := "testvmi-importer"

		r := createFakeReconcileVmi()
		importerPodName := GetImporterPodName(r.vmi.Name)

		Expect(importerPodName).To(Equal(expectedImporterPodName))
	})
})

var _ = Describe("newImporterPod", func() {
	It("Should success to create the importer pod", func() {
		r := createFakeReconcileVmi()
		importerPod, err := r.newImporterPod()

		isController := true
		blockOwnerDeletion := true
		expectedImporterPod := &corev1.Pod{
			TypeMeta: v1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "testvmi-importer",
				Namespace: "default",
				OwnerReferences: []v1.OwnerReference{
					{
						APIVersion:         "hypercloud.tmaxanc.com/v1alpha1",
						Kind:               "VirtualMachineImage",
						Name:               "testvmi",
						UID:                "",
						Controller:         &isController,
						BlockOwnerDeletion: &blockOwnerDeletion,
					},
				},
			},
			Spec: *createImporterPodSpec(),
		}

		Expect(err).ToNot(HaveOccurred())
		Expect(importerPod).To(Equal(expectedImporterPod))
	})
})

func createFakeReconcileVmiWithImporterPod() (*ReconcileVirtualMachineImage, *corev1.Pod) {
	ip := &corev1.Pod{
		TypeMeta: v1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "testvmi-importer",
			Namespace: "default",
		},
		Spec: *createImporterPodSpec(),
	}
	return createFakeReconcileVmi(ip), ip
}

func createImporterPodSpec() *corev1.PodSpec {
	podSpec := &corev1.PodSpec{
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &[]int64{0}[0],
		},
		Containers: []corev1.Container{
			{
				Name:  "testvmi-importer",
				Image: ImportPodImage,
				Args:  []string{"-v=" + ImportPodVerbose},
				VolumeDevices: []corev1.VolumeDevice{
					{Name: DataVolName, DevicePath: WriteBlockPath},
				},
				VolumeMounts: []corev1.VolumeMount{
					{Name: ScratchVolName, MountPath: ScratchDataDir},
				},
				Env: []corev1.EnvVar{
					{Name: ImporterSource, Value: "http"},
					{Name: ImporterEndpoint, Value: "https://download.cirros-cloud.net/contrib/0.3.0/cirros-0.3.0-i386-disk.img"},
					{Name: ImporterContentType, Value: "kubevirt"},
					{Name: ImporterImageSize, Value: "3Gi"},
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
						ClaimName: "testvmi-pvc",
					},
				},
			},
			{
				Name: ScratchVolName,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: "testvmi-scratch-pvc",
					},
				},
			},
		},
	}
	return podSpec
}