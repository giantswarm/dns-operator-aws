package route53

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/pkg/errors"
)

func (s *Service) DeleteRoute53() error {
	s.scope.V(2).Info("Deleting hosted DNS zone")
	hostedZoneID, err := s.describeWorkloadClusterZone()
	if IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	// First delete delegation record from managament
	err = s.changeManagementClusterDelegation("DELETE")
	if IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	// We need to delete all records first before we can delete the hosted zone
	err = s.deleteAllWorkloadClusterRecords("DELETE")
	if err != nil {
		return err
	}

	// Finally delete DNS zone for workload cluster
	err = s.deleteWorkloadClusterZone(hostedZoneID)
	if IsNotFound(err) {
		return nil
	} else if err != nil {
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
		// Fall through
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
	// Search host zone by DNSName
	input := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(fmt.Sprintf("%s.%s", s.scope.Name(), s.scope.BaseDomain())),
	}
	out, err := s.Route53Client.ListHostedZonesByName(input)
	if err != nil {
		return "", err
	}
	if len(out.HostedZones) == 0 {
		return "", &Route53Error{Code: http.StatusNotFound, msg: route53.ErrCodeHostedZoneNotFound}
	}

	if *out.HostedZones[0].Name != fmt.Sprintf("%s.%s.", s.scope.Name(), s.scope.BaseDomain()) {
		return "", &Route53Error{Code: http.StatusNotFound, msg: route53.ErrCodeHostedZoneNotFound}
	}

	return *out.HostedZones[0].Id, nil
}

func (s *Service) listWorkloadClusterNSRecords() ([]*route53.ResourceRecord, error) {
	hostZoneID, err := s.describeWorkloadClusterZone()
	if err != nil {
		return nil, err
	}

	// First entry is always NS record
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

	changes := []*route53.Change{
		{
			Action: aws.String(action),
			ResourceRecordSet: &route53.ResourceRecordSet{
				Name: aws.String(fmt.Sprintf("*.%s.%s", s.scope.Name(), s.scope.BaseDomain())),
				Type: aws.String("CNAME"),
				TTL:  aws.Int64(300),
				ResourceRecords: []*route53.ResourceRecord{
					{
						Value: aws.String(fmt.Sprintf("ingress.%s.%s", s.scope.Name(), s.scope.BaseDomain())),
					},
				},
			},
		},
		{
			Action: aws.String(action),
			ResourceRecordSet: &route53.ResourceRecordSet{
				Name: aws.String(fmt.Sprintf("api.%s.%s", s.scope.Name(), s.scope.BaseDomain())),
				Type: aws.String("A"),
				AliasTarget: &route53.AliasTarget{
					DNSName:              aws.String(s.scope.APIEndpoint()),
					EvaluateTargetHealth: aws.Bool(false),
					HostedZoneId:         aws.String(canonicalHostedZones[s.scope.Region()]),
				},
			},
		},
	}

	input := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(hostZoneID),
		ChangeBatch:  &route53.ChangeBatch{Changes: changes},
	}

	_, err = s.Route53Client.ChangeResourceRecordSets(input)
	if IsAlreadyExists(err) {
		// if record already exists, continue with bastion
	} else if err != nil {
		s.scope.Info("failed to change base DNS records", "error", err.Error())
		return err
	}

	// bastion is optional and the operation is transactions so all must succeed
	if s.scope.BastionIP() != "" {
		changes := []*route53.Change{
			{
				Action: aws.String(action),
				ResourceRecordSet: &route53.ResourceRecordSet{
					Name: aws.String(fmt.Sprintf("bastion1.%s.%s", s.scope.Name(), s.scope.BaseDomain())),
					Type: aws.String("A"),
					TTL:  aws.Int64(300),
					ResourceRecords: []*route53.ResourceRecord{
						{
							Value: aws.String(s.scope.BastionIP()),
						},
					},
				},
			},
		}

		input := &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(hostZoneID),
			ChangeBatch:  &route53.ChangeBatch{Changes: changes},
		}

		_, err := s.Route53Client.ChangeResourceRecordSets(input)
		if IsAlreadyExists(err) {
			// update record
			input.ChangeBatch.Changes[0].Action = aws.String("UPSERT")
			_, err := s.Route53Client.ChangeResourceRecordSets(input)
			if err != nil {
				s.scope.Info("failed to update bastion DNS records", "error", err.Error())
				return err
			}

		} else if err != nil {
			s.scope.Info("failed to change bastion DNS records", "error", err.Error())
			return err
		}
	}

	return nil
}

