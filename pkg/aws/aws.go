package aws

import (
	"context"
	"fmt"
	"log"

	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infraType "github.com/anirudhAgniRedhat/openshift-metadata-manager/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2Types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var ClusterTagKey = "kubernetes.io/cluster/%s"

const ClusterTagValue = "owned"

func ListAWSResources() ([]infraType.CloudResource, error) {
	ctx := context.TODO()
	var resources []infraType.CloudResource

	k8sClient := getK8sClient()
	clusterName, err := getClusterNameFromInfrastructure(k8sClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster name: %w", err)
	}
	ClusterTagKey = fmt.Sprintf(ClusterTagKey, clusterName)

	awscfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Collect all resources
	if ec2Res, err := listEC2Instances(ctx, awscfg); err == nil {
		resources = append(resources, ec2Res...)
	}
	if s3Res, err := listS3Buckets(ctx, awscfg); err == nil {
		resources = append(resources, s3Res...)
	}
	if ebsRes, err := listEBSVolumes(ctx, awscfg); err == nil {
		resources = append(resources, ebsRes...)
	}
	if lbRes, err := listLoadBalancers(ctx, awscfg); err == nil {
		resources = append(resources, lbRes...)
	}
	if iamRes, err := listIAMRoles(ctx, awscfg); err == nil {
		resources = append(resources, iamRes...)
	}
	if vpcRes, err := listVPCs(ctx, awscfg); err == nil {
		resources = append(resources, vpcRes...)
	}
	if subnetRes, err := listSubnets(ctx, awscfg); err == nil {
		resources = append(resources, subnetRes...)
	}

	return resources, nil
}

func listEC2Instances(ctx context.Context, cfg aws.Config) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource
	client := ec2.NewFromConfig(cfg)

	result, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("error describing EC2 instances: %w", err)
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if hasTag(instance.Tags, ClusterTagKey, ClusterTagValue) {
				resources = append(resources, infraType.CloudResource{
					CloudProvider: infraType.CloudPlatformAWS,
					Type:          infraType.CloudResourceTypeAWSEC2Instance,
					ID:            aws.ToString(instance.InstanceId),
					Name:          getNameFromTags(instance.Tags),
					Tags:          convertEC2Tags(instance.Tags),
				})
			}
		}
	}
	return resources, nil
}

func listS3Buckets(ctx context.Context, cfg aws.Config) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource
	client := s3.NewFromConfig(cfg)

	result, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("error listing S3 buckets: %w", err)
	}

	for _, bucket := range result.Buckets {
		tagResult, err := client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
			Bucket: bucket.Name,
		})
		if err != nil {
			continue // Skip buckets without tags or access issues
		}

		if hasS3Tag(tagResult.TagSet, ClusterTagKey, ClusterTagValue) {
			resources = append(resources, infraType.CloudResource{
				CloudProvider: infraType.CloudPlatformAWS,
				Type:          infraType.CloudResourceTypeAWSS3Bucket,
				ID:            aws.ToString(bucket.Name),
				Name:          aws.ToString(bucket.Name),
				Tags:          convertS3Tags(tagResult.TagSet),
			})
		}
	}
	return resources, nil
}

func listEBSVolumes(ctx context.Context, cfg aws.Config) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource
	client := ec2.NewFromConfig(cfg)

	result, err := client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{})
	if err != nil {
		return nil, fmt.Errorf("error describing EBS volumes: %w", err)
	}

	for _, volume := range result.Volumes {
		if hasTag(volume.Tags, ClusterTagKey, ClusterTagValue) {
			resources = append(resources, infraType.CloudResource{
				CloudProvider: infraType.CloudPlatformAWS,
				Type:          infraType.CloudResourceTypeAWSEBSVolume,
				ID:            aws.ToString(volume.VolumeId),
				Name:          getNameFromTags(volume.Tags),
				Tags:          convertEC2Tags(volume.Tags),
			})
		}
	}
	return resources, nil
}

func listIAMRoles(ctx context.Context, cfg aws.Config) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource
	client := iam.NewFromConfig(cfg)

	result, err := client.ListRoles(ctx, &iam.ListRolesInput{})
	if err != nil {
		return nil, fmt.Errorf("error listing IAM roles: %w", err)
	}

	for _, role := range result.Roles {
		tagResult, err := client.ListRoleTags(ctx, &iam.ListRoleTagsInput{
			RoleName: role.RoleName,
		})
		if err != nil {
			continue
		}

		if hasIAMTag(tagResult.Tags, ClusterTagKey, ClusterTagValue) {
			resources = append(resources, infraType.CloudResource{
				CloudProvider: infraType.CloudPlatformAWS,
				Type:          infraType.CloudResourceTypeAWSIAMRole,
				ID:            aws.ToString(role.RoleId),
				Name:          aws.ToString(role.RoleName),
				Tags:          convertIAMTags(tagResult.Tags),
			})
		}
	}
	return resources, nil
}

