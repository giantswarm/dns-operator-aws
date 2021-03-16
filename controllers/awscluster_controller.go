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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

  capa "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"

)

const (
	CAPIWatchFilterLabel = "cluster.x-k8s.io/watch-filter"
	capiReleaseComponent = "cluster-api-core"
	cacpReleaseComponent = "cluster-api-control-plane"
	capaReleaseComponent = "cluster-api-provider-aws"
	capzReleaseComponent = "cluster-api-provider-azure"
)

// AWSClusterReconciler reconciles a AWSCluster object
type AWSClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=awsclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=awsclusters/status,verbs=get;update;patch

func (r *AWSClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("awscluster", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *AWSClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructureclusterxk8siov1alpha3.AWSCluster{}).
		Complete(r)
}
