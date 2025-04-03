package azure

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"

	"log"
	"os"
	_ "strings"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	//"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	//"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"

	infraType "github.com/anirudhAgniRedhat/openshift-metadata-manager/types"
)

var (
	ClusterTagKey   = "kubernetes.io/cluster/%s"
	ClusterTagValue = "owned"
)

func ListAzureResources() ([]infraType.CloudResource, error) {
	ctx := context.TODO()
	var resources []infraType.CloudResource

	k8sClient := getK8sClient()
	resourceGroup, clusterName, err := getClusterResourceGroup(k8sClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info: %w", err)
	}
	ClusterTagKey = fmt.Sprintf(ClusterTagKey, clusterName)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("azure authentication failed: %w", err)
	}

	// List all resources
	if vmRes, err := listVirtualMachines(ctx, cred, resourceGroup); err == nil {
		resources = append(resources, vmRes...)
	}
	if diskRes, err := listDisks(ctx, cred, resourceGroup); err == nil {
		resources = append(resources, diskRes...)
	}
	if netRes, err := listVirtualNetworks(ctx, cred, resourceGroup); err == nil {
		resources = append(resources, netRes...)
	}
	if lbRes, err := listLoadBalancers(ctx, cred, resourceGroup); err == nil {
		resources = append(resources, lbRes...)
	}
	if ipRes, err := listPublicIPs(ctx, cred, resourceGroup); err == nil {
		resources = append(resources, ipRes...)
	}
	if saRes, err := listStorageAccounts(ctx, cred, resourceGroup); err == nil {
		resources = append(resources, saRes...)
	}
	if subnetRes, err := listSubnets(ctx, cred, resourceGroup); err == nil {
		resources = append(resources, subnetRes...)
	}
	if nsgRes, err := listNetworkSecurityGroups(ctx, cred, resourceGroup); err == nil {
		resources = append(resources, nsgRes...)
	}

	return resources, nil
}

func listVirtualMachines(ctx context.Context, cred *azidentity.DefaultAzureCredential, resourceGroup string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource
	client, err := armcompute.NewVirtualMachinesClient(os.Getenv("AZURE_SUBSCRIPTION_ID"), cred, nil)
	if err != nil {
		return nil, err
	}

	pager := client.NewListPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing VMs: %w", err)
		}

		for _, vm := range page.Value {
			resources = append(resources, infraType.CloudResource{
				CloudProvider: infraType.CloudPlatformAzure,
				Type:          infraType.CloudResourceTypeAzureVM,
				ID:            *vm.ID,
				Name:          *vm.Name,
				Tags:          convertTags(vm.Tags),
			})
		}
	}
	return resources, nil
}

func listDisks(ctx context.Context, cred *azidentity.DefaultAzureCredential, resourceGroup string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource
	client, err := armcompute.NewDisksClient(os.Getenv("AZURE_SUBSCRIPTION_ID"), cred, nil)
	if err != nil {
		return nil, err
	}

	pager := client.NewListByResourceGroupPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing disks: %w", err)
		}

		for _, disk := range page.Value {
			resources = append(resources, infraType.CloudResource{
				CloudProvider: infraType.CloudPlatformAzure,
				Type:          infraType.CloudResourceTypeAzureManagedDisk,
				ID:            *disk.ID,
				Name:          *disk.Name,
				Tags:          convertTags(disk.Tags),
			})
		}
	}
	return resources, nil
}

func listVirtualNetworks(ctx context.Context, cred *azidentity.DefaultAzureCredential, resourceGroup string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource
	client, err := armnetwork.NewVirtualNetworksClient(os.Getenv("AZURE_SUBSCRIPTION_ID"), cred, nil)
	if err != nil {
		return nil, err
	}

	pager := client.NewListPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing VNets: %w", err)
		}

		for _, vnet := range page.Value {
			resources = append(resources, infraType.CloudResource{
				CloudProvider: infraType.CloudPlatformAzure,
				Type:          infraType.CloudResourceTypeAzureVirtualNetwork,
				ID:            *vnet.ID,
				Name:          *vnet.Name,
				Tags:          convertTags(vnet.Tags),
			})
		}
	}
	return resources, nil
}

