package virtualmachineimage

import (
	"context"

	hypercloudv1alpha1 "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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

// Reconcile reads that state of the cluster for a VirtualMachineImage object and makes changes based on the state read
func (r *ReconcileVirtualMachineImage) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling VirtualMachineImage")

	// Fetch the VirtualMachineImage instance
	instance := &hypercloudv1alpha1.VirtualMachineImage{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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

	// Update state to Creating
	if instance.Status.State != hypercloudv1alpha1.VirtualMachineImageStateCreating {
		i := instance.DeepCopy()
		i.Status.State = hypercloudv1alpha1.VirtualMachineImageStateCreating
		if err := r.client.Status().Update(context.Background(), i); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
