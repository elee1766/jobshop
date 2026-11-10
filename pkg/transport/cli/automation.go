package cli

import "github.com/gfx-labs/jobshop/pkg/transport/rpc"

type automationCmd struct {
	List automationListCmd `cmd:"" help:"List durable hook invocations, newest first (admin)."`
	Show automationShowCmd `cmd:"" help:"Show a hook invocation, attempts and ordered external-call evidence (admin)."`
}

type automationListCmd struct {
	RunID   string `name:"run" help:"Filter by workflow run ticket ID."`
	Trigger string `help:"Filter by trigger name."`
	State   string `help:"Filter by invocation state."`
	Limit   int    `help:"Maximum invocations to return." default:"50"`
}

func (c automationListCmd) Run(g *Globals) error {
	return g.call(rpc.MethodAutomationList, rpc.AutomationListParams{
		Run: c.RunID, Trigger: c.Trigger, State: c.State, Limit: c.Limit,
	}, nil)
}

type automationShowCmd struct {
	ID string `arg:"" help:"Durable hook invocation ID."`
}

func (c automationShowCmd) Run(g *Globals) error {
	return g.call(rpc.MethodAutomationShow, rpc.IDParams{ID: c.ID}, nil)
}
