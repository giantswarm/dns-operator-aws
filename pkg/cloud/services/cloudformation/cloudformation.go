package cloudformation

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func (s *Service) DeleteCloudFormation() error {
	s.scope.Info("Deleting hosted DNS zone")
	return nil
}

func (s *Service) ReconcileCloudFormation() error {
	s.scope.Info("Reconciling hosted DNS zone")
	clusterName := s.scope.InfraCluster().GetClusterName()

	// Describe or create.
	err := s.describeStack()
	if IsNotFound(err) {
		err = s.createStack()
		if err != nil {
			return err
		}

		s.scope.V(2).Info("Created new hosted zone stack for cluster", clusterName)
	} else if err != nil {
		return err
	}
	if err == nil {
		s.scope.Info(fmt.Sprintf("HostedZone Stack already exists for cluster %s", clusterName))
		return nil
	}

	//s.scope.Info(err.Error())
	s.scope.Info(fmt.Sprintf("Zone does not exists, will create new zone for cluster %s", clusterName))
	return nil
}

func (s *Service) describeStack() error {
	s.scope.Info(s.scope.InfraCluster().GetName())
	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.scope.InfraCluster().GetName()),
	}
	out, err := s.CloudFormationClient.DescribeStacks(input)
	if err != nil {
		return err
	}
	s.scope.Info(*out.Stacks[0].StackName)
	return nil
}

func (s *Service) createStack() error {
	return nil
}
