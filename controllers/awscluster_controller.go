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
	"time"

	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/giantswarm/dns-operator-aws/pkg/cloud/scope"
	"github.com/giantswarm/dns-operator-aws/pkg/cloud/services/route53"
	"github.com/giantswarm/dns-operator-aws/pkg/key"

	capa "sigs.k8s.io/cluster-api-provider-aws/api/v1beta1"
)

// AWSClusterReconciler reconciles a AWSCluster object
type AWSClusterReconciler struct {
	client.Client

	ResolverRulesOwnerAccountId string
	AssociateResolverRules      bool
	Log                         logr.Logger
	ManagementClusterBaseDomain string
	ManagementClusterName       string
	ManagementClusterNamespace  string
	WorkloadClusterBaseDomain   string
	Scheme                      *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=awsclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=awsclusters/status,verbs=get;update;patch

func (r *AWSClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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
		return reconcile.Result{}, err
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

	// Fetch AWSClusterRole from the cluster.
	awsClusterRoleIdentity := &capa.AWSClusterRoleIdentity{}
	err = r.Get(ctx, client.ObjectKey{Name: awsCluster.Spec.IdentityRef.Name}, awsClusterRoleIdentity)
	if err != nil {
		return reconcile.Result{}, err
	}
	// Fetch bastion IP
	// bastion might not exist depending on cluster configuration so there can be empty string here
	var bastionIP string
	{
		addrType := "ExternalIP"
		// if the cluster is private, use the InternalIP instead of ExernalIP
		if awsCluster.Annotations[annotation.AWSVPCMode] == annotation.AWSVPCModePrivate {
			addrType = "InternalIP"
		}

		bastionMachineList := &capi.MachineList{}
		err = r.List(ctx, bastionMachineList, client.MatchingLabels{
			"cluster.x-k8s.io/cluster-name": cluster.Name,
			"cluster.x-k8s.io/role":         "bastion",
		},
		)

		if err != nil {
			return reconcile.Result{}, err
		}
		if len(bastionMachineList.Items) > 0 {
			for _, addr := range bastionMachineList.Items[0].Status.Addresses {
				if addr.Type == capi.MachineAddressType(addrType) {
					bastionIP = addr.Address
					break
				}
			}
		}
	}

	// Create the workload cluster scope.
	clusterScope, err := scope.NewClusterScope(scope.ClusterScopeParams{
		ARN:                         awsClusterRoleIdentity.Spec.RoleArn,
		AssociateResolverRules:      r.AssociateResolverRules,
		BaseDomain:                  r.WorkloadClusterBaseDomain,
		BastionIP:                   bastionIP,
		Logger:                      log,
		AWSCluster:                  awsCluster,
		ResolverRulesOwnerAccountId: r.ResolverRulesOwnerAccountId,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	var managementAWSCluster capa.AWSCluster
	err = r.Get(ctx, client.ObjectKey{Name: r.ManagementClusterName, Namespace: r.ManagementClusterNamespace}, &managementAWSCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Error(err, "failed to get AWSCluster CR for management cluster, exiting")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Fetch AWSClusterRole from the cluster.
	awsManagementClusterRoleIdentity := &capa.AWSClusterRoleIdentity{}
	err = r.Get(ctx, client.ObjectKey{Name: managementAWSCluster.Spec.IdentityRef.Name}, awsManagementClusterRoleIdentity)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Create the management cluster scope.
	managementScope, err := scope.NewManagementClusterScope(scope.ManagementClusterScopeParams{
		ARN:        awsManagementClusterRoleIdentity.Spec.RoleArn,
		BaseDomain: r.ManagementClusterBaseDomain,
		Logger:     log,
		AWSCluster: &managementAWSCluster,
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
	clusterScope.Logger().Info("Reconciling AWSCluster normal")

	awsCluster := clusterScope.AWSCluster

	if !controllerutil.ContainsFinalizer(awsCluster, key.DNSFinalizerName) {
		patchHelper, err := patch.NewHelper(awsCluster, r.Client)
		if err != nil {
			return ctrl.Result{}, err
		}
		controllerutil.AddFinalizer(awsCluster, key.DNSFinalizerName)
		err = patchHelper.Patch(ctx, awsCluster)
		if err != nil {
			clusterScope.Logger().Error(err, "failed to add finalizer on AWSCluster")
			return ctrl.Result{}, err
		}
		clusterScope.Logger().Info("successfully added finalizer to AWSCluster")
	}

	route53Service := route53.NewService(clusterScope, managementScope)
	if err := route53Service.ReconcileRoute53(); err != nil {
		clusterScope.Logger().Error(err, "error creating route53")
		return reconcile.Result{}, err
	}

	conditions.MarkTrue(awsCluster, key.DNSZoneReady)
	patchHelper, err := patch.NewHelper(awsCluster, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = patchHelper.Patch(ctx, awsCluster)
	if err != nil {
		clusterScope.Logger().Error(err, "failed to set DNSZoneReady condition")
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: time.Minute * 5,
	}, nil
}

func (r *AWSClusterReconciler) reconcileDelete(ctx context.Context, clusterScope *scope.ClusterScope, managementScope *scope.ManagementClusterScope) (reconcile.Result, error) {
	clusterScope.Logger().Info("Reconciling AWSCluster delete")

	route53Service := route53.NewService(clusterScope, managementScope)

	if err := route53Service.DeleteRoute53(); err != nil {
		clusterScope.Logger().Error(err, "error deleting route53")
		return reconcile.Result{}, err
	}

	clusterScope.Logger().Info("removing finalizer")
	awsCluster := &capa.AWSCluster{}
	err := r.Get(ctx, client.ObjectKey{Name: clusterScope.AWSCluster.Name, Namespace: clusterScope.AWSCluster.Namespace}, awsCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, err
	}

	// AWSCluster is deleted so remove the finalizer.
	if controllerutil.ContainsFinalizer(awsCluster, key.DNSFinalizerName) {
		patchHelper, err := patch.NewHelper(awsCluster, r.Client)
		if err != nil {
			return ctrl.Result{}, err
		}
		controllerutil.RemoveFinalizer(awsCluster, key.DNSFinalizerName)
		err = patchHelper.Patch(ctx, awsCluster)
		if err != nil {
			clusterScope.Logger().Error(err, "failed to remove finalizer from AWSCluster")
			return ctrl.Result{}, err
		}
		clusterScope.Logger().Info("successfully removed finalizer from AWSCluster")
	}

	return ctrl.Result{}, nil
}
