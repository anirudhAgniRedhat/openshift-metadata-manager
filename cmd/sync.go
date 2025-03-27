package cmd

import (
	"context"
	"fmt"
	"github.com/anirudhAgniRedhat/openshift-metadata-manager/pkg/gcp"
	configv1 "github.com/openshift/api/config/v1"
	"log"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/anirudhAgniRedhat/openshift-metadata-manager/pkg/aws"
	"github.com/anirudhAgniRedhat/openshift-metadata-manager/pkg/azure"
	infraType "github.com/anirudhAgniRedhat/openshift-metadata-manager/types"
	"github.com/spf13/cobra"

	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/rest"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize resource tags",
	Long:  "Fetch and update tags for cloud platform resources",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Starting infrastructure resource synchronization...")

		k8sClient := getK8sClient()

		cloudPlatform, err := getCloudPlatform(k8sClient)
		if err != nil {
			log.Fatalf("Error determining cloud platform: %v", err)
		}

		// Allow platform override
		if cloudPlatform != "" {
			cloudPlatform = infraType.CloudPlatform(cloudPlatform)
		}

		switch cloudPlatform {
		case infraType.CloudPlatformAWS:
			fmt.Println("🔄 Processing AWS resources...")
			resources, err := aws.ListAWSResources()
			if err != nil {
				log.Fatalf("Error determining AWS resources: %v", err)
				return
			}
			for _, res := range resources {
				fmt.Printf("%s (%s): %s\n", res.ID, res.Type, res.Name)
			}
		case infraType.CloudPlatformAzure:
			fmt.Println("🔄 Processing Azure resources...")
			azure.ListAzureResources()
		case infraType.CloudPlatformGCP:
			fmt.Println("🔄 Processing GCP resources...")
			gcp.ListGCPResources()
		default:
			fmt.Println("❌ Unsupported cloud platform:", cloudPlatform)
			os.Exit(1)
		}

		fmt.Println("✅ Synchronization complete")
	},
}

func init() {
	RootCmd.AddCommand(syncCmd)
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
