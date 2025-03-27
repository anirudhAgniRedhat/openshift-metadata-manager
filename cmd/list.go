package cmd

import (
	"fmt"
	"github.com/anirudhAgniRedhat/openshift-metadata-manager/pkg/azure"
	"github.com/anirudhAgniRedhat/openshift-metadata-manager/pkg/gcp"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/anirudhAgniRedhat/openshift-metadata-manager/pkg/aws"
	infraType "github.com/anirudhAgniRedhat/openshift-metadata-manager/types"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List cloud resources associated with the cluster",
	Long:  "Display infrastructure resources managed by the OpenShift cluster",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📋 Listing cluster resources...")

		// Get cluster platform
		k8sClient := getK8sClient()
		cloudPlatform, err := getCloudPlatform(k8sClient)
		if err != nil {
			log.Fatalf("Error determining cloud platform: %v", err)
		}

		var resources []infraType.CloudResource
		switch cloudPlatform {
		case infraType.CloudPlatformAWS:
			resources, err = aws.ListAWSResources()
			if err != nil {
				log.Fatalf("Failed to list AWS resources: %v", err)
			}

		case infraType.CloudPlatformAzure:
			// Placeholder for Azure implementation
			resources, err = azure.ListAzureResources()
			if err != nil {
				log.Fatalf("Failed to list Azure resources: %v", err)
			}
			//log.Fatalf("Not Implemented Yet")
		case infraType.CloudPlatformGCP:
			resources, err = gcp.ListGCPResources()
			if err != nil {
				log.Fatalf("Failed to list GCP resources: %v", err)
			}

		default:
			log.Fatalf("Unsupported platform: %s", cloudPlatform)
		}

		printResourceTable(resources)
	},
}

func printResourceTable(resources []infraType.CloudResource) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintln(w, "CLOUD\tTYPE\tID\tNAME\tTAGS")

	for _, res := range resources {
		// Format tags as comma-separated key=value pairs
		var tagPairs []string
		for k, v := range res.Tags {
			tagPairs = append(tagPairs, fmt.Sprintf("%s=%s", k, v))
		}
		tags := strings.Join(tagPairs, ", ")

		// Truncate long tags to 50 characters
		if len(tags) > 50 {
			tags = tags[:47] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			res.CloudProvider,
			res.Type,
			res.ID,
			res.Name,
			tags,
		)
	}
}

//func getK8sClient() client.Client {
//	cfg, err := ctrl.GetConfig()
//	if err != nil {
//		log.Fatalf("Error getting Kubernetes config: %v", err)
//	}
//
//	configv1.AddToScheme(scheme.Scheme)
//	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
//	if err != nil {
//		log.Fatalf("Error creating Kubernetes client: %v", err)
//	}
//	return k8sClient
//}
//
//func getCloudPlatform(k8sClient client.Client) (infraType.CloudPlatform, error) {
//	infra := &configv1.Infrastructure{}
//	if err := k8sClient.Get(context.Background(), client.ObjectKey{Name: "cluster"}, infra); err != nil {
//		return "", fmt.Errorf("failed to get Infrastructure resource: %w", err)
//	}
//	return mapPlatformType(infra.Status.PlatformStatus.Type), nil
//}
//
//func mapPlatformType(platformType configv1.PlatformType) infraType.CloudPlatform {
//	switch platformType {
//	case configv1.AWSPlatformType:
//		return infraType.CloudPlatformAWS
//	case configv1.AzurePlatformType:
//		return infraType.CloudPlatformAzure
//	case configv1.GCPPlatformType:
//		return infraType.CloudPlatformGCP
//	default:
//		return infraType.CloudPlatformUnknown
//	}
//}

func init() {
	RootCmd.AddCommand(listCmd)
}
