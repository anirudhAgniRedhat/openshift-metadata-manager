package cmd

import (
	"context"
	"fmt"
	infraType "github.com/anirudhAgniRedhat/openshift-metadata-manager/types"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	kubeconfigPath string
	platform       string
	dryRun         bool
)

var RootCmd = &cobra.Command{
	Use:   "openshift-metadata-manager",
	Short: "OpenShift infrastructure resource tagger",
	Long: `A tool to manage tags on cloud infrastructure resources for OpenShift clusters.

Examples:
  # Sync tags for auto-detected platform
  openshift-metadata-manager sync
  
  # Sync tags for AWS with dry-run
  openshift-metadata-manager sync --platform aws --dry-run
  
  # Show version information
  openshift-metadata-manager version`,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
}

// Add a version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("openshift-metadata-manager v0.1.0")
	},
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&kubeconfigPath, "kubeconfig", "k", "", "Path to kubeconfig file")
	RootCmd.PersistentFlags().StringVarP(&platform, "platform", "p", "", "Override cloud platform (aws, azure, gcp)")
	RootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Dry run mode")
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getK8sClient() client.Client {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		fmt.Println("Error getting Kubernetes config:", err)
		os.Exit(1)
	}

	// Register OpenShift config API schema
	configv1.AddToScheme(scheme.Scheme)

	// Create a new controller-runtime client
	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		fmt.Println("Error creating Kubernetes client:", err)
		os.Exit(1)
	}
	return k8sClient
}

func getCloudPlatform(k8sClient client.Client) (infraType.CloudPlatform, error) {
	infra := &configv1.Infrastructure{}
	infraKey := client.ObjectKey{Name: "cluster"}

	// Fetch the Infrastructure resource
	if err := k8sClient.Get(context.Background(), infraKey, infra); err != nil {
		return "", fmt.Errorf("failed to get Infrastructure resource: %v", err)
	}
	return mapPlatformType(infra.Status.PlatformStatus.Type), nil
}

func mapPlatformType(platformType configv1.PlatformType) infraType.CloudPlatform {
	switch platformType {
	case configv1.AWSPlatformType:
		return infraType.CloudPlatformAWS
	case configv1.AzurePlatformType:
		return infraType.CloudPlatformAzure
	case configv1.GCPPlatformType:
		return infraType.CloudPlatformGCP
	case configv1.IBMCloudPlatformType:
		return infraType.CloudPlatformIBM
	case configv1.OpenStackPlatformType:
		return infraType.CloudPlatformOpenStack
	default:
		return infraType.CloudPlatformUnknown
	}
}
