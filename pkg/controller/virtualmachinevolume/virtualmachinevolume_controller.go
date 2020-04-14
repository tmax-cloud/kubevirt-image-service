package virtualmachinevolume

import (
	"context"
	goerrors "errors"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	hc "kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_virtualmachinevolume")

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
	// Create a new controller
	c, err := controller.New("virtualmachinevolume-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource VirtualMachineVolume
	err = c.Watch(&source.Kind{Type: &hc.VirtualMachineVolume{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource PVCs and requeue the owner VirtualMachineVolume
	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hc.VirtualMachineVolume{},
	})
	if err != nil {
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
	log    logr.Logger
}

// Reconcile reads that state of the cluster for a VirtualMachineVolume object and makes changes based on the state read
// and what is in the VirtualMachineVolume.Spec
func (r *ReconcileVirtualMachineVolume) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log = log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	r.log.Info("Reconciling VirtualMachineVolume")

	if err := r.fetchVolumeFromName(request.NamespacedName); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	ret, err := r.volumeReconcile()
	if err != nil {
		if err2 := r.updateState(hc.VirtualMachineVolumeStateError); err2 != nil {
			return reconcile.Result{}, err2
		}
		return ret, err
	}
	r.log.Info("Reconciling OK")

	return ret, nil
}

// fetchVolumeFromName fetches the VirtualMachineVolume instance
func (r *ReconcileVirtualMachineVolume) fetchVolumeFromName(namespacedName types.NamespacedName) error {
	v := &hc.VirtualMachineVolume{}
	if err := r.client.Get(context.Background(), namespacedName, v); err != nil {
		return err
	}
	r.volume = v.DeepCopy()
	return nil
}

func (r *ReconcileVirtualMachineVolume) volumeReconcile() (reconcile.Result, error) {
	// Get virtualMachineImage
	image, err := r.getImage()
	if err != nil {
		if errors.IsNotFound(err) {
			r.log.Info("VirtualMachineImage is not exists")
			return reconcile.Result{}, goerrors.New("VirtualMachineImage is not exists")
		}
		return reconcile.Result{}, err
	}

	// Check virtualMachineImage state is available
	if image.Status.State != hc.VirtualMachineImageStateAvailable {
		r.log.Info("VirtualMachineImage state is not available")
		return reconcile.Result{}, goerrors.New("VirtualMachineImage state is not available")
	}

	// Check pvc size
	originalSize := image.Spec.PVC.Resources.Requests[corev1.ResourceStorage]
	requestedSize := r.volume.Spec.Capacity[corev1.ResourceStorage]

	// Restored pvc size should not be smaller than original pvc size
	if originalSize.Value() > requestedSize.Value() {
		r.log.Info("VirtualMachineVolume size should be greater than or equal to VirtualMachineImage size", "VirtualMachineImage size is", originalSize, "Requested size is", requestedSize)
		return reconcile.Result{}, errors.NewBadRequest("VirtualMachineVolume size should be greater than or equal to VirtualMachineImage size")
	}

	// Get virtualMachineVolume pvc
	if _, err := r.getRestoredPvc(); err != nil {
		if !errors.IsNotFound(err) {
			return reconcile.Result{}, err
		}

		// VirtualMachineVolume pvc is not found. Creating a new one.
		if err := r.updateState(hc.VirtualMachineVolumeStateCreating); err != nil {
			return reconcile.Result{}, err
		}

		_, err := r.restorePvc(image)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Restoring successful. Change state to Available
		if err := r.updateState(hc.VirtualMachineVolumeStateAvailable); err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileVirtualMachineVolume) updateState(state hc.VirtualMachineVolumeState) error {
	r.log.Info("Update VirtualMachineVolume state", "from", r.volume.Status.State, "to", state)
	r.volume.Status.State = state
	err := r.client.Status().Update(context.Background(), r.volume)
	if err != nil {
		r.log.Info("Failed to update virtual machine volume state")
	}
	return err
}

func (r *ReconcileVirtualMachineVolume) getImage() (*hc.VirtualMachineImage, error) {
	image := &hc.VirtualMachineImage{}
	err := r.client.Get(context.Background(), types.NamespacedName{Name: r.volume.Spec.VirtualMachineImage.Name, Namespace: r.volume.Namespace}, image)
	if err != nil {
		return nil, err
	}
	return image, err
}
