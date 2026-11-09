package firewall

import "context"

type Action string

type Protocol string

type Rule struct {
	Action      Action
	CidrBlock   string
	Description string
	Port        string
	Protocol    Protocol
}

type Firewall interface {
	ListRules(ctx context.Context) ([]*Rule, error)
	ApplyRules(ctx context.Context, rules []*Rule) error
}
