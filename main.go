package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

var (
	optRegion      string
	optClusterName string
	optInterval    int64

	optShowVersion bool
	Version        string
	Revision       string
)

func init() {
	flag.StringVar(&optRegion, "region", "", "target cluster region")
	flag.StringVar(&optClusterName, "cluster", "default", "target cluster name")
	flag.Int64Var(&optInterval, "interval", 1, "request interval by node")

	flag.BoolVar(&optShowVersion, "version", false, "show version info")
	flag.Parse()
}

func showVersion() {
	fmt.Printf("%s %s", Version, Revision)
}

func main() {
	if optShowVersion {
		showVersion()
		return
	}

	clusterName := optClusterName

	ctx := context.TODO()
	var cfg aws.Config
	var err error
	if optRegion != "" {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(optRegion))
	} else {
		cfg, err = config.LoadDefaultConfig(ctx)
	}
	if err != nil {
		log.Fatalf("failed load aws client: %v", err)
	}
	svc := ecs.NewFromConfig(cfg)

	containerInstanceArns, err := getClusterInstanceArns(ctx, svc, clusterName)
	if err != nil {
		log.Fatalf("failed get container instance in %s cluster: %v", clusterName, err)
	}

	if len(containerInstanceArns) == 0 {
		log.Fatalf("%s cluster does not have container instance.", clusterName)
	}

	for _, arn := range containerInstanceArns {
		err = updateContainerAgent(ctx, svc, clusterName, arn)
		if err != nil {
			log.Printf("failed update container agent that on %s in %s cluster: %v", arn, clusterName, err)
		} else {
			log.Printf("did requested to update container agent that on %s in %s cluster", arn, clusterName)
		}

		if optInterval > 0 {
			time.Sleep(time.Duration(optInterval) * time.Second)
		}
	}
}

func getClusterInstanceArns(ctx context.Context, svc *ecs.Client, clusterName string) ([]string, error) {
	input := &ecs.ListContainerInstancesInput{
		Cluster: aws.String(clusterName),
	}

	result, err := svc.ListContainerInstances(ctx, input)
	if err != nil {
		return nil, err
	}

	arns := make([]string, 0, len(result.ContainerInstanceArns))
	for _, inst := range result.ContainerInstanceArns {
		arns = append(arns, inst)
	}

	return arns, nil
}

func updateContainerAgent(ctx context.Context, svc *ecs.Client, clusterName, arn string) error {
	input := &ecs.UpdateContainerAgentInput{
		Cluster:           aws.String(clusterName),
		ContainerInstance: aws.String(arn),
	}

	_, err := svc.UpdateContainerAgent(ctx, input)
	if err != nil {
		return err
	}

	return nil
}