func (s *Service) deleteAllWorkloadClusterRecords(action string) error {
	hostZoneID, err := s.describeWorkloadClusterZone()
	if err != nil {
		return err
	}
	i := &route53.ListResourceRecordSetsInput{HostedZoneId: aws.String(hostZoneID)}
	o, err := s.Route53Client.ListResourceRecordSets(i)

	if err != nil {
		s.scope.Error(err, "failed to list DNS records", "error", err.Error())
		return err
	}
	var changes []*route53.Change
	for _, r := range o.ResourceRecordSets {
		// skip deletion of the undeletable default records
		if *r.Type == "SOA" || *r.Type == "NS" {
			continue
		}
		c := &route53.Change{
			Action: aws.String(action),
			ResourceRecordSet: &route53.ResourceRecordSet{
				Name:            r.Name,
				Type:            r.Type,
				TTL:             r.TTL,
				ResourceRecords: r.ResourceRecords,
				AliasTarget:     r.AliasTarget,
			},
		}
		changes = append(changes, c)
	}

	input := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(hostZoneID),
		ChangeBatch:  &route53.ChangeBatch{Changes: changes},
	}

	_, err = s.Route53Client.ChangeResourceRecordSets(input)
	if err != nil {
		s.scope.Info("failed to delete DNS records", "error", err.Error())
		return err
	}

	return nil
}

func (s *Service) describeManagementClusterZone() (string, error) {
	input := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(s.managementScope.BaseDomain()),
	}
	out, err := s.ManagementRoute53Client.ListHostedZonesByName(input)
	if err != nil {
		s.scope.Info(err.Error())
		return "", err
	}
	if len(out.HostedZones) == 0 {
		return "", &Route53Error{Code: http.StatusNotFound, msg: route53.ErrCodeHostedZoneNotFound}
	}

	if *out.HostedZones[0].Name != fmt.Sprintf("%s.", s.managementScope.BaseDomain()) {
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
						Name:            aws.String(fmt.Sprintf("%s.%s", s.scope.Name(), s.managementScope.BaseDomain())),
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
	if s.scope.PrivateZone() && s.scope.VPC() == "" {
		s.scope.Info("VPC ID is not ready yet for Private Hosted Zone")
		return aws.ErrMissingEndpoint

	}

	now := time.Now()
	input := &route53.CreateHostedZoneInput{
		CallerReference: aws.String(now.UTC().String()),
		Name:            aws.String(fmt.Sprintf("%s.%s.", s.scope.Name(), s.scope.BaseDomain())),
	}
	if s.scope.PrivateZone() {
		input.VPC = &route53.VPC{
			VPCId:     aws.String(s.scope.VPC()),
			VPCRegion: aws.String(s.scope.Region()),
		}
	}
	o, err := s.Route53Client.CreateHostedZone(input)
	if err != nil {
		return errors.Wrapf(err, "failed to create hosted zone for cluster: %s", s.scope.Name())
	}

	if s.scope.PrivateZone() {
		// associate management cluster VPC and any other specified VPCs to the provate hosted zone
		vpcs := append(s.scope.AdditionalVPCToAssign(), s.managementScope.VPC())
		for _, vpc := range vpcs {
			i := &route53.AssociateVPCWithHostedZoneInput{
				HostedZoneId: o.HostedZone.Id,
				VPC: &route53.VPC{
					VPCId:     aws.String(vpc),
					VPCRegion: aws.String(s.managementScope.Region()),
				},
			}
			_, err := s.Route53Client.AssociateVPCWithHostedZone(i)
			if err != nil {
				return errors.Wrapf(err, "failed to associate private hosted zone with vpc %s, for WC cluster %s", vpc, s.scope.Name())
			}
		}
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
