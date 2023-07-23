package main

import (
	"fmt"
	"io/ioutil"

	"common"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Get configuration values
		// projectName := ctx.Project()
		stackName := ctx.Stack()
		cfg := config.New(ctx, "")

		// Setup the environment
		var env common.Environment
		cfg.RequireObject("environment", &env)

		// Tags for all resources
		// defaultTags := pulumi.StringMap{
		// 	"EnvironmentName": pulumi.String(env.Name),
		// 	"ManagedBy":       pulumi.String("Pulumi"),
		// 	"PulumiProject":   pulumi.String(ctx.Project()),
		// 	"PulumiStack":     pulumi.String(stackName),
		// }

		// Create the zinfrastructure bucket
		infraBucketConfig := common.Bucket{
			Prefix: fmt.Sprintf("%s-infra", common.BucketPrefix),
		}

		cfg.RequireObject("infraBucket", &infraBucketConfig)
		infraBucketConfig.Name = fmt.Sprintf("%s-%s", infraBucketConfig.Prefix, stackName)

		// bucketTags :=
		bucket, err := s3.NewBucket(ctx, "infrastructure-bucket", &s3.BucketArgs{
			Bucket: pulumi.String(infraBucketConfig.Name),
			Acl:    pulumi.String("private"),
			ServerSideEncryptionConfiguration: &s3.BucketServerSideEncryptionConfigurationArgs{
				Rule: &s3.BucketServerSideEncryptionConfigurationRuleArgs{
					ApplyServerSideEncryptionByDefault: &s3.BucketServerSideEncryptionConfigurationRuleApplyServerSideEncryptionByDefaultArgs{
						SseAlgorithm: pulumi.String("AES256"),
					},
				},
			},
			Tags: pulumi.StringMap{
				"DataClassification": pulumi.String(infraBucketConfig.DataClassification),
				"EnvironmentName":    pulumi.String(env.Name),
				"ManagedBy":          pulumi.String("Pulumi"),
				"Name":               pulumi.String(infraBucketConfig.Name),
				"PulumiProject":      pulumi.String(ctx.Project()),
				"PulumiStack":        pulumi.String(stackName),
			},
		})
		if err != nil {
			return err
		}

		// Upload Ubuntu userdata shell script to infrastructure bucket
		userDataFilePath := "../../userdata/ubuntu.sh"
		userDataContent, err := ioutil.ReadFile(userDataFilePath)
		if err != nil {
			return err
		}

		userData, err := s3.NewBucketObject(ctx, "userDataUbuntu", &s3.BucketObjectArgs{
			Bucket:  bucket.ID(),
			Key:     pulumi.String("userdata/ubuntu.sh"),
			Content: pulumi.String(string(userDataContent)),
		})
		if err != nil {
			return err
		}

		// Create an IAM policy allowing read access to the infrastructure bucket
		policyName := fmt.Sprintf("%s-ec2-core", stackName)
		policy, err := iam.NewPolicy(ctx, policyName, &iam.PolicyArgs{
			Description: pulumi.String("Allows reading objects from the zombix-infra S3 bucket"),
			Policy: pulumi.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Action": [
							"s3:GetObject",
							"s3:ListBucket"
						],
						"Resource": [
							"arn:aws:s3:::%s",
							"arn:aws:s3:::%s/*"
						]
					},
					{
						"Effect": "Allow",
						"Action": [
							"tag:GetResources"
						],
						"Resource": ["*"]
					}
				]
			}`, bucket.ID(), bucket.ID()),
			Tags: pulumi.StringMap{
				"EnvironmentName": pulumi.String(env.Name),
				"ManagedBy":       pulumi.String("Pulumi"),
				"Name":            pulumi.String(policyName),
				"PulumiProject":   pulumi.String(ctx.Project()),
				"PulumiStack":     pulumi.String(stackName),
			},
		})
		if err != nil {
			return err
		}

		// Exports
		ctx.Export("bucketName", bucket.ID())
		ctx.Export("bucketArn", bucket.Arn)
		ctx.Export("userDataUrl", pulumi.Sprintf("s3://%s/%s", bucket.ID(), userData.Key))
		ctx.Export("ec2InstanceIamPolicy", policy.Arn)

		return nil
	})
}
