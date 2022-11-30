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

	capa "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

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
		accountID                   string
		enableLeaderElection        bool
		metricsAddr                 string
		workloadClusterBaseDomain   string
		managementClusterARN        string
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
	flag.StringVar(&managementClusterARN, "management-cluster-arn", "", "Assumed role name for management cluster DNS zone delegation operation.")
	flag.StringVar(&managementClusterBaseDomain, "management-cluster-basedomain", "", "Domain for management cluster, e.g. installation.eu-west-1.aws.domain.tld.")
	flag.StringVar(&managementClusterName, "management-cluster-name", "", "Management cluster CR name.")
	flag.StringVar(&managementClusterNamespace, "management-cluster-namespace", "", "Management cluster CR namespace.")
	flag.StringVar(&accountID, "account-id", "", "id of the cluster account")
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
		AccountID:                   accountID,
		AssociateResolverRules:      associateResolverRules,
		Log:                         ctrl.Log.WithName("controllers").WithName("AWSCluster"),
		ManagementClusterARN:        managementClusterARN,
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
