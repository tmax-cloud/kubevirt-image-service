/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package virtualmachineimage

import (
	"context"
	goerrors "errors"
	"github.com/go-logr/logr"
	snapshotv1beta1 "github.com/kubernetes-csi/external-snapshotter/v2/pkg/apis/volumesnapshot/v1beta1"
	"github.com/tmax-cloud/kubevirt-image-service/controllers/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	hc "github.com/tmax-cloud/kubevirt-image-service/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileVirtualMachineImage reconciles a VirtualMachineImage object
type ReconcileVirtualMachineImage struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	vmi    *hc.VirtualMachineImage
}

// +kubebuilder:rbac:groups=hypercloud.tmaxanc.com,resources=virtualmachineimages,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hypercloud.tmaxanc.com,resources=virtualmachineimages/status,verbs=get;update;patch

// Reconcile reads that state of the cluster for a VirtualMachineImage object and makes changes based on the state read
func (r *ReconcileVirtualMachineImage) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("VirtualMachineImage", req.NamespacedName)

	cachedVmi := &hc.VirtualMachineImage{}
	if err := r.Client.Get(context.TODO(), req.NamespacedName, cachedVmi); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil // Deleted VMI. Return and don't requeue.
		}
		return reconcile.Result{}, err
	}
	r.vmi = cachedVmi.DeepCopy()

	syncAll := func() error {
		if err := r.validateVirtualMachineImageSpec(); err != nil {
			return err
		}
		// If the pvc doesn't exist, create a pvc and update vmim's status to creating
		if err := r.syncPvc(); err != nil {
			return err
		}
		// If the pvc import is not complete, create a importer pod
		// If the pvc import is complete, delete the importer pod and update imported value to true
		if err := r.syncImporterPod(); err != nil {
			return err
		}
		// If the pvc import is complete, create a snapshot and update vmim's status to available
		if err := r.syncSnapshot(); err != nil {
			return err
		}
		return nil
	}
	if err := syncAll(); err != nil {
		// TODO: Setup Error reason
		if err2 := r.updateStateWithReadyToUse(hc.VirtualMachineImageStateError, corev1.ConditionFalse, "SeeMessages", err.Error()); err2 != nil {
			return reconcile.Result{}, err2
		}
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager creates new controller for VirtualMachineImage
func (r *ReconcileVirtualMachineImage) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hc.VirtualMachineImage{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Pod{}).
		Owns(&snapshotv1beta1.VolumeSnapshot{}).
		Complete(r)
}

// updateStateWithReadyToUse updates readyToUse and State. Other Status fields are not affected. vmi must be DeepCopy to avoid polluting the cache.
func (r *ReconcileVirtualMachineImage) updateStateWithReadyToUse(state hc.VirtualMachineImageState, readyToUseStatus corev1.ConditionStatus,
	reason, message string) error {
	r.vmi.Status.Conditions = util.SetConditionByType(r.vmi.Status.Conditions, hc.ConditionReadyToUse, readyToUseStatus, reason, message)
	r.vmi.Status.State = state
	return r.Client.Status().Update(context.TODO(), r.vmi)
}

func (r *ReconcileVirtualMachineImage) validateVirtualMachineImageSpec() error {
	if r.vmi.Spec.PVC.VolumeMode == nil || *r.vmi.Spec.PVC.VolumeMode != corev1.PersistentVolumeBlock {
		return goerrors.New("VolumeMode in pvc is invalid. Only 'Block' can be used")
	}
	_, found := r.vmi.Spec.PVC.Resources.Requests[corev1.ResourceStorage]
	if !found {
		return goerrors.New("storage request in pvc is missing")
	}
	if _, err := r.getSource(); err != nil {
		return err
	}
	return nil
}

func (r *ReconcileVirtualMachineImage) getSource() (string, error) {
	if r.vmi.Spec.Source.HTTP != "" && r.vmi.Spec.Source.HostPath != nil {
		return "", goerrors.New("only one source is possible")
	} else if r.vmi.Spec.Source.HTTP != "" {
		return SourceHTTP, nil
	} else if r.vmi.Spec.Source.HostPath != nil {
		return SourceHostPath, nil
	} else {
		return "", goerrors.New("vmim source is not set")
	}
}
