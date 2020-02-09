package main

import (
	"flag"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

var (
	optRegion      string
	optClusterName string
)

func init() {
	flag.StringVar(&optRegion, "region", "", "target cluster region")
	flag.StringVar(&optClusterName, "cluster", "default", "target cluster name")

	flag.Parse()
}

func main() {
	clusterName := optClusterName
	c := &aws.Config{}
	if optRegion != "" {
		c.Region = aws.String(optRegion)
	}
	svc := ecs.New(session.New(), c)

	containerInstanceArns, err := getClusterInstanceArns(svc, clusterName)
	if err != nil {
		log.Fatalf("failed get container instance in %s cluster: %v", clusterName, err)
	}

	if len(containerInstanceArns) == 0 {
		log.Fatalf("%s cluster does not have container instance.", clusterName)
	}

	for _, arn := range containerInstanceArns {
		err = updateContainerAgent(svc, clusterName, arn)
		if err != nil {
			log.Printf("failed update container agent that on %s in %s cluster: %v", arn, clusterName, err)
		} else {
			log.Printf("did requested to update container agent that on %s in %s cluster", arn, clusterName)
		}
	}
}

func getClusterInstanceArns(svc *ecs.ECS, clusterName string) ([]string, error) {
	input := &ecs.ListContainerInstancesInput{
		Cluster: aws.String(clusterName),
	}

	result, err := svc.ListContainerInstances(input)
	if err != nil {
		return nil, err
	}

	arns := make([]string, 0, len(result.ContainerInstanceArns))
	for _, inst := range result.ContainerInstanceArns {
		arns = append(arns, *inst)
	}

	return arns, nil
}

func updateContainerAgent(svc *ecs.ECS, clusterName, arn string) error {
	input := &ecs.UpdateContainerAgentInput{
		Cluster:           aws.String(clusterName),
		ContainerInstance: aws.String(arn),
	}

	_, err := svc.UpdateContainerAgent(input)
	if err != nil {
		return err
	}

	return nil
}
