package rpc

import "github.com/gfx-labs/jobshop/pkg/storage/store"

// AutomationListParams selects durable hook invocations for the operator.
type AutomationListParams struct {
	Run     string `json:"run,omitempty"`
	Trigger string `json:"trigger,omitempty"`
	State   string `json:"state,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

// AutomationResult links an event snapshot to attempts, calls and planned actions.
type AutomationResult struct {
	Invocation store.Record   `json:"invocation"`
	Attempts   []store.Record `json:"attempts"`
	Calls      []store.Record `json:"calls"`
	Truncated  bool           `json:"truncated,omitempty"`
}
