package main

import (
	"fmt"
	"io/ioutil"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type Environment struct {
	Name        string
	Description string
}

type InfraBucket struct {
	Name               string
	Prefix             string
	DataClassification DataClassification
}

type DataClassification string

const (
	CONFIDENTIAL DataClassification = "confidential"
	PUBLIC       DataClassification = "public"
	RESTRICTED   DataClassification = "restricted"
	SENSITIVE    DataClassification = "sensitive"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackName := ctx.Stack()
		cfg := config.New(ctx, "")

		// Setup the environment
		var env Environment
		cfg.RequireObject("environment", &env)

		// Create the zombix infrastructure bucket
		infraBucket := InfraBucket{
			Prefix: "zombix-infra",
		}

		cfg.RequireObject("infraBucket", &infraBucket)
		infraBucket.Name = fmt.Sprintf("%s-%s", infraBucket.Prefix, stackName)

		bucket, err := s3.NewBucket(ctx, "infrastructure-bucket", &s3.BucketArgs{
			Bucket: pulumi.String(infraBucket.Name),
			Acl:    pulumi.String("private"),
			ServerSideEncryptionConfiguration: &s3.BucketServerSideEncryptionConfigurationArgs{
				Rule: &s3.BucketServerSideEncryptionConfigurationRuleArgs{
					ApplyServerSideEncryptionByDefault: &s3.BucketServerSideEncryptionConfigurationRuleApplyServerSideEncryptionByDefaultArgs{
						SseAlgorithm: pulumi.String("AES256"),
					},
				},
			},
			Tags: pulumi.StringMap{
				"DataClassification": pulumi.String(infraBucket.DataClassification),
				"EnvironmentName":    pulumi.String(env.Name),
				"Name":               pulumi.String(infraBucket.Name),
			},
		})
		if err != nil {
			return err
		}

		// Upload Ubuntu userdata shell script to infrastructure bucket
		userDataFilePath := "../userdata/ubuntu.sh"
		userDataContent, err := ioutil.ReadFile(userDataFilePath)
		if err != nil {
			return err
		}

		_, err = s3.NewBucketObject(ctx, "fileObject", &s3.BucketObjectArgs{
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

		return nil
	})
}
