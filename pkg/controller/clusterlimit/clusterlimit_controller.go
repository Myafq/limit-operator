package clusterlimit

import (
	"context"
	"reflect"

	limitv1alpha1 "github.com/myafq/limit-operator/pkg/apis/limit/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_clusterlimit")

// Add creates a new ClusterLimit Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileClusterLimit{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clusterlimit-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ClusterLimit
	err = c.Watch(&source.Kind{Type: &limitv1alpha1.ClusterLimit{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	// Watch for changes in all Namespace objects
	err = c.Watch(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ClusterLimit
	err = c.Watch(&source.Kind{Type: &corev1.LimitRange{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &limitv1alpha1.ClusterLimit{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileClusterLimit implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileClusterLimit{}

// ReconcileClusterLimit reconciles a ClusterLimit object
type ReconcileClusterLimit struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ClusterLimit object and makes changes based on the state read
// and what is in the ClusterLimit.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileClusterLimit) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconciling ClusterLimit", "Triggered by", request)
	// Will trigger on every namespace, clusterlimits and child limitrange change and will reconcile all clusterlimits and all namespaces
	// Fetch the ClusterLimits instance
	allClusterLimits := &limitv1alpha1.ClusterLimitList{}
	err := r.client.List(context.TODO(), &client.ListOptions{}, allClusterLimits)
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

	for _, cl := range allClusterLimits.Items {
		matchedNamespaces := &corev1.NamespaceList{}
		err := r.client.List(context.TODO(), client.MatchingLabels(cl.Spec.NamespaceSelector.MatchLabels), matchedNamespaces)
		if err != nil {
			log.Error(err, "Failed to list Namespaces.")
			return reconcile.Result{}, err
		}
		// log.Info("Namespace search", "ClusterLimit", cl.Name, "Matched Namespaces", len(matchedNamespaces.Items))
		for _, ns := range matchedNamespaces.Items {
			if val, _ := ns.ObjectMeta.Annotations["limit.myafq.com/unlimited"]; val == "true" {
				// log.Info("Skipping namespace because of unlimited annotation.", "Namespace", ns.Name)
				continue
			}
			nlr := r.newLimitRange(&cl, ns.Name)
			limit := &corev1.LimitRange{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cl.Name, Namespace: ns.Name}, limit)
			if err != nil && errors.IsNotFound(err) {
				// Define a new Service object
				log.Info("Creating new namespace LimitRange.", "Namespace", ns.Name, "LimitRange", cl.Name)
				err = r.client.Create(context.TODO(), nlr)
				if err != nil {
					log.Error(err, "Failed to create new LimitRange.", "Namespace", nlr.Namespace, "Name", nlr.Name)
					return reconcile.Result{}, err
				}
			} else if err != nil {
				log.Error(err, "Failed to get LimitRange.")
				return reconcile.Result{}, err
			}
			if !reflect.DeepEqual(nlr.Spec, limit.Spec) {
				log.Info("Updating namespace LimitRange.", "Namespace", nlr.Namespace, "Name", nlr.Name)
				err = r.client.Update(context.TODO(), nlr)

				if err != nil {
					log.Error(err, "Failed to update LimitRange status.", "Namespace", nlr.Namespace, "Name", nlr.Name)
					return reconcile.Result{}, err
				}
			}

		}
		allLimits := &corev1.LimitRangeList{}
		errr := r.client.List(context.TODO(), &client.ListOptions{}, allLimits)
		if errr != nil {
			log.Error(err, "Failed to list LimitRanges.")
			return reconcile.Result{}, err
		}
		nsEnforced := []string{}
		for _, lim := range allLimits.Items {
			for _, owner := range lim.OwnerReferences {
				if owner.Kind == "ClusterLimit" && owner.Name == cl.Name {
					nsEnforced = append(nsEnforced, lim.Namespace)
				}
			}
		}
		// Update status.NamespacesEnforced if needed
		if !areSame(nsEnforced, cl.Status.NamespacesEnforced) {
			cl.Status.NamespacesEnforced = nsEnforced
			err := r.client.Status().Update(context.TODO(), &cl)
			if err != nil {
				log.Error(err, "Failed to update ClusterLimits status.")
				return reconcile.Result{}, err
			}
		}
	}
	return reconcile.Result{}, nil
}
func areSame(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, ae := range a {
		if !includes(ae, b) {
			return false
		}
	}
	for _, be := range b {
		if !includes(be, a) {
			return false
		}
	}
	return true
}

func includes(a string, b []string) bool {
	for _, be := range b {
		if a == be {
			return true
		}
	}
	return false
}

// newLimitRange returns Limit range with spec from ClusterLimit
func (r *ReconcileClusterLimit) newLimitRange(cl *limitv1alpha1.ClusterLimit, ns string) *corev1.LimitRange {
	labels := map[string]string{
		"clusterlimit": cl.Name,
	}
	lr := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cl.Name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: cl.Spec.LimitRange,
	}
	controllerutil.SetControllerReference(cl, lr, r.scheme)
	return lr

}
