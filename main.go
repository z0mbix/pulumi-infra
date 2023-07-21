package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type InfraBucket struct {
	Name               string
	Prefix             string
	DataClassification string
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		envName := ctx.Stack()
		cfg := config.New(ctx, "")

		// Create the zombix infrastructure bucket
		var infraBucket InfraBucket
		cfg.RequireObject("infraBucket", &infraBucket)
		infraBucket.Name = fmt.Sprintf("%s-%s", infraBucket.Prefix, envName)

		newBucket, err := s3.NewBucket(ctx, "infrastructure-bucket", &s3.BucketArgs{
			BucketName:    pulumi.String(infraBucket.Name),
			AccessControl: s3.BucketAccessControlPrivate,
			Tags: s3.BucketTagArray{
				s3.BucketTagArgs{
					Key:   pulumi.String("DataClassification"),
					Value: pulumi.String(infraBucket.DataClassification),
				},
				s3.BucketTagArgs{
					Key:   pulumi.String("EnvironmentName"),
					Value: pulumi.String(envName),
				},
				s3.BucketTagArgs{
					Key:   pulumi.String("Name"),
					Value: pulumi.String(infraBucket.Name),
				},
			},
		})
		if err != nil {
			return err
		}

		// Export things
		ctx.Export("newBucketName", newBucket.ID())
		return nil
	})
}
