package cmd

import (
	"fmt"
	"github.com/anirudhAgniRedhat/openshift-metadata-manager/pkg/gcp"
	"log"
	"strings"

	"github.com/anirudhAgniRedhat/openshift-metadata-manager/pkg/aws"
	"github.com/anirudhAgniRedhat/openshift-metadata-manager/pkg/azure"
	infraType "github.com/anirudhAgniRedhat/openshift-metadata-manager/types"
	"github.com/spf13/cobra"
)

var (
	validateTags []string
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate resource tags",
	Long: `Validate tags on cloud infrastructure resources.
	
Checks if specified tags exist and have correct values on all resources.`,
	Example: `  # Validate tags for auto-detected platform
  openshift-metadata-manager validate --tags cost-center,environment
	
  # Validate specific tags for AWS
  openshift-metadata-manager validate --platform aws --tags project-id,owner`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Starting tag validation...")

		if len(validateTags) == 0 {
			log.Fatal("No tags specified for validation. Use --tags flag to specify tags.")
		}

		k8sClient := getK8sClient()
		cloudPlatform, err := getCloudPlatform(k8sClient)
		if err != nil {
			log.Fatalf("Error determining cloud platform: %v", err)
		}

		// Allow platform override
		if platform != "" {
			cloudPlatform = infraType.CloudPlatform(platform)
		}
		parsedTags, err := parseTags(validateTags)
		if err != nil {
			log.Fatalf("Error determining tags: %v", err)
		}

		fmt.Printf("Validating tags on %s platform\n", cloudPlatform)
		fmt.Printf("Tags to validate: %v\n", validateTags)

		switch cloudPlatform {
		case infraType.CloudPlatformAWS:
			err = aws.IsValidAWSTag(parsedTags)
		case infraType.CloudPlatformAzure:
			err = azure.IsValidAzureTag(parsedTags)
		case infraType.CloudPlatformGCP:
			err = gcp.IsValidGCPTag(parsedTags)
		default:
			log.Fatalf("Unsupported platform for validation: %s", cloudPlatform)
		}

		if err != nil {
			log.Fatalf("Validation failed: %v", err)
		}

		fmt.Println("✅ All specified valid")
	},
}

func init() {
	validateCmd.Flags().StringSliceVarP(&validateTags, "tags", "t", []string{},
		"Comma-separated list of tags to validate")
	RootCmd.AddCommand(validateCmd)
}

func parseTags(tagStrings []string) (map[string]string, error) {
	tags := make(map[string]string)

	for _, ts := range tagStrings {
		// Split into key=value pair
		parts := strings.SplitN(ts, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid tag format: %s. Must be KEY=VALUE", ts)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("tag key cannot be empty in: %s", ts)
		}

		tags[key] = value
	}

	return tags, nil
}
