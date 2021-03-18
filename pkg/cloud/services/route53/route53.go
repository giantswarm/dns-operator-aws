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
var baseDomain = "k8s.gauss.eu-west-1.aws.gigantic.io"

func (s *Service) DeleteRoute53() error {
	s.scope.V(2).Info("Deleting hosted DNS zone")
	hostedZoneID, err := s.describeRoute53Zone()

	if IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	if err := s.changeRoute53Records("DELETE"); err != nil {
		return err
	}

	err = s.deleteRoute53Zone(hostedZoneID)
	if err != nil {
		return err
	}
	s.scope.V(2).Info(fmt.Sprintf("Deleting hosted zone completed successfully for cluster %s", s.scope.Name()))
	return nil
}

func (s *Service) ReconcileRoute53() error {
	s.scope.V(2).Info("Reconciling hosted DNS zone")

	// Describe or create.
	_, err := s.describeRoute53Zone()
	if IsNotFound(err) {
		err = s.createRoute53Zone()
		if err != nil {
			return err
		}
		s.scope.Info(fmt.Sprintf("Created new hosted zone for cluster %s", s.scope.Name()))
	} else if err != nil {
		return err
	}

	if err := s.changeRoute53Records("CREATE"); err != nil {
		return err
	}

	return nil
}

func (s *Service) describeRoute53Zone() (string, error) {
	input := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(fmt.Sprintf("%s.%s", s.scope.Name(), baseDomain)),
	}
	out, err := s.Route53Client.ListHostedZonesByName(input)
	if err != nil {
		return "", err
	}
	if len(out.HostedZones) == 0 {
		return "", &Route53Error{Code: http.StatusNotFound, msg: route53.ErrCodeHostedZoneNotFound}
	}

	if *out.HostedZones[0].Name != fmt.Sprintf("%s.%s.", s.scope.Name(), baseDomain) {
		return "", &Route53Error{Code: http.StatusNotFound, msg: route53.ErrCodeHostedZoneNotFound}
	}

	return *out.HostedZones[0].Id, nil
}

func (s *Service) changeRoute53Records(action string) error {
	s.scope.Info(s.scope.APIEndpoint())
	if s.scope.APIEndpoint() == "" {
		s.scope.Info("API endpoint is not ready yet.")
		return nil
	}

	hostZoneID, err := s.describeRoute53Zone()
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
						Name: aws.String(fmt.Sprintf("*.%s.%s", s.scope.Name(), baseDomain)),
						Type: aws.String("CNAME"),
						TTL:  aws.Int64(300),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(fmt.Sprintf("ingress.%s.%s", s.scope.Name(), baseDomain)),
							},
						},
					},
				},
				{
					Action: aws.String(action),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(fmt.Sprintf("api.%s.%s", s.scope.Name(), baseDomain)),
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

func (s *Service) createRoute53Zone() error {
	now := time.Now()
	input := &route53.CreateHostedZoneInput{
		CallerReference: aws.String(now.UTC().String()),
		Name:            aws.String(fmt.Sprintf("%s.%s.", s.scope.Name(), baseDomain)),
	}
	_, err := s.Route53Client.CreateHostedZone(input)
	if err != nil {
		return errors.Wrapf(err, "failed to create hosted zone for cluster: %s", s.scope.Name())
	}
	return nil
}

func (s *Service) deleteRoute53Zone(hostedZoneID string) error {
	input := &route53.DeleteHostedZoneInput{
		Id: aws.String(hostedZoneID),
	}
	_, err := s.Route53Client.DeleteHostedZone(input)
	if err != nil {
		return errors.Wrapf(err, "failed to delete hosted zone for cluster: %s", s.scope.Name())
	}
	return nil
}
