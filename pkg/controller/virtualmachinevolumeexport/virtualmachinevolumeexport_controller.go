package virtualmachinevolumeexport

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
)

// Add creates a new VirtualMachineVolumeExport Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileVirtualMachineVolumeExport{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("virtualmachinevolumeexport-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource VirtualMachineVolumeExport
	err = c.Watch(&source.Kind{Type: &hc.VirtualMachineVolumeExport{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource PVCs and requeue the owner VirtualMachineVolumeExport
	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hc.VirtualMachineVolumeExport{},
	})
	if err != nil {
		return err
	}
	// Watch for changes to secondary resource Pods and requeue the owner VirtualMachineVolumeExport
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hc.VirtualMachineVolumeExport{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileVirtualMachineVolumeExport implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileVirtualMachineVolumeExport{}

// ReconcileVirtualMachineVolumeExport reconciles a VirtualMachineVolumeExport object
type ReconcileVirtualMachineVolumeExport struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	scheme    *runtime.Scheme
	vmvExport *hc.VirtualMachineVolumeExport
}

// Reconcile reads that state of the cluster for a VirtualMachineVolumeExport object and makes changes based on the state read
// and what is in the VirtualMachineVolumeExport.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileVirtualMachineVolumeExport) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	klog.Infof("Start sync VirtualMachineVolumeExport %s", request.NamespacedName)
	defer func() {
		klog.Infof("End sync VirtualMachineVolumeExport %s", request.NamespacedName)
	}()

	cachedVmvExport := &hc.VirtualMachineVolumeExport{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, cachedVmvExport); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil // Deleted VMI. Return and don't requeue.
		}
		return reconcile.Result{}, err
	}
	r.vmvExport = cachedVmvExport.DeepCopy()
	syncAll := func() error {
		// check if virtual machine volume to export is available
		if err := r.validateVirtualMachineVolume(); err != nil {
			if err2 := r.updateStateWithReadyToUse(hc.VirtualMachineVolumeExportStatePending, corev1.ConditionFalse, "VmvExportIsInPending", err.Error()); err2 != nil {
				return err2
			}
			return err
		}
		// if there is no pvc, update state to creating and create pvc
		if err := r.syncExportPvc(); err != nil {
			return err
		}
		// if pvc export is not completed, create exporter pod if it not exist
		// if there is an exporter pod and state is complete, update completed to yes and delete it
		if err := r.syncExporterPod(); err != nil {
			return err
		}
		// if destination is local, create local pod and update readytouse to true if it not exist
		if destination := r.getDestination(); destination == ExporterDestinationLocal {
			if err := r.syncLocalPod(); err != nil {
				return err
			}
		}
		return nil
	}
	if err := syncAll(); err != nil {
		if r.vmvExport.Status.State == hc.VirtualMachineVolumeExportStatePending {
			return reconcile.Result{RequeueAfter: hc.VirtualMachineVolumeExportReconcileAgain}, nil
		}
		if err2 := r.updateStateWithReadyToUse(hc.VirtualMachineVolumeExportStateError, corev1.ConditionFalse, "VmvExportIsInError", err.Error()); err2 != nil {
			return reconcile.Result{}, err2
		}
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileVirtualMachineVolumeExport) getDestination() string {
	var destination string
	if r.vmvExport.Spec.Destination.Local != nil {
		destination = ExporterDestinationLocal
	}
	return destination
}

// updateStateWithReadyToUse updates conditions and state. Other Status fields are not affected. vmvExport must be DeepCopy to avoid polluting the cache.
func (r *ReconcileVirtualMachineVolumeExport) updateStateWithReadyToUse(state hc.VirtualMachineVolumeExportState, readyToUseStatus corev1.ConditionStatus,
	reason, message string) error {
	r.vmvExport.Status.Conditions = util.SetConditionByType(r.vmvExport.Status.Conditions, hc.VirtualMachineVolumeExportConditionReadyToUse, readyToUseStatus, reason, message)
	r.vmvExport.Status.State = state
	return r.client.Status().Update(context.TODO(), r.vmvExport)
}

func (r *ReconcileVirtualMachineVolumeExport) validateVirtualMachineVolume() error {
	klog.Infof("Validate vmv for vmvExport %s", r.vmvExport.Name)
	// check if vmv is exist
	vmVolume := &hc.VirtualMachineVolume{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: r.vmvExport.Namespace, Name: r.vmvExport.Spec.VirtualMachineVolume.Name}, vmVolume)
	if err != nil {
		return goerrors.New("fail to get virtual machine volume")
	}
	// check if vmv is available
	found, cond := util.GetConditionByType(vmVolume.Status.Conditions, hc.VirtualMachineVolumeConditionReadyToUse)
	if !found || cond.Status == corev1.ConditionUnknown {
		return goerrors.New("VirtualMachineVolume state is not determined yet")
	} else if cond.Status == corev1.ConditionFalse {
		return goerrors.New("VirtualMachineVolume state is not in the condition")
	}
	return nil
}
