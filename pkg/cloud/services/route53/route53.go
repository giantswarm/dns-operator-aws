package route53

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/pkg/errors"
)

// 🤷‍♂️
var baseDomain = "gauss.eu-west-1.aws.gigantic.io"

func (s *Service) DeleteRoute53() error {
	s.scope.V(2).Info("Deleting hosted DNS zone")
	hostedZoneID, err := s.describeWorkloadClusterZone()
	if IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	if err := s.changeManagementClusterDelegation("DELETE"); err != nil {
		return err
	}

	if err := s.changeWorkloadClusterRecords("DELETE"); err != nil {
		return err
	}

	err = s.deleteWorkloadClusterZone(hostedZoneID)
	if err != nil {
		return err
	}
	s.scope.V(2).Info(fmt.Sprintf("Deleting hosted zone completed successfully for cluster %s", s.scope.Name()))
	return nil
}

func (s *Service) ReconcileRoute53() error {
	s.scope.V(2).Info("Reconciling hosted DNS zone")

	// Describe or create.
	_, err := s.describeWorkloadClusterZone()
	if IsNotFound(err) {
		err = s.createWorkloadClusterZone()
		if err != nil {
			return err
		}
		s.scope.Info(fmt.Sprintf("Created new hosted zone for cluster %s", s.scope.Name()))
	} else if err != nil {
		return err
	}

	err = s.changeWorkloadClusterRecords("CREATE")
	if IsNotFound(err) {
		// fall trough
	} else if err != nil {
		return err
	}

	err = s.changeManagementClusterDelegation("CREATE")
	if IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}

func (s *Service) describeWorkloadClusterZone() (string, error) {
	input := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(fmt.Sprintf("%s.k8s.%s", s.scope.Name(), baseDomain)),
	}
	out, err := s.Route53Client.ListHostedZonesByName(input)
	if err != nil {
		return "", err
	}
	if len(out.HostedZones) == 0 {
		return "", &Route53Error{Code: http.StatusNotFound, msg: route53.ErrCodeHostedZoneNotFound}
	}

	if *out.HostedZones[0].Name != fmt.Sprintf("%s.k8s.%s.", s.scope.Name(), baseDomain) {
		return "", &Route53Error{Code: http.StatusNotFound, msg: route53.ErrCodeHostedZoneNotFound}
	}

	return *out.HostedZones[0].Id, nil
}

func (s *Service) listWorkloadClusterNSRecords() ([]*route53.ResourceRecord, error) {
	hostZoneID, err := s.describeWorkloadClusterZone()
	if err != nil {
		return nil, err
	}

	// first entry is always NS record
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(hostZoneID),
		MaxItems:     aws.String("1"),
	}

	output, err := s.Route53Client.ListResourceRecordSets(input)
	if err != nil {
		return nil, err
	}
	return output.ResourceRecordSets[0].ResourceRecords, nil
}

func (s *Service) changeWorkloadClusterRecords(action string) error {
	s.scope.Info(s.scope.APIEndpoint())
	if s.scope.APIEndpoint() == "" {
		s.scope.Info("API endpoint is not ready yet.")
		return aws.ErrMissingEndpoint
	}

	hostZoneID, err := s.describeWorkloadClusterZone()
	if err != nil {
		return err
	}

	input := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(hostZoneID),
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(action),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(fmt.Sprintf("*.%s.k8s.%s", s.scope.Name(), baseDomain)),
						Type: aws.String("CNAME"),
						TTL:  aws.Int64(300),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(fmt.Sprintf("ingress.%s.k8s.%s", s.scope.Name(), baseDomain)),
							},
						},
					},
				},
				{
					Action: aws.String(action),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(fmt.Sprintf("api.%s.k8s.%s", s.scope.Name(), baseDomain)),
						Type: aws.String("A"),
						AliasTarget: &route53.AliasTarget{
							DNSName:              aws.String(s.scope.APIEndpoint()),
							EvaluateTargetHealth: aws.Bool(false),
							HostedZoneId:         aws.String(canonicalHostedZones[s.scope.Region()]),
						},
					},
				},
			},
		},
	}

	_, err = s.Route53Client.ChangeResourceRecordSets(input)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) describeManagementClusterZone() (string, error) {
	input := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(baseDomain),
	}
	out, err := s.ManagementRoute53Client.ListHostedZonesByName(input)
	if err != nil {
		s.scope.Info(err.Error())
		return "", err
	}
	if len(out.HostedZones) == 0 {
		return "", &Route53Error{Code: http.StatusNotFound, msg: route53.ErrCodeHostedZoneNotFound}
	}

	if *out.HostedZones[0].Name != fmt.Sprintf("%s.", baseDomain) {
		return "", &Route53Error{Code: http.StatusNotFound, msg: route53.ErrCodeHostedZoneNotFound}
	}

	return *out.HostedZones[0].Id, nil
}

func (s *Service) changeManagementClusterDelegation(action string) error {
	hostZoneID, err := s.describeManagementClusterZone()
	if err != nil {
		return err
	}

	records, err := s.listWorkloadClusterNSRecords()
	if err != nil {
		return err
	}

	input := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(hostZoneID),
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(action),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            aws.String(fmt.Sprintf("%s.k8s.%s", s.scope.Name(), baseDomain)),
						Type:            aws.String("NS"),
						TTL:             aws.Int64(300),
						ResourceRecords: records,
					},
				},
			},
		},
	}

	_, err = s.ManagementRoute53Client.ChangeResourceRecordSets(input)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) createWorkloadClusterZone() error {
	now := time.Now()
	input := &route53.CreateHostedZoneInput{
		CallerReference: aws.String(now.UTC().String()),
		Name:            aws.String(fmt.Sprintf("%s.k8s.%s.", s.scope.Name(), baseDomain)),
	}
	_, err := s.Route53Client.CreateHostedZone(input)
	if err != nil {
		return errors.Wrapf(err, "failed to create hosted zone for cluster: %s", s.scope.Name())
	}
	return nil
}

func (s *Service) deleteWorkloadClusterZone(hostedZoneID string) error {
	input := &route53.DeleteHostedZoneInput{
		Id: aws.String(hostedZoneID),
	}
	_, err := s.Route53Client.DeleteHostedZone(input)
	if err != nil {
		return errors.Wrapf(err, "failed to delete hosted zone for cluster: %s", s.scope.Name())
	}
	return nil
}
