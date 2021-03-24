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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/giantswarm/dns-operator-aws/pkg/cloud/scope"
	"github.com/giantswarm/dns-operator-aws/pkg/cloud/services/route53"

	capa "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
)

const (
	CAPIWatchFilterLabel                    = "cluster.x-k8s.io/watch-filter"
	CAPAReleaseComponent                    = "cluster-api-provider-aws"
	dnsFinalizerName                        = "dns-operator-aws.finalizers.giantswarm.io"
	DNSZoneReady         capi.ConditionType = "DNSZoneReady"
)

// AWSClusterReconciler reconciles a AWSCluster object
type AWSClusterReconciler struct {
	client.Client

	Log                         logr.Logger
	ManagementClusterARN        string
	ManagementClusterBaseDomain string
	WorkloadClusterARN          string
	WorkloadClusterBaseDomain   string
	Scheme                      *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=awsclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=awsclusters/status,verbs=get;update;patch

func (r *AWSClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("awscluster", req.NamespacedName)

	awsCluster := &capa.AWSCluster{}
	err := r.Get(ctx, req.NamespacedName, awsCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, awsCluster.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, nil
	}
	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return reconcile.Result{}, err
	}

	log = log.WithValues("cluster", awsCluster.Name)

	// Return early if the object or Cluster is paused.
	if annotations.IsPaused(cluster, awsCluster) {
		log.Info("AWSCluster or linked Cluster is marked as paused. Won't reconcile")
		return ctrl.Result{}, nil
	}

	// Create the workload cluster scope.
	clusterScope, err := scope.NewClusterScope(scope.ClusterScopeParams{
		ARN:        r.WorkloadClusterARN,
		BaseDomain: r.WorkloadClusterBaseDomain,
		Logger:     log,
		AWSCluster: awsCluster,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Create the management cluster scope.
	managementScope, err := scope.NewManagementClusterScope(scope.ManagementClusterScopeParams{
		ARN:        r.ManagementClusterARN,
		BaseDomain: r.ManagementClusterBaseDomain,
		Logger:     log,
		AWSCluster: awsCluster,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Handle deleted clusters
	if !awsCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, clusterScope, managementScope)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, clusterScope, managementScope)
}

func (r *AWSClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capa.AWSCluster{}).
		Complete(r)
}

func (r *AWSClusterReconciler) reconcileNormal(ctx context.Context, clusterScope *scope.ClusterScope, managementScope *scope.ManagementClusterScope) (reconcile.Result, error) {
	clusterScope.Info("Reconciling AWSCluster normal")

	awsCluster := clusterScope.AWSCluster
	// If the AWSCluster doesn't have the finalizer, add it.
	controllerutil.AddFinalizer(awsCluster, dnsFinalizerName)

	route53Service := route53.NewService(clusterScope, managementScope)
	if err := route53Service.ReconcileRoute53(); err != nil {
		clusterScope.Error(err, "error creating route53")
		return reconcile.Result{}, err
	}

	conditions.MarkTrue(awsCluster, DNSZoneReady)
	err := r.Client.Status().Update(ctx, awsCluster)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *AWSClusterReconciler) reconcileDelete(ctx context.Context, clusterScope *scope.ClusterScope, managementScope *scope.ManagementClusterScope) (reconcile.Result, error) {
	clusterScope.Info("Reconciling AWSCluster delete")

	route53Service := route53.NewService(clusterScope, managementScope)

	if err := route53Service.DeleteRoute53(); err != nil {
		clusterScope.Error(err, "error deleting route53")
		return reconcile.Result{}, err
	}

	// AWSCluster is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(clusterScope.AWSCluster, dnsFinalizerName)

	return reconcile.Result{}, nil
}
