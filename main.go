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

package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	capa "sigs.k8s.io/cluster-api-provider-aws/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/dns-operator-aws/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = capa.AddToScheme(scheme)
	_ = capi.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var (
		associateResolverRules      bool
		resolverRulesOwnerAccountId string
		enableLeaderElection        bool
		metricsAddr                 string
		workloadClusterBaseDomain   string
		managementClusterBaseDomain string
		managementClusterName       string
		managementClusterNamespace  string
	)
	flag.BoolVar(&associateResolverRules, "associate-resolver-rules", false,
		"Enable associating all resolver rules in aws account to the workload cluster VPC "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")

	flag.StringVar(&workloadClusterBaseDomain, "workload-cluster-basedomain", "", "Domain for workload cluster, e.g. installation.eu-west-1.aws.domain.tld")
	flag.StringVar(&managementClusterBaseDomain, "management-cluster-basedomain", "", "Domain for management cluster, e.g. installation.eu-west-1.aws.domain.tld.")
	flag.StringVar(&managementClusterName, "management-cluster-name", "", "Management cluster CR name.")
	flag.StringVar(&managementClusterNamespace, "management-cluster-namespace", "", "Management cluster CR namespace.")
	flag.StringVar(&resolverRulesOwnerAccountId, "account-id", "", "AWS account id owner of the dns resolver rules that will be associated with the VPC.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "d43d4591.giantswarm.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.AWSClusterReconciler{
		Client:                      mgr.GetClient(),
		ResolverRulesOwnerAccountId: resolverRulesOwnerAccountId,
		AssociateResolverRules:      associateResolverRules,
		Log:                         ctrl.Log.WithName("controllers").WithName("AWSCluster"),
		ManagementClusterBaseDomain: managementClusterBaseDomain,
		ManagementClusterName:       managementClusterName,
		ManagementClusterNamespace:  managementClusterNamespace,
		WorkloadClusterBaseDomain:   workloadClusterBaseDomain,
		Scheme:                      mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AWSCluster")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