func listVPCs(ctx context.Context, cfg aws.Config) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource
	client := ec2.NewFromConfig(cfg)

	result, err := client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, fmt.Errorf("error describing VPCs: %w", err)
	}

	for _, vpc := range result.Vpcs {
		if hasTag(vpc.Tags, ClusterTagKey, ClusterTagValue) {
			resources = append(resources, infraType.CloudResource{
				CloudProvider: infraType.CloudPlatformAWS,
				Type:          infraType.CloudResourceTypeAWSVPC,
				ID:            aws.ToString(vpc.VpcId),
				Name:          getNameFromTags(vpc.Tags),
				Tags:          convertEC2Tags(vpc.Tags),
			})
		}
	}
	return resources, nil
}

func listSubnets(ctx context.Context, cfg aws.Config) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource
	client := ec2.NewFromConfig(cfg)

	result, err := client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{})
	if err != nil {
		return nil, fmt.Errorf("error describing subnets: %w", err)
	}

	for _, subnet := range result.Subnets {
		if hasTag(subnet.Tags, ClusterTagKey, ClusterTagValue) {
			resources = append(resources, infraType.CloudResource{
				CloudProvider: infraType.CloudPlatformAWS,
				Type:          infraType.CloudResourceTypeAWSSubnet,
				ID:            aws.ToString(subnet.SubnetId),
				Name:          getNameFromTags(subnet.Tags),
				Tags:          convertEC2Tags(subnet.Tags),
			})
		}
	}
	return resources, nil
}

func listLoadBalancers(ctx context.Context, cfg aws.Config) ([]infraType.CloudResource, error) {
	var resources []infraType.CloudResource
	client := elasticloadbalancingv2.NewFromConfig(cfg)

	result, err := client.DescribeLoadBalancers(ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, fmt.Errorf("error describing load balancers: %w", err)
	}

	for _, lb := range result.LoadBalancers {
		tagResult, err := client.DescribeTags(ctx, &elasticloadbalancingv2.DescribeTagsInput{
			ResourceArns: []string{aws.ToString(lb.LoadBalancerArn)},
		})
		if err != nil {
			continue
		}

		for _, tagDesc := range tagResult.TagDescriptions {
			if hasELBv2Tag(tagDesc.Tags, ClusterTagKey, ClusterTagValue) {
				resources = append(resources, infraType.CloudResource{
					CloudProvider: infraType.CloudPlatformAWS,
					Type:          infraType.CloudResourceTypeAWSLoadBalancer,
					ID:            aws.ToString(lb.LoadBalancerArn),
					Name:          aws.ToString(lb.LoadBalancerName),
					Tags:          convertELBv2Tags(tagDesc.Tags),
				})
			}
		}
	}
	return resources, nil
}

// Helper functions
func getNameFromTags(tags []types.Tag) string {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == "Name" {
			return aws.ToString(tag.Value)
		}
	}
	return ""
}

func hasTag(tags []types.Tag, key, value string) bool {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key && aws.ToString(tag.Value) == value {
			return true
		}
	}
	return false
}

func hasS3Tag(tags []s3Types.Tag, key, value string) bool {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key && aws.ToString(tag.Value) == value {
			return true
		}
	}
	return false
}

func hasIAMTag(tags []iamTypes.Tag, key, value string) bool {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key && aws.ToString(tag.Value) == value {
			return true
		}
	}
	return false
}

func hasELBv2Tag(tags []elbv2Types.Tag, key, value string) bool {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key && aws.ToString(tag.Value) == value {
			return true
		}
	}
	return false
}

func convertEC2Tags(tags []types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}

func convertS3Tags(tags []s3Types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}

func convertIAMTags(tags []iamTypes.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}

func convertELBv2Tags(tags []elbv2Types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}

// Kubernetes client functions (keep existing implementations)
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

func getClusterNameFromInfrastructure(k8sClient client.Client) (string, error) {
	infra := &configv1.Infrastructure{}
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Name: "cluster"}, infra); err != nil {
		return "", fmt.Errorf("failed to get Infrastructure resource: %w", err)
	}
	return infra.Status.InfrastructureName, nil
}
