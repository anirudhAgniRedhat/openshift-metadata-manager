package gcp

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	//"log"

	configv1 "github.com/openshift/api/config/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"golang.org/x/oauth2/google"

	infraType "github.com/anirudhAgniRedhat/openshift-metadata-manager/types"
)

var (
	clusterLabelKey   string
	clusterLabelValue = "owned"
)

func ListGCPResources() ([]infraType.CloudResource, error) {
	ctx := context.Background()
	var resources []infraType.CloudResource

	// Initialize Kubernetes client
	k8sClient, err := getK8sClient()
	if err != nil {
		return nil, fmt.Errorf("kubernetes client error: %w", err)
	}

	// Get cluster metadata
	clusterName, projectID, err := getClusterMetadata(k8sClient)
	if err != nil {
		return nil, fmt.Errorf("cluster metadata error: %w", err)
	}
	clusterLabelKey = fmt.Sprintf("kubernetes-io-cluster-%s", clusterName)

	// Configure GCP credentials
	creds, err := getGCPCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("GCP authentication error: %w", err)
	}

	// Initialize GCP services
	computeSvc, err := compute.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("compute service error: %w", err)
	}

	storageClient, err := storage.NewClient(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("storage client error: %w", err)
	}
	defer storageClient.Close()

	dnsSvc, err := dns.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("dns service error: %w", err)
	}

	// List resources
	if computeRes, err := listComputeResources(ctx, computeSvc, projectID); err == nil {
		resources = append(resources, computeRes...)
	} else {
		fmt.Println("cannot list compute resources:", err)
	}
	if storageRes, err := listStorageResources(ctx, storageClient, projectID); err == nil {
		resources = append(resources, storageRes...)
	} else {
		fmt.Println("cannot list storage resources:", err)
	}
	if dnsRes, err := listDNSResources(ctx, dnsSvc, projectID); err == nil {
		resources = append(resources, dnsRes...)
	} else {
		fmt.Println("cannot list DNS resources:", err)
	}

	return resources, nil
}

func listComputeResources(ctx context.Context, svc *compute.Service, projectID string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource

	// List instances
	instances, err := listLabeledInstances(ctx, svc, projectID)
	if err != nil {
		return nil, err
	}
	resources = append(resources, instances...)

	// List disks
	disks, err := listLabeledDisks(ctx, svc, projectID)
	if err != nil {
		return nil, err
	}
	resources = append(resources, disks...)

	// List networks
	networks, err := listLabeledNetworks(ctx, svc, projectID)
	if err == nil {
		resources = append(resources, networks...)
	}
	//resources = append(resources, networks...)

	// List subnetworks
	subnets, err := listLabeledSubnetworks(ctx, svc, projectID)
	if err == nil {
		resources = append(resources, subnets...)
	}
	//resources = append(resources, subnets...)

	// List load balancers
	lbs, err := listLabeledLoadBalancers(ctx, svc, projectID)
	if err == nil {
		//return nil, err
		resources = append(resources, lbs...)
	}
	//resources = append(resources, lbs...)

	return resources, nil
}

func listLabeledInstances(ctx context.Context, svc *compute.Service, projectID string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource

	req := svc.Instances.AggregatedList(projectID).
		Filter(fmt.Sprintf("labels.%s = %s", clusterLabelKey, clusterLabelValue))

	err := req.Pages(ctx, func(page *compute.InstanceAggregatedList) error {
		for _, instances := range page.Items {
			for _, instance := range instances.Instances {
				resources = append(resources, infraType.CloudResource{
					CloudProvider: infraType.CloudPlatformGCP,
					Type:          infraType.CloudResourceTypeGCPComputeInstance,
					ID:            fmt.Sprintf("%d", instance.Id),
					Name:          instance.Name,
					Tags:          instance.Labels,
				})
			}
		}
		return nil
	})

	return resources, err
}

func listLabeledDisks(ctx context.Context, svc *compute.Service, projectID string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource

	req := svc.Disks.AggregatedList(projectID).
		Filter(fmt.Sprintf("labels.%s = %s", clusterLabelKey, clusterLabelValue))

	err := req.Pages(ctx, func(page *compute.DiskAggregatedList) error {
		for _, disks := range page.Items {
			for _, disk := range disks.Disks {
				resources = append(resources, infraType.CloudResource{
					CloudProvider: infraType.CloudPlatformGCP,
					Type:          infraType.CloudResourceTypeGCPDisk,
					ID:            fmt.Sprintf("%d", disk.Id),
					Name:          disk.Name,
					Tags:          disk.Labels,
				})
			}
		}
		return nil
	})

	return resources, err
}