func listLoadBalancers(ctx context.Context, cred *azidentity.DefaultAzureCredential, resourceGroup string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource
	client, err := armnetwork.NewLoadBalancersClient(os.Getenv("AZURE_SUBSCRIPTION_ID"), cred, nil)
	if err != nil {
		return nil, err
	}

	pager := client.NewListPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing LBs: %w", err)
		}

		for _, lb := range page.Value {
			resources = append(resources, infraType.CloudResource{
				CloudProvider: infraType.CloudPlatformAzure,
				Type:          infraType.CloudResourceTypeAzureLoadBalancer,
				ID:            *lb.ID,
				Name:          *lb.Name,
				Tags:          convertTags(lb.Tags),
			})
		}
	}
	return resources, nil
}

func listPublicIPs(ctx context.Context, cred *azidentity.DefaultAzureCredential, resourceGroup string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource
	client, err := armnetwork.NewPublicIPAddressesClient(os.Getenv("AZURE_SUBSCRIPTION_ID"), cred, nil)
	if err != nil {
		return nil, err
	}

	pager := client.NewListPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing Public IPs: %w", err)
		}

		for _, ip := range page.Value {
			resources = append(resources, infraType.CloudResource{
				CloudProvider: infraType.CloudPlatformAzure,
				Type:          infraType.CloudResourceTypeAzurePublicIP,
				ID:            *ip.ID,
				Name:          *ip.Name,
				Tags:          convertTags(ip.Tags),
			})
		}
	}
	return resources, nil
}

func listStorageAccounts(ctx context.Context, cred *azidentity.DefaultAzureCredential, resourceGroup string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource
	client, err := armstorage.NewAccountsClient(os.Getenv("AZURE_SUBSCRIPTION_ID"), cred, nil)
	if err != nil {
		return nil, err
	}

	pager := client.NewListByResourceGroupPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing Storage Accounts: %w", err)
		}

		for _, sa := range page.Value {
			resources = append(resources, infraType.CloudResource{
				CloudProvider: infraType.CloudPlatformAzure,
				Type:          infraType.CloudResourceTypeAzureStorageAccount,
				ID:            *sa.ID,
				Name:          *sa.Name,
				Tags:          convertTags(sa.Tags),
			})
		}
	}
	return resources, nil
}

func listSubnets(ctx context.Context, cred *azidentity.DefaultAzureCredential, resourceGroup string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource
	client, err := armnetwork.NewSubnetsClient(os.Getenv("AZURE_SUBSCRIPTION_ID"), cred, nil)
	if err != nil {
		return nil, err
	}

	vnetClient, _ := armnetwork.NewVirtualNetworksClient(os.Getenv("AZURE_SUBSCRIPTION_ID"), cred, nil)
	vnetPager := vnetClient.NewListPager(resourceGroup, nil)

	for vnetPager.More() {
		vnetPage, err := vnetPager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing VNets: %w", err)
		}

		for _, vnet := range vnetPage.Value {
			subnetPager := client.NewListPager(resourceGroup, *vnet.Name, nil)
			for subnetPager.More() {
				subnetPage, err := subnetPager.NextPage(ctx)
				if err != nil {
					return nil, fmt.Errorf("error listing subnets: %w", err)
				}

				for _, subnet := range subnetPage.Value {
					resources = append(resources, infraType.CloudResource{
						CloudProvider: infraType.CloudPlatformAzure,
						Type:          infraType.CloudResourceTypeAzureSubnet,
						ID:            *subnet.ID,
						Name:          *subnet.Name,
						//Tags:          convertTags(subnet.Etag),
					})
				}
			}
		}
	}
	return resources, nil
}

func listNetworkSecurityGroups(ctx context.Context, cred *azidentity.DefaultAzureCredential, resourceGroup string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource
	client, err := armnetwork.NewSecurityGroupsClient(os.Getenv("AZURE_SUBSCRIPTION_ID"), cred, nil)
	if err != nil {
		return nil, err
	}

	pager := client.NewListPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing NSGs: %w", err)
		}

		for _, nsg := range page.Value {
			resources = append(resources, infraType.CloudResource{
				CloudProvider: infraType.CloudPlatformAzure,
				Type:          infraType.CloudResourceTypeAzureNetworkSecurityGroup,
				ID:            *nsg.ID,
				Name:          *nsg.Name,
				Tags:          convertTags(nsg.Tags),
			})
		}
	}
	return resources, nil
}

func convertTags(azureTags map[string]*string) map[string]string {
	tags := make(map[string]string)
	for k, v := range azureTags {
		if v != nil {
			tags[k] = *v
		}
	}
	return tags
}

func getK8sClient() client.Client {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		log.Fatalf("Error getting Kubernetes config: %v", err)
	}

	configv1.AddToScheme(scheme.Scheme)
	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}
	return k8sClient
}

