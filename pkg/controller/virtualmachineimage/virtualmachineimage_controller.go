package virtualmachineimage

import (
	"context"
	"github.com/go-logr/logr"
	snapshotv1alpha1 "github.com/kubernetes-csi/external-snapshotter/pkg/apis/volumesnapshot/v1alpha1"
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
	err = c.Watch(&source.Kind{Type: &hc.VirtualMachineImage{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource PVCs and requeue the owner VirtualMachineImage
	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hc.VirtualMachineImage{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hc.VirtualMachineImage{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &snapshotv1alpha1.VolumeSnapshot{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &hc.VirtualMachineImage{},
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
	vmi    *hc.VirtualMachineImage
	log    logr.Logger
}

func (r *ReconcileVirtualMachineImage) fetchVmiFromName(namespacedName types.NamespacedName) error {
	vmi := &hc.VirtualMachineImage{}
	if err := r.client.Get(context.Background(), namespacedName, vmi); err != nil {
		return err
	}
	r.vmi = vmi.DeepCopy()
	return nil
}

func (r *ReconcileVirtualMachineImage) updateState(state hc.VirtualMachineImageState) error {
	r.log.Info("Update VirtualMachineImage state ", "from", r.vmi.Status.State, "to", state)
	r.vmi.Status.State = state
	return r.client.Status().Update(context.Background(), r.vmi)
}

func (r *ReconcileVirtualMachineImage) isState(state hc.VirtualMachineImageState) bool {
	return r.vmi.Status.State == state
}

// Reconcile reads that state of the cluster for a VirtualMachineImage object and makes changes based on the state read
func (r *ReconcileVirtualMachineImage) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log = log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	r.log.Info("Reconciling Start")

	if err := r.fetchVmiFromName(request.NamespacedName); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if err := r.getPvc(false); err != nil {
		if !errors.IsNotFound(err) {
			if err2 := r.updateState(hc.VirtualMachineImageStatePvcCreatingError); err2 != nil {
				return reconcile.Result{}, err2
			}
			return reconcile.Result{}, err
		}
		// PVC가 없으니 생성
		if err = r.createPvc(false); err != nil {
			if err2 := r.updateState(hc.VirtualMachineImageStatePvcCreatingError); err2 != nil {
				return reconcile.Result{}, err2
			}
			return reconcile.Result{}, err
		}
		// PVC 생성 완료, 다음 상태인 Importing으로 변경
		if err := r.updateState(hc.VirtualMachineImageStateImporting); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	if r.isState(hc.VirtualMachineImageStateImporting) {
		// 임포팅을 위해 스크래치 pvc가 필요
		if err := r.getPvc(true); err != nil {
			if !errors.IsNotFound(err) {
				if err2 := r.updateState(hc.VirtualMachineImageStatePvcCreatingError); err2 != nil {
					return reconcile.Result{}, err2
				}
				return reconcile.Result{}, err
			}
			if err = r.createPvc(true); err != nil {
				if err2 := r.updateState(hc.VirtualMachineImageStatePvcCreatingError); err2 != nil {
					return reconcile.Result{}, err2
				}
				return reconcile.Result{}, err
			}
		}
		// 임포터 파드 확인
		ip, err := r.getImporterPod()
		if err != nil {
			if !errors.IsNotFound(err) {
				if err2 := r.updateState(hc.VirtualMachineImageStateImportingError); err2 != nil {
					return reconcile.Result{}, err2
				}
				return reconcile.Result{}, err
			}
			// 임포터 파드가 없으니 생성
			_, err = r.createImporterPod()
			if err != nil {
				if errors.IsAlreadyExists(err) {
					return reconcile.Result{Requeue: true}, nil
				}
				if err2 := r.updateState(hc.VirtualMachineImageStateImportingError); err2 != nil {
					return reconcile.Result{}, err2
				}
				return reconcile.Result{}, err
			}
			// 임포터 파드가 끝날 때 까지 대기하기 위해 리큐 없이 리턴
			// 임포터 파드가 끝나면 다시 조정루프에 들어온다.
			return reconcile.Result{}, nil
		}
		// 임포터 파드가 끝났는지 확인
		if r.isCompleted(ip) {
			// 임포터파드 삭제
			if err := r.deleteImporterPod(); err != nil {
				if err2 := r.updateState(hc.VirtualMachineImageStateImportingError); err2 != nil {
					return reconcile.Result{}, err2
				}
				return reconcile.Result{}, err
			}
			// 스크래치 pvc 삭제
			if err := r.deletePvc(true); err != nil {
				if err2 := r.updateState(hc.VirtualMachineImageStateImportingError); err2 != nil {
					return reconcile.Result{}, err2
				}
				return reconcile.Result{}, err
			}
			// 다음 상태인 스냅샤팅 상태로 변경
			if err := r.updateState(hc.VirtualMachineImageStateSnapshotting); err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	// 스냅샷을 찍을 필요가 있을 때는 pvc를 새로 만들었을 때나 스냅샷이 지워졌을 때
	if r.isState(hc.VirtualMachineImageStateSnapshotting) || r.isState(hc.VirtualMachineImageStateAvailable) {
		if _, err := r.getSnapshot(); err != nil {
			if !errors.IsNotFound(err) {
				if err2 := r.updateState(hc.VirtualMachineImageStateSnapshottingError); err2 != nil {
					return reconcile.Result{}, err2
				}
				return reconcile.Result{}, err
			}
			// 스냅샷이 없으므로 생성
			if _, err = r.createSnapshot(); err != nil {
				if err2 := r.updateState(hc.VirtualMachineImageStateSnapshottingError); err2 != nil {
					return reconcile.Result{}, err2
				}
				return reconcile.Result{}, err
			}
			// 상태를 정상 상태로 변경
			if err := r.updateState(hc.VirtualMachineImageStateAvailable); err != nil {
				return reconcile.Result{}, err
			}
		}
		// 스냅샷이 있는 경우, 하지만 VirtualMachineImageStateSnapshotting 상태는 이전 스냅샷을 삭제하고 다시 만들어야 한다
		// PVC가 삭제된 경우 다시 임포트를 할 때 기존 pvc를 보고 있는 스냅샷이 있기 때문에 삭제한다.
		if r.isState(hc.VirtualMachineImageStateSnapshotting) {
			if err := r.deleteSnapshot(); err != nil {
				if err2 := r.updateState(hc.VirtualMachineImageStateSnapshottingError); err2 != nil {
					return reconcile.Result{}, err2
				}
				return reconcile.Result{}, err
			}
			// 삭제에 성공했으면 스냅샷을 다시 생성할 수 있도록 requeue 한다.
			return reconcile.Result{Requeue: true}, nil
		}
	}

	r.log.Info("Reconciling OK")
	return reconcile.Result{}, nil
}
