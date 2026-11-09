package tencentcloud

import (
	"context"

	"github.com/dnridwn/cfgate/internal/firewall"
	"github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common/profile"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/lighthouse/v20200324"
)

const (
	ACCEPT firewall.Action = "ACCEPT"
	DROP   firewall.Action = "DROP"
)

const (
	TCP  firewall.Protocol = "TCP"
	UDP  firewall.Protocol = "UDP"
	ICMP firewall.Protocol = "ICMP"
)

type tencentCloud struct {
	Region        string
	InstanceID    string
	latestVersion uint64
	c             *lighthouse.Client
}

func NewTencentCloud(secretID, secretKey, region, instanceID string) (firewall.Firewall, error) {
	cred := common.NewCredential(secretID, secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "lighthouse.intl.tencentcloudapi.com"

	c, err := lighthouse.NewClient(cred, region, cpf)
	if err != nil {
		return nil, err
	}

	return &tencentCloud{
		Region:     region,
		InstanceID: instanceID,
		c:          c,
	}, nil
}

func (tc *tencentCloud) ListRules(ctx context.Context) ([]*firewall.Rule, error) {
	foundRules := []*firewall.Rule{}

	lv := uint64(0)
	defer func() {
		if lv > tc.latestVersion {
			tc.latestVersion = lv
		}
	}()

	chunkSize := int64(100)
	offset := -1 * chunkSize
	for {
		offset += chunkSize
		req := lighthouse.NewDescribeFirewallRulesRequest()
		req.InstanceId = &tc.InstanceID
		req.Limit = common.Int64Ptr(chunkSize)
		req.Offset = &offset

		resp, err := tc.c.DescribeFirewallRules(req)
		if err != nil {
			return nil, err
		}

		var version uint64
		if resp.Response.FirewallVersion != nil {
			version = *resp.Response.FirewallVersion
		}

		if version > lv {
			lv = version
		}

		if len(resp.Response.FirewallRuleSet) == 0 {
			break
		}

		for _, rule := range resp.Response.FirewallRuleSet {
			foundRules = append(foundRules, &firewall.Rule{
				Action:      firewall.Action(*rule.Action),
				CidrBlock:   *rule.CidrBlock,
				Description: *rule.FirewallRuleDescription,
				Port:        *rule.Port,
				Protocol:    firewall.Protocol(*rule.Protocol),
			})
		}
	}

	return foundRules, nil
}

func (tc *tencentCloud) ApplyRules(ctx context.Context, rules []*firewall.Rule) error {
	req := lighthouse.NewModifyFirewallRulesRequest()
	req.InstanceId = &tc.InstanceID
	req.FirewallVersion = &tc.latestVersion

	for _, rule := range rules {
		req.FirewallRules = append(req.FirewallRules, &lighthouse.FirewallRule{
			Action:                  common.StringPtr(string(rule.Action)),
			CidrBlock:               common.StringPtr(rule.CidrBlock),
			FirewallRuleDescription: common.StringPtr(rule.Description),
			Port:                    common.StringPtr(rule.Port),
			Protocol:                common.StringPtr(string(rule.Protocol)),
		})
	}

	_, err := tc.c.ModifyFirewallRules(req)
	return err
}
