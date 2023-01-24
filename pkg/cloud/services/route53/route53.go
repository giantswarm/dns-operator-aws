package route53

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/pkg/errors"
)

func (s *Service) DeleteRoute53() error {
	s.scope.Logger().V(2).Info("Deleting hosted DNS zone")
	hostedZoneID, err := s.describeWorkloadClusterZone()
	if IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	// delegation is only done for public zones
	if !s.scope.PrivateZone() {
		// First delete delegation record from managament
		err = s.changeManagementClusterDelegation("DELETE")
		if IsNotFound(err) {
			return nil
		} else if err != nil {
			return err
		}
	}
	// We need to delete all records first before we can delete the hosted zone
	err = s.deleteAllWorkloadClusterRecords("DELETE")
	if err != nil {
		return errors.Wrapf(err, "failed to delete")
	}

	// Finally delete DNS zone for workload cluster
	err = s.deleteWorkloadClusterZone(hostedZoneID)
	if IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}
	s.scope.Logger().V(2).Info(fmt.Sprintf("Deleting hosted zone completed successfully for cluster %s", s.scope.Name()))

	return nil
}

func (s *Service) ReconcileRoute53() error {
	s.scope.Logger().Info("Reconciling hosted DNS zone")

	// Describe or create.
	_, err := s.describeWorkloadClusterZone()
	if IsNotFound(err) {
		err = s.createWorkloadClusterZone()
		if err != nil {
			return err
		}
		s.scope.Logger().Info(fmt.Sprintf("Created new hosted zone for cluster %s", s.scope.Name()))
	} else if err != nil {
		return err
	}

	err = s.changeWorkloadClusterRecords("CREATE")
	if IsNotFound(err) {
		// Fall through
	} else if err != nil {
		return errors.Wrap(err, "failed creating workload cluster DNS records")
	}

	// delegation only make sense for public zones
	if !s.scope.PrivateZone() {
		err = s.changeManagementClusterDelegation("CREATE")
		if IsNotFound(err) {
			return nil
		} else if err != nil {
			return err
		}
	}

	err = s.associateResolverRules()
	if err != nil {
		return err
	}

	return nil
}

