package main

import (
	"fmt"
	"io/ioutil"

	"common"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackName := ctx.Stack()
		cfg := config.New(ctx, "")

		// Setup the environment
		var env common.Environment
		cfg.RequireObject("environment", &env)

		// Create the zombix infrastructure bucket
		infraBucketConfig := common.Bucket{
			Prefix: "zombix-infra",
		}

		cfg.RequireObject("infraBucket", &infraBucketConfig)
		infraBucketConfig.Name = fmt.Sprintf("%s-%s", infraBucketConfig.Prefix, stackName)

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
				"Name":               pulumi.String(infraBucketConfig.Name),
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

		// Exports
		ctx.Export("bucketName", bucket.ID())
		ctx.Export("bucketArn", bucket.Arn)
		ctx.Export("userData", pulumi.Sprintf("s3://%s/%s", bucket.ID(), userData.Key))

		return nil
	})
}
