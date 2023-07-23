package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type Ami struct {
	Distro string
	Id     string
	Owner  string
	Name   string
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Get configuration values
		projectName := ctx.Project()
		stackName := ctx.Stack()
		cfg := config.New(ctx, "")
		instanceType := "t3.micro"
		if param := cfg.Get("instanceType"); param != "" {
			instanceType = param
		}

		vpcId := cfg.Require("vpcId")
		subnetId := cfg.Require("subnetId")
		sshKeypairName := cfg.Require("sshKeypairName")

		// Lookup the latest AMI
		var ami Ami
		cfg.RequireObject("ami", &ami)

		amiResult, err := ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
			Filters: []ec2.GetAmiFilter{
				{Name: "name", Values: []string{ami.Name}},
			},
			MostRecent: pulumi.BoolRef(true),
			Owners:     []string{ami.Owner},
		}, nil)
		if err != nil {
			return err
		}
		ami.Id = amiResult.Id

		// Security Group
		securityGroup, err := ec2.NewSecurityGroup(ctx, fmt.Sprintf("nodes-%s", stackName), &ec2.SecurityGroupArgs{
			Description: pulumi.String("Allow SSH inbound"),
			VpcId:       pulumi.String(vpcId),
			Ingress: ec2.SecurityGroupIngressArray{
				&ec2.SecurityGroupIngressArgs{
					Description: pulumi.String("SSH"),
					FromPort:    pulumi.Int(22),
					ToPort:      pulumi.Int(22),
					Protocol:    pulumi.String("tcp"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
					},
				},
			},
			Egress: ec2.SecurityGroupEgressArray{
				&ec2.SecurityGroupEgressArgs{
					Description: pulumi.String("Allow all outbound traffic - Yᵒᵘ Oᶰˡʸ Lᶤᵛᵉ Oᶰᶜᵉ"),
					FromPort:    pulumi.Int(0),
					ToPort:      pulumi.Int(0),
					Protocol:    pulumi.String("-1"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
					},
				},
			},
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("allow_ssh"),
				"Environment": pulumi.String("dev"),
			},
		})
		if err != nil {
			return err
		}

		// User data to start a HTTP server in the EC2 instance
		// userData := "#!/usr/bin/env bash\necho \"Hello, World from Pulumi!\" > index.html\nnohup python3 -m http.server 80 &\n"

		coreStack, err := pulumi.NewStackReference(ctx, fmt.Sprintf("%s/core/%s", projectName, stackName), nil)
		if err != nil {
			return err
		}
		userData := coreStack.GetStringOutput(pulumi.String("userDataUrl"))

		// Create and launch an EC2 instance into the public subnet.
		server, err := ec2.NewInstance(ctx, "server", &ec2.InstanceArgs{
			Ami:          pulumi.String(ami.Id),
			InstanceType: pulumi.String(instanceType),
			KeyName:      pulumi.String(sshKeypairName),
			SubnetId:     pulumi.String(subnetId),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("node"),
			},
			UserData: userData,
			VpcSecurityGroupIds: pulumi.StringArray{
				securityGroup.ID(),
			},
		})
		if err != nil {
			return err
		}

		// Export the instance's publicly accessible IP address and hostname.
		ctx.Export("ip", server.PublicIp)
		ctx.Export("hostname", server.PublicDns)
		ctx.Export("url", server.PublicDns.ApplyT(func(publicDns string) (string, error) {
			return fmt.Sprintf("http://%v", publicDns), nil
		}).(pulumi.StringOutput))
		return nil
	})
}
