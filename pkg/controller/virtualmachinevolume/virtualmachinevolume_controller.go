package virtualmachinevolume

import (
	"context"
	goerrors "errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	"kubevirt-image-service/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

// ReconcileInterval is an time to reconcile again when in Pending State
const ReconcileInterval = 1 * time.Second

// Add creates a new VirtualMachineVolume Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileVirtualMachineVolume{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("virtualmachinevolume-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	if err := c.Watch(&source.Kind{Type: &hc.VirtualMachineVolume{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}
	if err := c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}},
		&handler.EnqueueRequestForOwner{IsController: true, OwnerType: &hc.VirtualMachineVolume{}}); err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileVirtualMachineVolume implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileVirtualMachineVolume{}

// ReconcileVirtualMachineVolume reconciles a VirtualMachineVolume object
type ReconcileVirtualMachineVolume struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	volume *hc.VirtualMachineVolume
}

// Reconcile reads that state of the cluster for a VirtualMachineVolume object and makes changes based on the state read
// and what is in the VirtualMachineVolume.Spec
func (r *ReconcileVirtualMachineVolume) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	klog.Infof("Start sync VirtualMachineVolume %s", request.NamespacedName)
	defer func() {
		klog.Infof("End sync VirtualMachineVolume %s", request.NamespacedName)
	}()

	cachedVolume := &hc.VirtualMachineVolume{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, cachedVolume); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil // Deleted Volume. Return and don't requeue.
		}
		return reconcile.Result{}, err
	}
	r.volume = cachedVolume.DeepCopy()

	if err := r.validateVolumeSpec(); err != nil {
		if err2 := r.updateStateWithReadyToUse(hc.VirtualMachineVolumeStatePending, corev1.ConditionFalse, "VmVolumeIsInPending", err.Error()); err2 != nil {
			return reconcile.Result{}, err2
		}
		return reconcile.Result{RequeueAfter: ReconcileInterval}, nil
	}

	if err := r.syncVolumePvc(); err != nil {
		if err2 := r.updateStateWithReadyToUse(hc.VirtualMachineVolumeStateError, corev1.ConditionFalse, "VmVolumeIsInError", err.Error()); err2 != nil {
			return reconcile.Result{}, err2
		}
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileVirtualMachineVolume) validateVolumeSpec() error {
	// Validate VirtualMachineImageName
	image := &hc.VirtualMachineImage{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: r.volume.Spec.VirtualMachineImage.Name, Namespace: r.volume.Namespace}, image); err != nil {
		if errors.IsNotFound(err) {
			return goerrors.New("VirtualMachineImage is not exists")
		}
		return err
	}

	// Check virtualMachineImage state is available
	found, cond := util.GetConditionByType(image.Status.Conditions, hc.ConditionReadyToUse)
	if !found || cond.Status != corev1.ConditionTrue {
		klog.Info("VirtualMachineImage state is not available")
		return goerrors.New("VirtualMachineImage state is not available")
	}

	return nil
}

// updateStateWithReadyToUse updates readyToUse condition type and State.
func (r *ReconcileVirtualMachineVolume) updateStateWithReadyToUse(state hc.VirtualMachineVolumeState, readyToUseStatus corev1.ConditionStatus,
	reason, message string) error {
	r.volume.Status.Conditions = util.SetConditionByType(r.volume.Status.Conditions, hc.VirtualMachineVolumeConditionReadyToUse, readyToUseStatus, reason, message)
	r.volume.Status.State = state
	return r.client.Status().Update(context.TODO(), r.volume)
}
