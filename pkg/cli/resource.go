package cli

import "github.com/elee1766/jobshop/pkg/rpc"

// resourceCmd is the provider side of resource slots (#407): a provider
// command or service calls these on the admin socket.
type resourceCmd struct {
	Add       resourceAddCmd       `cmd:"" help:"Add a ready slot to a provider resource, or update its env (admin or pool grant)."`
	Remove    resourceRemoveCmd    `cmd:"" help:"Remove a slot; a held slot goes when its claim ends (admin or pool grant)."`
	Unhealthy resourceUnhealthyCmd `cmd:"" help:"Mark a slot gone; its claim is requeued without counting a failure (admin or pool grant)."`
	List      resourceListCmd      `cmd:"" help:"List slots and resource occupancy (admin or pool grant)."`
}

type resourceAddCmd struct {
	Resource string            `arg:"" help:"Resource name."`
	ID       string            `arg:"" help:"Slot id."`
	Env      map[string]string `help:"Env the claim gets, k=v (repeatable)."`
}

func (c resourceAddCmd) Run(g *Globals) error {
	return g.call(rpc.MethodResourceAdd, rpc.SlotParams{Resource: c.Resource, Slot: c.ID, Env: c.Env}, nil)
}

type resourceRemoveCmd struct {
	Resource string `arg:"" help:"Resource name."`
	ID       string `arg:"" help:"Slot id."`
}

func (c resourceRemoveCmd) Run(g *Globals) error {
	return g.call(rpc.MethodResourceRemove, rpc.SlotParams{Resource: c.Resource, Slot: c.ID}, nil)
}

type resourceUnhealthyCmd struct {
	Resource string `arg:"" help:"Resource name."`
	ID       string `arg:"" help:"Slot id."`
	Reason   string `short:"r" help:"Why, for the requeued ticket's comment."`
}

func (c resourceUnhealthyCmd) Run(g *Globals) error {
	return g.call(rpc.MethodResourceUnhealthy, rpc.SlotParams{Resource: c.Resource, Slot: c.ID, Reason: c.Reason}, nil)
}

type resourceListCmd struct{}

func (resourceListCmd) Run(g *Globals) error {
	return g.call(rpc.MethodResourceList, nil, nil)
}