func getClusterResourceGroup(k8sClient client.Client) (string, string, error) {
	infra := &configv1.Infrastructure{}
	if err := k8sClient.Get(context.Background(),
		client.ObjectKey{Name: "cluster"}, infra); err != nil {
		return "", "", fmt.Errorf("failed to get Infrastructure: %w", err)
	}

	if infra.Status.PlatformStatus == nil || infra.Status.PlatformStatus.Azure == nil {
		return "", "", fmt.Errorf("azure platform status not found")
	}

	return infra.Status.PlatformStatus.Azure.ResourceGroupName, infra.Status.InfrastructureName, nil
}

func UpdateResourceTags(resources []infraType.CloudResource, tags map[string]string) error {
	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("azure authentication failed: %w", err)
	}

	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		return fmt.Errorf("AZURE_SUBSCRIPTION_ID environment variable not set")
	}

	var errs []error

	for _, resource := range resources {
		var err error
		switch resource.Type {
		case infraType.CloudResourceTypeAzureVM:
			err = updateVMTags(ctx, cred, subscriptionID, resource, tags)
		case infraType.CloudResourceTypeAzureManagedDisk:
			err = updateDiskTags(ctx, cred, subscriptionID, resource, tags)
		case infraType.CloudResourceTypeAzureVirtualNetwork:
			err = updateVNetTags(ctx, cred, subscriptionID, resource, tags)
		case infraType.CloudResourceTypeAzureLoadBalancer:
			err = updateLBTags(ctx, cred, subscriptionID, resource, tags)
		case infraType.CloudResourceTypeAzurePublicIP:
			err = updateIPTags(ctx, cred, subscriptionID, resource, tags)
		case infraType.CloudResourceTypeAzureStorageAccount:
			err = updateStorageTags(ctx, cred, subscriptionID, resource, tags)
		case infraType.CloudResourceTypeAzureSubnet:
			err = updateSubnetTags(ctx, cred, subscriptionID, resource, tags)
		case infraType.CloudResourceTypeAzureNetworkSecurityGroup:
			err = updateNSGTags(ctx, cred, subscriptionID, resource, tags)
		default:
			err = fmt.Errorf("unsupported resource type: %s", resource.Type)
		}

		if err != nil {
			errs = append(errs, fmt.Errorf("failed to update %s (%s): %w",
				resource.ID, resource.Type, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("encountered %d errors: %v", len(errs), errs)
	}
	return nil
}

// Virtual Machine Tags Update
func updateVMTags(ctx context.Context, cred azcore.TokenCredential, subscriptionID string,
	resource infraType.CloudResource, tags map[string]string) error {

	client, err := armcompute.NewVirtualMachinesClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	parsedID, err := arm.ParseResourceID(resource.ID)
	if err != nil {
		return fmt.Errorf("failed to parse VM ID: %w", err)
	}

	mergedTags := mergeAzureTags(resource.Tags, tags)

	poller, err := client.BeginUpdate(ctx,
		parsedID.ResourceGroupName,
		parsedID.Name,
		armcompute.VirtualMachineUpdate{
			Tags: mergedTags,
		}, nil)
	if err != nil {
		return err
	}

	_, err = poller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: 5 * time.Second,
	})
	return err
}

// Managed Disk Tags Update
func updateDiskTags(ctx context.Context, cred azcore.TokenCredential, subscriptionID string,
	resource infraType.CloudResource, tags map[string]string) error {

	client, err := armcompute.NewDisksClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	parsedID, err := arm.ParseResourceID(resource.ID)
	if err != nil {
		return fmt.Errorf("failed to parse Disk ID: %w", err)
	}

	mergedTags := mergeAzureTags(resource.Tags, tags)

	poller, err := client.BeginUpdate(ctx,
		parsedID.ResourceGroupName,
		parsedID.Name,
		armcompute.DiskUpdate{
			Tags: mergedTags,
		}, nil)
	if err != nil {
		return err
	}

	_, err = poller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: 5 * time.Second,
	})
	return err
}

// Virtual Network Tags Update
func updateVNetTags(ctx context.Context, cred azcore.TokenCredential, subscriptionID string,
	resource infraType.CloudResource, tags map[string]string) error {

	client, err := armnetwork.NewVirtualNetworksClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	parsedID, err := arm.ParseResourceID(resource.ID)
	if err != nil {
		return fmt.Errorf("failed to parse VNet ID: %w", err)
	}

	mergedTags := mergeAzureTags(resource.Tags, tags)

	_, err = client.UpdateTags(ctx,
		parsedID.ResourceGroupName,
		parsedID.Name,
		armnetwork.TagsObject{
			Tags: mergedTags,
		}, nil)
	return err
}

