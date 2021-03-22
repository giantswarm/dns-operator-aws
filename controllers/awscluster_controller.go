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

	"github.com/giantswarm/dns-operator-aws/pkg/cloud/scope"
	"github.com/giantswarm/dns-operator-aws/pkg/cloud/services/route53"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	capa "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
)

const (
	CAPIWatchFilterLabel = "cluster.x-k8s.io/watch-filter"
	capaReleaseComponent = "cluster-api-provider-aws"
	dnsFinalizerName     = "dns-operator-aws.finalizers.giantswarm.io"
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

	log = log.WithValues("cluster", awsCluster.Name)

	// Create the workload cluster scope.
	clusterScope, err := scope.NewClusterScope(scope.ClusterScopeParams{
		// "arn:aws:iam::180547736195:role/GiantSwarmAWSOperator",
		ARN: r.WorkloadClusterARN,
		// "gauss.eu-west-1.aws.gigantic.io"
		BaseDomain: r.WorkloadClusterBaseDomain,
		Logger:     log,
		AWSCluster: awsCluster,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Create the management cluster scope.
	managementScope, err := scope.NewManagementClusterScope(scope.ManagementClusterScopeParams{
		// "arn:aws:iam::822380749555:role/GiantSwarmAWSOperator",
		ARN: r.ManagementClusterARN,
		// "gauss.eu-west-1.aws.gigantic.io"
		BaseDomain: r.ManagementClusterBaseDomain,
		Logger:     log,
		AWSCluster: awsCluster,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Handle deleted clusters
	if !awsCluster.DeletionTimestamp.IsZero() {
		return reconcileDelete(clusterScope, managementScope)
	}

	// Handle non-deleted clusters
	return reconcileNormal(clusterScope, managementScope)
}

func (r *AWSClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capa.AWSCluster{}).
		Complete(r)
}

func reconcileNormal(clusterScope *scope.ClusterScope, managementScope *scope.ManagementClusterScope) (reconcile.Result, error) {
	clusterScope.Info("Reconciling AWSCluster normal")

	awsCluster := clusterScope.AWSCluster
	// If the AWSCluster doesn't have the finalizer, add it.
	controllerutil.AddFinalizer(awsCluster, dnsFinalizerName)

	route53Service := route53.NewService(clusterScope, managementScope)
	if err := route53Service.ReconcileRoute53(); err != nil {
		clusterScope.Error(err, "error creating route53")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func reconcileDelete(clusterScope *scope.ClusterScope, managementScope *scope.ManagementClusterScope) (reconcile.Result, error) {
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