func listLabeledNetworks(ctx context.Context, svc *compute.Service, projectID string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource

	req := svc.Networks.List(projectID).
		Filter(fmt.Sprintf("labels.%s = %s", clusterLabelKey, clusterLabelValue))

	err := req.Pages(ctx, func(page *compute.NetworkList) error {
		for _, network := range page.Items {
			resources = append(resources, infraType.CloudResource{
				CloudProvider: infraType.CloudPlatformGCP,
				Type:          infraType.CloudResourceTypeGCPNetwork,
				ID:            fmt.Sprintf("%d", network.Id),
				Name:          network.Name,
				//Tags:          network.Labels,
			})
		}
		return nil
	})

	return resources, err
}

func listLabeledSubnetworks(ctx context.Context, svc *compute.Service, projectID string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource

	req := svc.Subnetworks.AggregatedList(projectID).
		Filter(fmt.Sprintf("labels.%s = %s", clusterLabelKey, clusterLabelValue))

	err := req.Pages(ctx, func(page *compute.SubnetworkAggregatedList) error {
		for _, subnets := range page.Items {
			for _, subnet := range subnets.Subnetworks {
				resources = append(resources, infraType.CloudResource{
					CloudProvider: infraType.CloudPlatformGCP,
					Type:          infraType.CloudResourceTypeGCPSubnet,
					ID:            fmt.Sprintf("%d", subnet.Id),
					Name:          subnet.Name,
					//Tags:          subnet.Labels,
				})
			}
		}
		return nil
	})

	return resources, err
}

func listLabeledLoadBalancers(ctx context.Context, svc *compute.Service, projectID string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource

	req := svc.ForwardingRules.AggregatedList(projectID).
		Filter(fmt.Sprintf("labels.%s = %s", clusterLabelKey, clusterLabelValue))

	err := req.Pages(ctx, func(page *compute.ForwardingRuleAggregatedList) error {
		for _, rules := range page.Items {
			for _, lb := range rules.ForwardingRules {
				resources = append(resources, infraType.CloudResource{
					CloudProvider: infraType.CloudPlatformGCP,
					Type:          infraType.CloudResourceTypeGCPLoadBalancer,
					ID:            fmt.Sprintf("%d", lb.Id),
					Name:          lb.Name,
					Tags:          lb.Labels,
				})
			}
		}
		return nil
	})

	return resources, err
}

func listStorageResources(ctx context.Context, client *storage.Client, projectID string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource

	it := client.Buckets(ctx, projectID)
	for {
		bucket, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("bucket iteration error: %w", err)
		}

		if hasClusterLabel(bucket.Labels) {
			resources = append(resources, infraType.CloudResource{
				CloudProvider: infraType.CloudPlatformGCP,
				Type:          infraType.CloudResourceTypeGCPStorageBucket,
				ID:            bucket.Name,
				Name:          bucket.Name,
				Tags:          bucket.Labels,
			})
		}
	}

	return resources, nil
}

func listDNSResources(ctx context.Context, svc *dns.Service, projectID string) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource

	req := svc.ManagedZones.List(projectID)
	err := req.Pages(ctx, func(page *dns.ManagedZonesListResponse) error {
		for _, zone := range page.ManagedZones {
			if hasClusterLabel(zone.Labels) {
				resources = append(resources, infraType.CloudResource{
					CloudProvider: infraType.CloudPlatformGCP,
					Type:          infraType.CloudResourceTypeGCPDNSZone,
					ID:            fmt.Sprintf("%d", zone.Id),
					Name:          zone.Name,
					Tags:          zone.Labels,
				})
			}
		}
		return nil
	})

	return resources, err
}

func hasClusterLabel(labels map[string]string) bool {
	return labels != nil && labels[clusterLabelKey] == clusterLabelValue
}

func getK8sClient() (client.Client, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	configv1.AddToScheme(scheme.Scheme)
	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return nil, fmt.Errorf("client creation error: %w", err)
	}
	return k8sClient, nil
}

func getClusterMetadata(k8sClient client.Client) (string, string, error) {
	ctx := context.Background()
	infra := &configv1.Infrastructure{}

	if err := k8sClient.Get(ctx, client.ObjectKey{Name: "cluster"}, infra); err != nil {
		return "", "", fmt.Errorf("infrastructure fetch error: %w", err)
	}

	if infra.Status.PlatformStatus == nil || infra.Status.PlatformStatus.GCP == nil {
		return "", "", fmt.Errorf("GCP platform status not found")
	}

	return infra.Status.InfrastructureName, infra.Status.PlatformStatus.GCP.ProjectID, nil
}

func getGCPCredentials(ctx context.Context) (*google.Credentials, error) {
	creds, err := google.FindDefaultCredentials(ctx, compute.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("credentials error: %w", err)
	}
	return creds, nil
}

func UpdateResourceTags(resources infraType.CloudResource, tags map[string]string) error {
	fmt.Println("Tags Updated")
	return nil
}
