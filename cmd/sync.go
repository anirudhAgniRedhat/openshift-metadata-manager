package cmd

import (
	"fmt"
	"github.com/anirudhAgniRedhat/openshift-metadata-manager/pkg/aws"
	"github.com/anirudhAgniRedhat/openshift-metadata-manager/pkg/azure"
	"github.com/anirudhAgniRedhat/openshift-metadata-manager/pkg/gcp"
	infraType "github.com/anirudhAgniRedhat/openshift-metadata-manager/types"
	"github.com/spf13/cobra"
	"log"
)

var (
	tagsToSync []string
	//dryRun     bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize metadata across cloud resources",
	Long: `Apply consistent metadata tags/labels to all cloud resources 
associated with an OpenShift cluster. Existing tags will be preserved.`,
	Example: `  # Sync tags for auto-detected platform
  openshift-metadata-manager sync --tags Owner=DevOps,Environment=Production
  
  # Dry run for AWS
  openshift-metadata-manager sync --platform aws --tags CostCenter=1234 --dry-run`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔄 Starting metadata synchronization...")

		// Parse and validate tags
		if len(tagsToSync) == 0 {
			log.Fatal("No tags specified for synchronization")
		}
		tagMap, err := parseTags(tagsToSync)
		if err != nil {
			log.Fatal(err)
		}

		// Detect platform
		k8sClient := getK8sClient()
		cloudPlatform, err := getCloudPlatform(k8sClient)
		if err != nil {
			log.Fatalf("Platform detection error: %v", err)
		}

		// Allow platform override
		if platform != "" {
			cloudPlatform = infraType.CloudPlatform(platform)
		}

		// Execute platform-specific sync
		switch cloudPlatform {
		case infraType.CloudPlatformAWS:
			syncAWSTags(tagMap)
		case infraType.CloudPlatformAzure:
			syncAzureTags(tagMap)
		case infraType.CloudPlatformGCP:
			syncGCPTags(tagMap)
		default:
			log.Fatalf("Metadata sync not supported for platform: %s", cloudPlatform)
		}

		fmt.Println("✅ Metadata synchronization completed")
	},
}

// Platform-specific sync implementations
func syncAWSTags(tags map[string]string) {
	fmt.Printf("🔄 Syncing %d tags to AWS resources\n", len(tags))

	resources, err := aws.ListAWSResources()
	if err != nil {
		log.Fatalf("Failed to list AWS resources: %v", err)
	}
	for _, res := range resources {
		fmt.Println("Resources to be updated:", res.Type)
		fmt.Println("ResourceID :", res.ID, res.Name)
	}

	if err := aws.UpdateResourceTags(resources, tags); err != nil {
		log.Printf("  ❌ Error updating tags: %v", err)
	} else {
		fmt.Println("  ✓ Tags updated successfully")
	}

	//for _, res := range resources {
	//	fmt.Printf("Processing %s (%s)\n", res.ID, res.Type)
	//
	//	updateNeeded := false
	//	for k, v := range tags {
	//		if current, exists := res.Tags[k]; !exists || current != v {
	//			updateNeeded = true
	//			break
	//		}
	//	}
	//
	//	if !updateNeeded {
	//		fmt.Println("  ✓ Tags already up-to-date")
	//		continue
	//	}
	//
	//	if dryRun {
	//		fmt.Println("  🔄 [Dry Run] Tag changes:")
	//		printTagDiff(res.Tags, tags)
	//		continue
	//	}
	//
	//	if err := aws.UpdateResourceTags(res, tags); err != nil {
	//		log.Printf("  ❌ Error updating tags: %v", err)
	//	} else {
	//		fmt.Println("  ✓ Tags updated successfully")
	//	}
	//}
}

func syncAzureTags(tags map[string]string) {
	fmt.Printf("🔄 Syncing %d tags to Azure resources\n", len(tags))

	resources, err := azure.ListAzureResources()
	if err != nil {
		log.Fatalf("Failed to list Azure resources: %v", err)
	}

	for _, res := range resources {
		fmt.Printf("Processing %s (%s)\n", res.ID, res.Type)

		newTags := mergeTags(res.Tags, tags)
		if dryRun {
			fmt.Println("  🔄 [Dry Run] Tag changes:")
			printTagDiff(res.Tags, newTags)
			continue
		}

		if err := azure.UpdateResourceTags(res, newTags); err != nil {
			log.Printf("  ❌ Error updating tags: %v", err)
		} else {
			fmt.Println("  ✓ Tags updated successfully")
		}
	}
}

func syncGCPTags(tags map[string]string) {
	fmt.Printf("🔄 Syncing %d labels to GCP resources\n", len(tags))

	resources, err := gcp.ListGCPResources()
	if err != nil {
		log.Fatalf("Failed to list GCP resources: %v", err)
	}

	for _, res := range resources {
		fmt.Printf("Processing %s (%s)\n", res.ID, res.Type)

		newLabels := mergeTags(res.Tags, tags)
		if dryRun {
			fmt.Println("  🔄 [Dry Run] Label changes:")
			printTagDiff(res.Tags, newLabels)
			continue
		}

		if err := gcp.UpdateResourceTags(res, newLabels); err != nil {
			log.Printf("  ❌ Error updating labels: %v", err)
		} else {
			fmt.Println("  ✓ Labels updated successfully")
		}
	}
}

// Helper functions
func mergeTags(existing, updates map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range existing {
		merged[k] = v
	}
	for k, v := range updates {
		merged[k] = v
	}
	return merged
}

func printTagDiff(oldTags, newTags map[string]string) {
	for k, newVal := range newTags {
		if oldVal, exists := oldTags[k]; exists {
			if oldVal != newVal {
				fmt.Printf("    ~ %-20s: %-30s → %s\n", k, oldVal, newVal)
			}
		} else {
			fmt.Printf("    + %-20s: %s\n", k, newVal)
		}
	}
	for k := range oldTags {
		if _, exists := newTags[k]; !exists {
			fmt.Printf("    - %s\n", k)
		}
	}
}

func init() {
	syncCmd.Flags().StringSliceVarP(&tagsToSync, "tags", "t", []string{},
		"Tags to sync in KEY=VALUE format (comma-separated)")
	syncCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false,
		"Preview changes without applying")
	syncCmd.MarkFlagRequired("tags")

	RootCmd.AddCommand(syncCmd)
}
