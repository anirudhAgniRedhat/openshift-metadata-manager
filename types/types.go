package types

type CloudPlatform string

const CloudPlatformAWS CloudPlatform = "AWS"
const CloudPlatformAzure CloudPlatform = "Azure"
const CloudPlatformOpenStack CloudPlatform = "OpenStack"
const CloudPlatformGCP CloudPlatform = "GCP"
const CloudPlatformUnknown CloudPlatform = "Unknown"
const CloudPlatformIBM = "IBM"

//type AwsCloudResources struct {
//	EC2Instances []string
//	S3Buckets    []string
//	EBSVolumes   []string
//	IAMRoles     []string
//	IAMUsers     []string
//	VPCs         []string
//	Subnets      []string
//}
//
//type AzureCloudResources struct {
//	VMs                  []string
//	VPCs                 []string
//	ManagedDisks         []string
//	VirtualNetworks      []string
//	LoadBalancers        []string
//	PublicIPs            []string
//	StorageAccounts      []string
//	Subnets              []string
//	NetworkSecurityGroup []string
//}
//
//type GCPCloudResources struct {
//	ComputesInstances []string
//	Disks             []string
//	Networks          []string
//	Subnets           []string
//	StorageBuckets    []string
//	LoadBalancers     []string
//	DNSZones          []string
//}
//
//type CloudResource struct {
//	AwsCloudResources   AwsCloudResources   `json:"awsCloudResources"`
//	AzureCloudResources AzureCloudResources `json:"azureCloudResources"`
//	GCPCloudResources   GCPCloudResources   `json:"gcpCloudResources"`
//}

type CloudResourceType string

const (
	// AWS Resource Types
	CloudResourceTypeAWSEC2Instance  CloudResourceType = "AWSEC2Instance"
	CloudResourceTypeAWSS3Bucket     CloudResourceType = "AWSS3Bucket"
	CloudResourceTypeAWSEBSVolume    CloudResourceType = "AWSEBSVolume"
	CloudResourceTypeAWSIAMRole      CloudResourceType = "AWSIAMRole"
	CloudResourceTypeAWSIAMUser      CloudResourceType = "AWSIAMUser"
	CloudResourceTypeAWSVPC          CloudResourceType = "AWSVPC"
	CloudResourceTypeAWSLoadBalancer CloudResourceType = "AWSLoadBalancer"
	CloudResourceTypeAWSSubnet       CloudResourceType = "AWSSubnet"

	// Azure Resource Types
	CloudResourceTypeAzureVM                   CloudResourceType = "AzureVM"
	CloudResourceTypeAzureVPC                  CloudResourceType = "AzureVPC"
	CloudResourceTypeAzureManagedDisk          CloudResourceType = "AzureManagedDisk"
	CloudResourceTypeAzureVirtualNetwork       CloudResourceType = "AzureVirtualNetwork"
	CloudResourceTypeAzureLoadBalancer         CloudResourceType = "AzureLoadBalancer"
	CloudResourceTypeAzurePublicIP             CloudResourceType = "AzurePublicIP"
	CloudResourceTypeAzureStorageAccount       CloudResourceType = "AzureStorageAccount"
	CloudResourceTypeAzureSubnet               CloudResourceType = "AzureSubnet"
	CloudResourceTypeAzureNetworkSecurityGroup CloudResourceType = "AzureNetworkSecurityGroup"

	// GCP Resource Types
	CloudResourceTypeGCPComputeInstance CloudResourceType = "GCPComputeInstance"
	CloudResourceTypeGCPDisk            CloudResourceType = "GCPDisk"
	CloudResourceTypeGCPNetwork         CloudResourceType = "GCPNetwork"
	CloudResourceTypeGCPSubnet          CloudResourceType = "GCPSubnet"
	CloudResourceTypeGCPStorageBucket   CloudResourceType = "GCPStorageBucket"
	CloudResourceTypeGCPLoadBalancer    CloudResourceType = "GCPLoadBalancer"
	CloudResourceTypeGCPDNSZone         CloudResourceType = "GCPDNSZone"
)

type CloudResource struct {
	CloudProvider CloudPlatform
	Type          CloudResourceType
	ID            string
	Name          string
	Tags          map[string]string
}