// Load Balancer Tags Update
func updateLBTags(ctx context.Context, cred azcore.TokenCredential, subscriptionID string,
	resource infraType.CloudResource, tags map[string]string) error {

	client, err := armnetwork.NewLoadBalancersClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	parsedID, err := arm.ParseResourceID(resource.ID)
	if err != nil {
		return fmt.Errorf("failed to parse LB ID: %w", err)
	}

	mergedTags := mergeAzureTags(resource.Tags, tags)

	_, err = client.UpdateTags(ctx,
		parsedID.ResourceGroupName,
		parsedID.Name,
		armnetwork.TagsObject{
			Tags: mergedTags,
		}, nil)
	return err
}

// Public IP Tags Update
func updateIPTags(ctx context.Context, cred azcore.TokenCredential, subscriptionID string,
	resource infraType.CloudResource, tags map[string]string) error {

	client, err := armnetwork.NewPublicIPAddressesClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	parsedID, err := arm.ParseResourceID(resource.ID)
	if err != nil {
		return fmt.Errorf("failed to parse Public IP ID: %w", err)
	}

	mergedTags := mergeAzureTags(resource.Tags, tags)

	_, err = client.UpdateTags(ctx,
		parsedID.ResourceGroupName,
		parsedID.Name,
		armnetwork.TagsObject{
			Tags: mergedTags,
		}, nil)
	return err
}

// Storage Account Tags Update
func updateStorageTags(ctx context.Context, cred azcore.TokenCredential, subscriptionID string,
	resource infraType.CloudResource, tags map[string]string) error {

	client, err := armstorage.NewAccountsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	parsedID, err := arm.ParseResourceID(resource.ID)
	if err != nil {
		return fmt.Errorf("failed to parse Storage Account ID: %w", err)
	}

	mergedTags := mergeAzureTags(resource.Tags, tags)

	_, err = client.Update(ctx,
		parsedID.ResourceGroupName,
		parsedID.Name,
		armstorage.AccountUpdateParameters{
			Tags: mergedTags,
		}, nil)
	return err
}

// Subnet Tags Update
func updateSubnetTags(ctx context.Context, cred azcore.TokenCredential, subscriptionID string,
	resource infraType.CloudResource, tags map[string]string) error {

	client, err := armnetwork.NewSubnetsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	parsedID, err := arm.ParseResourceID(resource.ID)
	if err != nil {
		return fmt.Errorf("failed to parse Subnet ID: %w", err)
	}

	// Get parent VNet name from resource ID
	vnetName := parsedID.Parent.Name
	subnetName := parsedID.Name

	// Get existing subnet configuration
	subnet, err := client.Get(ctx,
		parsedID.ResourceGroupName,
		vnetName,
		subnetName,
		nil)
	if err != nil {
		return fmt.Errorf("failed to get subnet: %w", err)
	}

	// Merge tags
	//mergedTags := mergeAzureTags(resource.Tags, tags)
	//subnet.Subnet = mergedTags

	// Update subnet
	poller, err := client.BeginCreateOrUpdate(ctx,
		parsedID.ResourceGroupName,
		vnetName,
		subnetName,
		subnet.Subnet,
		nil)
	if err != nil {
		return fmt.Errorf("failed to start subnet update: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: 5 * time.Second,
	})
	return err
}

// Network Security Group Tags Update
func updateNSGTags(ctx context.Context, cred azcore.TokenCredential, subscriptionID string,
	resource infraType.CloudResource, tags map[string]string) error {

	client, err := armnetwork.NewSecurityGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	parsedID, err := arm.ParseResourceID(resource.ID)
	if err != nil {
		return fmt.Errorf("failed to parse NSG ID: %w", err)
	}

	mergedTags := mergeAzureTags(resource.Tags, tags)

	_, err = client.UpdateTags(ctx,
		parsedID.ResourceGroupName,
		parsedID.Name,
		armnetwork.TagsObject{
			Tags: mergedTags,
		}, nil)
	return err
}

// Helper function to merge tags
func mergeAzureTags(existing map[string]string, newTags map[string]string) map[string]*string {
	merged := make(map[string]*string)

	// Copy existing tags
	for k, v := range existing {
		merged[k] = to.Ptr(v)
	}

	// Add/overwrite with new tags
	for k, v := range newTags {
		merged[k] = to.Ptr(v)
	}

	return merged
}
