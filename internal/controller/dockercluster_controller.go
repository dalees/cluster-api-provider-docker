/*
Copyright 2023.

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

package controller

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	//"github.com/capi-samples/cluster-api-provider-docker/pkg/container"
	infrav1 "github.com/dalees/cluster-api-provider-docker/api/v1alpha1"
	"github.com/dalees/cluster-api-provider-docker/pkg/container"
	"github.com/dalees/cluster-api-provider-docker/pkg/docker"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/patch"
)

// DockerClusterReconciler reconciles a DockerCluster object
type DockerClusterReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	ContainerRuntime container.Runtime
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=dockerclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=dockerclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=dockerclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DockerCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *DockerClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, rerr error) {
	logger := log.FromContext(ctx)
	logger.Info("Starting Reconcile on DockerCluster")
	defer func() {
		logger.Info("Reconcile complete on DockerCluster.")
	}()

	ctx = container.RuntimeInto(ctx, r.ContainerRuntime)

	// Get the instance of the DockerCluster from the request.
	dockerCluster := &infrav1.DockerCluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, dockerCluster); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get the owning type of the DockerCluster, which is the CAPI Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, dockerCluster.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if cluster == nil {
		logger.Info("Waiting for Cluster Controller to set OwnerRef on DockerCluster")
		return ctrl.Result{}, nil
	}
	// Now we have the cluster we can update the logger to include the cluster name in later log lines
	logger = logger.WithValues("cluster", klog.KObj(cluster))
	ctx = ctrl.LoggerInto(ctx, logger)

	// Support pausing of cluster updates
	if annotations.IsPaused(cluster, dockerCluster) {
		logger.Info("DockerCluster or owning Cluster is marked as paused, not reconciling")

		return ctrl.Result{}, nil
	}

	// Create a helper for managing a docker container hosting the loadbalancer.
	externalLoadBalancer, err := docker.NewLoadBalancer(ctx, cluster, dockerCluster)
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "failed to create helper for managing the externalLoadBalancer")
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(dockerCluster, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Always attempt to Patch the DockerCluster object and status after each reconciliation.
	defer func() {
		if err := patchHelper.Patch(ctx, dockerCluster); err != nil {
			logger.Error(err, "failed to patch DockerCluster")
			if rerr == nil {
				rerr = err
			}
		}
	}()

	// Handle deleted clusters
	if !dockerCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, dockerCluster, externalLoadBalancer)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, dockerCluster, externalLoadBalancer)
}

func (r *DockerClusterReconciler) reconcileNormal(ctx context.Context, dockerCluster *infrav1.DockerCluster, externalLoadBalancer *docker.LoadBalancer) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Starting Reconcile on normal Docker Cluster")
	defer func() {
		logger.Info("Reconcile normal complete")
	}()

	if !controllerutil.ContainsFinalizer(dockerCluster, infrav1.ClusterFinalizer) {
		controllerutil.AddFinalizer(dockerCluster, infrav1.ClusterFinalizer)

		return ctrl.Result{Requeue: true}, nil
	}

	// TODO: loadbalancer actual creation logic

	dockerCluster.Status.Ready = true

	return ctrl.Result{}, nil
}

func (r *DockerClusterReconciler) reconcileDelete(ctx context.Context, dockerCluster *infrav1.DockerCluster, externalLoadBalancer *docker.LoadBalancer) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Starting Reconcile on deleted Docker Cluster")

	defer func() {
		logger.Info("Reconcile delete complete")
	}()

	// TODO: loadbalancer actual delete

	// Cluster is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(dockerCluster, infrav1.ClusterFinalizer)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DockerClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.DockerCluster{}).
		Complete(r)
}