// associateResolverRules takes all the resolver rules and try to associate them with the workload cluster VPC.
func (s *Service) associateResolverRules() error {
	if s.scope.AssociateResolverRules() {
		resolverRules, err := s.listResolverRules()
		if err != nil {
			return errors.Wrap(err, "failed to list AWS Resolver rule")
		}

		associations, err := s.getResolverRuleAssociations(nil)
		if err != nil {
			return errors.Wrap(err, "failed to list AWS Resolver rules associations")
		}
		s.scope.Logger().Info("Got resolver rule associations", "associations", associations)

		vpcCidr := s.scope.VPCCidr()
		for _, rule := range resolverRules {
			if !s.associationsHasRule(associations, rule) {

				belong, err := ruleTargetsBelongToSubnet(rule.TargetIps, vpcCidr)
				if err != nil {
					s.scope.Logger().Error(err, "failed to check if the resolver rule belongs to the VPC", "ruleName", *rule.Name, "vpc", s.scope.VPC())
					continue
				}

				if !belong {
					s.scope.Logger().Info("No existing resolver rule association found, associating now", "rule", rule)
					i := &route53resolver.AssociateResolverRuleInput{
						Name:           rule.Name,
						VPCId:          aws.String(s.scope.VPC()),
						ResolverRuleId: rule.Id,
					}
					_, err = s.Route53ResolverClient.AssociateResolverRule(i)
					if err != nil {
						s.scope.Logger().Error(err, "failed to assign resolver rule to VPC", "ruleName", *rule.Name, "vpc", s.scope.VPC())
						continue
					}
				}
			}
		}
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

// changeWorkloadClusterRecords creates the DNS records required by the workload cluster like
// - a wildcard `CNAME` record pointing to the ingress record
// - an `A` dns record 'api' pointing to the control plane LB
// - optionally an `A` dns record 'bastion1' pointing to the bastion machine IP
func (s *Service) changeWorkloadClusterRecords(action string) error {
	if s.scope.APIEndpoint() == "" {
		s.scope.Logger().Info("API endpoint is not ready yet.")
		return aws.ErrMissingEndpoint
	}

	hostZoneID, err := s.describeWorkloadClusterZone()
	if err != nil {
		return errors.Wrapf(err, "failed describing workload cluster hosted zone")
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
		s.scope.Logger().Info("failed to change base DNS records", "error", err.Error())
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
				s.scope.Logger().Info("failed to update bastion DNS records", "error", err.Error())
				return err
			}

		} else if err != nil {
			s.scope.Logger().Info("failed to change bastion DNS records", "error", err.Error())
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
		s.scope.Logger().Error(err, "failed to list DNS records", "error", err.Error())
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
	if len(changes) == 0 {
		// nothing to delete
		return nil
	}

	_, err = s.Route53Client.ChangeResourceRecordSets(input)
	if err != nil {
		s.scope.Logger().Info("failed to delete DNS records", "error", err.Error())
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
		s.scope.Logger().Info(err.Error())
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
		s.scope.Logger().Info("VPC ID is not ready yet for Private Hosted Zone")
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
		// associate management cluster VPC and any other specified VPCs to the private hosted zone
		vpcs := append(s.scope.AdditionalVPCToAssign(), s.managementScope.VPC())
		for _, vpc := range vpcs {
			if vpc == "" {
				continue
			}
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

func (s *Service) getResolverRuleAssociations(nextToken *string) ([]*route53resolver.ResolverRuleAssociation, error) {
	s.scope.Logger().Info("Fetching resolver rule associations", "nextToken", nextToken)
	ruleAssociations := []*route53resolver.ResolverRuleAssociation{}

	associations, err := s.Route53ResolverClient.ListResolverRuleAssociations(&route53resolver.ListResolverRuleAssociationsInput{
		MaxResults: aws.Int64(100),
		Filters: []*route53resolver.Filter{
			{
				Name:   aws.String("VPCId"),
				Values: []*string{aws.String(s.scope.VPC())},
			},
		},
		NextToken: nextToken,
	})
	if err != nil {
		return ruleAssociations, errors.Wrap(err, "failed to list AWS Resolver rule associations")
	}
	ruleAssociations = append(ruleAssociations, associations.ResolverRuleAssociations...)

	// If we have more than we can query in one call we need to recursively keep calling until we have all associations
	if associations.NextToken != nil && *associations.NextToken != "" {
		next, err := s.getResolverRuleAssociations(associations.NextToken)
		if err != nil {
			return ruleAssociations, errors.Wrap(err, "failed to list AWS Resolver rule associations")
		}
		ruleAssociations = append(ruleAssociations, next...)
	}

	return ruleAssociations, err
}

func (s *Service) associationsHasRule(associations []*route53resolver.ResolverRuleAssociation, rule *route53resolver.ResolverRule) bool {
	for _, a := range associations {
		if *a.ResolverRuleId == *rule.Id && (*a.Status == route53resolver.ResolverRuleAssociationStatusCreating || *a.Status == route53resolver.ResolverRuleAssociationStatusComplete) {
			return true
		}
	}
	return false
}

// listResolverRules fetches the resolver rules of type FORWARD. When an AWS account id is passed, only the rules
// from that account will be listed.
func (s *Service) listResolverRules() ([]*route53resolver.ResolverRule, error) {
	var resolverRules, filteredRules []*route53resolver.ResolverRule
	var resolverRulesInput *route53resolver.ListResolverRulesInput
	var resolverRulesOutput *route53resolver.ListResolverRulesOutput
	var err error

	resolverRulesInput = &route53resolver.ListResolverRulesInput{
		MaxResults: aws.Int64(100),
		Filters: []*route53resolver.Filter{
			{
				Name:   aws.String("TYPE"),
				Values: aws.StringSlice([]string{"FORWARD"}),
			},
		},
	}
	// Fetch first page of resolver rules.
	resolverRulesOutput, err = s.Route53ResolverClient.ListResolverRules(resolverRulesInput)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list resolver rules")
	}

	resolverRules = append(resolverRules, resolverRulesOutput.ResolverRules...)

	// If the response contains `NexToken` we need to send another request with the response token to get the next page.
	for resolverRulesOutput.NextToken != nil && *resolverRulesOutput.NextToken != "" {
		resolverRulesInput.NextToken = resolverRulesOutput.NextToken
		resolverRulesOutput, err = s.Route53ResolverClient.ListResolverRules(resolverRulesInput)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list resolver rules")
		}
		resolverRules = append(resolverRules, resolverRulesOutput.ResolverRules...)
	}

	creatorID := s.scope.ResolverRulesCreatorAccount()
	if creatorID != "" {
		for _, rule := range resolverRules {
			if *(rule.OwnerId) == creatorID {
				filteredRules = append(filteredRules, rule)
			}
		}

		return filteredRules, nil
	}

	return resolverRules, nil
}

// Checks if any of rule target IPs belongs to a CIDR range
func ruleTargetsBelongToSubnet(targetIps []*route53resolver.TargetAddress, vpcCidr string) (bool, error) {
	_, ipNetVpc, err := net.ParseCIDR(vpcCidr)
	if err != nil {
		return false, err
	}

	for _, targetIp := range targetIps {
		ipIP := net.ParseIP(*targetIp.Ip)
		if ipNetVpc.Contains(ipIP) {
			return true, nil
		}
	}

	return false, nil

}
