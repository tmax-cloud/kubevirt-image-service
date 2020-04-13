package virtualmachineimage

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	hypercloudv1alpha1 "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_virtualmachineimage")

// Add creates a new VirtualMachineImage Controller and adds it to the Manager. The Manager will set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileVirtualMachineImage{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("virtualmachineimage-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource VirtualMachineImage
	err = c.Watch(&source.Kind{Type: &hypercloudv1alpha1.VirtualMachineImage{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource PVCs and requeue the owner VirtualMachineImage
	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hypercloudv1alpha1.VirtualMachineImage{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileVirtualMachineImage implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileVirtualMachineImage{}

// ReconcileVirtualMachineImage reconciles a VirtualMachineImage object
type ReconcileVirtualMachineImage struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

func getPvcName(vmiName string) string {
	return vmiName + "-pvc"
}

func getPvcNamespacedName(vmi *hypercloudv1alpha1.VirtualMachineImage) types.NamespacedName {
	return types.NamespacedName{Namespace: vmi.Namespace, Name: getPvcName(vmi.Name)}
}

func getPvcObjectMeta(vmi *hypercloudv1alpha1.VirtualMachineImage) v1.ObjectMeta {
	return v1.ObjectMeta{
		Name:      getPvcName(vmi.Name),
		Namespace: vmi.Namespace,
	}
}

// Reconcile reads that state of the cluster for a VirtualMachineImage object and makes changes based on the state read
func (r *ReconcileVirtualMachineImage) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Start")

	// Fetch the VirtualMachineImage vmi
	vmi := &hypercloudv1alpha1.VirtualMachineImage{}
	err := r.client.Get(context.TODO(), request.NamespacedName, vmi)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Delete first if we need
	if isVmiToBeDeleted(vmi) {
		reqLogger.Info("Start finalize")
		return r.finalize(reqLogger, vmi)
	}

	// Add finalizer for this CR
	if err := r.addFinalizer(vmi); err != nil {
		reqLogger.Info("Add finalizer")
		return reconcile.Result{}, err
	}

	// Prevent shared cache pollution
	vmi = vmi.DeepCopy()

	// Create pvc if needed...
	pvcFound := &corev1.PersistentVolumeClaim{}
	if err := r.client.Get(context.Background(), getPvcNamespacedName(vmi), pvcFound); err != nil {
		reqLogger.Info("PVC not exists, Create a new one")
		// pvc not exists, create

		// Update state to PvcCreating
		if err := r.updateState(vmi, hypercloudv1alpha1.VirtualMachineImageStatePvcCreating); err != nil {
			reqLogger.Info("Failed to update state")
			return reconcile.Result{}, err
		}

		pvc := &corev1.PersistentVolumeClaim{
			TypeMeta: v1.TypeMeta{
				Kind:       "PersistentVolumeClaim",
				APIVersion: "v1",
			},
			ObjectMeta: getPvcObjectMeta(vmi),
			Spec:       vmi.Spec.PVC,
		}

		// Set vmi as the owner and controller
		if err := controllerutil.SetControllerReference(vmi, pvc, r.scheme); err != nil {
			reqLogger.Info("Failed set owner")
			return reconcile.Result{}, err
		}

		if err := r.client.Create(context.Background(), pvc); err != nil {
			if err := r.updateStateErrorWithReason(vmi, hypercloudv1alpha1.VirtualMachineImageReasonFailedCreatePvc); err != nil {
				return reconcile.Result{}, err
			}
			reqLogger.Info("Failed create")
			return reconcile.Result{}, err
		}

		// TODO: import...
	}

	if err := r.updateState(vmi, hypercloudv1alpha1.VirtualMachineImageStateAvailable); err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Reconciling OK")

	return reconcile.Result{}, nil
}

func (r *ReconcileVirtualMachineImage) updateState(vmi *hypercloudv1alpha1.VirtualMachineImage, state hypercloudv1alpha1.VirtualMachineImageState) error {
	vmi.Status.State = state
	return r.client.Status().Update(context.Background(), vmi)
}

func (r *ReconcileVirtualMachineImage) updateStateErrorWithReason(vmi *hypercloudv1alpha1.VirtualMachineImage, reason hypercloudv1alpha1.VirtualMachineImageReason) error {
	vmi.Status.State = hypercloudv1alpha1.VirtualMachineImageStateError
	vmi.Status.Reason = reason
	return r.client.Status().Update(context.Background(), vmi)
}
