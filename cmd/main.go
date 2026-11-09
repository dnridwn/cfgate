package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dnridwn/cfgate/internal/cloudflare"
	"github.com/dnridwn/cfgate/internal/firewall"
	"github.com/dnridwn/cfgate/internal/firewall/tencentcloud"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	cmd := &cobra.Command{}
	cmd.Use = ""

	cmdArgs := struct {
		Region      string
		InstanceID  string
		IncludeIPv6 bool
	}{}
	cmd.Flags().StringVarP(&cmdArgs.Region, "region", "r", "", "Region")
	cmd.MarkFlagRequired("region")
	cmd.Flags().StringVarP(&cmdArgs.InstanceID, "instance-id", "i", "", "ID of target instance")
	cmd.MarkFlagRequired("instance-id")
	cmd.Flags().BoolVarP(&cmdArgs.IncludeIPv6, "include-ipv6", "", false, "Include IPv6")

	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show the version of the CLI",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Version:", version)
		},
	})

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ipv4Cidrs, ipv6Cidrs, err := cloudflare.GetIPs()
		if err != nil {
			return err
		}

		rules := []*firewall.Rule{}
		for _, ipv4Cidr := range ipv4Cidrs {
			newRules := []*firewall.Rule{
				{
					Action:      tencentcloud.ACCEPT,
					CidrBlock:   ipv4Cidr,
					Description: "Auto-set by cfgate",
					Port:        "80",
					Protocol:    tencentcloud.TCP,
				},
				{
					Action:      tencentcloud.ACCEPT,
					CidrBlock:   ipv4Cidr,
					Description: "Auto-set by cfgate",
					Port:        "443",
					Protocol:    tencentcloud.TCP,
				},
			}

			rules = append(rules, newRules...)
		}

		if cmdArgs.IncludeIPv6 {
			for _, ipv6Cidr := range ipv6Cidrs {
				newRules := []*firewall.Rule{
					{
						Action:      tencentcloud.ACCEPT,
						CidrBlock:   ipv6Cidr,
						Description: "Auto-set by cfgate",
						Port:        "80",
						Protocol:    tencentcloud.TCP,
					},
					{
						Action:      tencentcloud.ACCEPT,
						CidrBlock:   ipv6Cidr,
						Description: "Auto-set by cfgate",
						Port:        "443",
						Protocol:    tencentcloud.TCP,
					},
				}

				rules = append(rules, newRules...)
			}
		}

		secretID := os.Getenv("TENCENTCLOUD_SECRET_ID")
		secretKey := os.Getenv("TENCENTCLOUD_SECRET_KEY")

		if secretID == "" || secretKey == "" {
			return fmt.Errorf("Invalid tencent cloud credential")
		}

		tc, err := tencentcloud.NewTencentCloud(secretID, secretKey, cmdArgs.Region, cmdArgs.InstanceID)
		if err != nil {
			return err
		}

		ctxListRules, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()
		currentRules, err := tc.ListRules(ctxListRules)
		if err != nil {
			return err
		}

		for _, currentRule := range currentRules {
			if currentRule.Protocol == tencentcloud.TCP && (currentRule.Port == "80" || currentRule.Port == "443") {
				continue
			}
			rules = append(rules, currentRule)
		}

		ctxApplyRules, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()
		err = tc.ApplyRules(ctxApplyRules, rules)

		return err
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
