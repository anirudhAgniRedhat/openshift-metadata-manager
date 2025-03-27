package main

import (
	"github.com/anirudhAgniRedhat/openshift-metadata-manager/cmd"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Important for cloud auth
)

func main() {
	cmd.Execute()
}

//package main
//
//import (
//	"context"
//	"fmt"
//	"github.com/anirudhAgniRedhat/openshift-metadata-manager/pkg/aws"
//	"github.com/anirudhAgniRedhat/openshift-metadata-manager/pkg/azure"
//	infraType "github.com/anirudhAgniRedhat/openshift-metadata-manager/types"
//	configv1 "github.com/openshift/api/config/v1"
//	_ "github.com/spf13/cobra"
//	"k8s.io/client-go/kubernetes/scheme"
//	_ "k8s.io/client-go/rest"
//	"log"
//	"os"
//	ctrl "sigs.k8s.io/controller-runtime"
//	"sigs.k8s.io/controller-runtime/pkg/client"
//)
//
//func getK8sClient() client.Client {
//	cfg, err := ctrl.GetConfig()
//	if err != nil {
//		fmt.Println("Error getting Kubernetes config:", err)
//		os.Exit(1)
//	}
//
//	// Register OpenShift config API schema
//	configv1.AddToScheme(scheme.Scheme)
//
//	// Create a new controller-runtime client
//	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
//	if err != nil {
//		fmt.Println("Error creating Kubernetes client:", err)
//		os.Exit(1)
//	}
//	return k8sClient
//}
//
//func getCloudPlatform(k8sClient client.Client) (infraType.CloudPlatform, error) {
//	infra := &configv1.Infrastructure{}
//	infraKey := client.ObjectKey{Name: "cluster"}
//
//	// Fetch the Infrastructure resource
//	if err := k8sClient.Get(context.Background(), infraKey, infra); err != nil {
//		return "", fmt.Errorf("failed to get Infrastructure resource: %v", err)
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
//	case configv1.IBMCloudPlatformType:
//		return infraType.CloudPlatformIBM
//	case configv1.OpenStackPlatformType:
//		return infraType.CloudPlatformOpenStack
//	default:
//		return infraType.CloudPlatformUnknown
//	}
//}
//
//func main() {
//	fmt.Println("🔍 Welcome to the infrastructure resource tagger for openshift clusters")
//
//	k8sClient := getK8sClient()
//
//	cloudPlatform, err := getCloudPlatform(k8sClient)
//	if err != nil {
//		fmt.Println("cannot fetch the cloud platform", err)
//		log.Fatal(err)
//	}
//	switch cloudPlatform {
//	case infraType.CloudPlatformAWS:
//		fmt.Println("Fetching and updating tags for the aws cloud platform resources")
//		aws.ListAWSResources()
//	case infraType.CloudPlatformAzure:
//		fmt.Println("Fetching and updating tags for the azure cloud platform resources")
//		azure.ListAzureResources()
//	case infraType.CloudPlatformGCP:
//		fmt.Println("Fetching and updating tags for the gcp cloud platform resources")
//	case infraType.CloudPlatformIBM:
//		fmt.Println("Fetching and updating tags for the ibm cloud platform resources")
//	case infraType.CloudPlatformOpenStack:
//		fmt.Println("Fetching and updating tags for the openstack cloud platform resources")
//	default:
//		fmt.Println("Unsupported cloud platform")
//	}
//
//	// Process AWS resources
//	//err := pkg.ListAWSResources()
//	//if err != nil {
//	//	log.Fatalf("❌ Error listing AWS resources: %v", err)
//	//}
//
//	fmt.Println("✅ Resource listing complete.")
//}
