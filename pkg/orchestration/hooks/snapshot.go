package hooks

import (
	"strconv"

	"github.com/elee1766/jobshop/pkg/domain"
)

// StepInput is a step ticket as a callback sees it.
func StepInput(t domain.Ticket) (Step, error) {
	f, err := domain.DecodeFields[domain.StepFields](t)
	return Step{ID: f.Step, TicketID: t.ID, Pool: t.Pool, Status: string(t.Status), Resolution: t.Resolution, Outcome: f.Outcome, Output: f.StructuredOutput, Inputs: f.Inputs, Failures: f.Failures, Retries: f.Retries}, err
}

// RunSnapshot is a run and its steps as a callback sees them, and the
// revision of the run and of every child (steps or not) that was read. The
// host checks that read set again before it commits what the callback
// returned, so a change in between supersedes the invocation.
func RunSnapshot(run domain.Ticket, f domain.RunFields, children []domain.Ticket) (*Run, map[string]int64, error) {
	in := &Run{ID: run.ID, Workflow: f.Workflow, Key: f.Key, Project: run.Project, Inputs: f.Inputs, Vars: f.Vars, State: f.State, Steps: []Step{}}
	readSet := map[string]int64{run.ID: run.Revision}
	for _, child := range children {
		readSet[child.ID] = child.Revision
		if child.Kind != domain.KindStep {
			continue
		}
		step, err := StepInput(child)
		if err != nil {
			return nil, nil, err
		}
		in.Steps = append(in.Steps, step)
	}
	return in, readSet, nil
}

// RunDue is the invocation a running callback workflow run is waiting
// for, if any: on_start until the first batch is applied, then
// on_step_completed for the first finished step whose current revision was
// not completed yet. Its key is stable for the same run, step revision and
// automation version, so enqueueing again is harmless (Put deduplicates).
// parent is the invocation that started the run, for the audit chain.
// ok is false when nothing is due. run must have f.Automation set.
func RunDue(run domain.Ticket, f domain.RunFields, children []domain.Ticket, parent string) (key string, in Invocation, ok bool, err error) {
	a := f.Automation
	input, readSet, err := RunSnapshot(run, f, children)
	if err != nil {
		return "", Invocation{}, false, err
	}
	in = Invocation{Run: run.ID, RunInput: input, ReadSet: readSet, Digest: a.Digest, ParentInvocation: parent}
	key = run.ID + "/start"
	if a.Version == 0 {
		in.Entry = a.OnStart
	} else {
		for _, child := range children {
			if child.Kind != domain.KindStep || !domain.StepFinished(child) || a.Completed[child.ID] == child.Revision {
				continue
			}
			step, err := StepInput(child)
			if err != nil {
				return "", Invocation{}, false, err
			}
			in.Entry, in.Step, in.StepRevision, in.StepInput = a.OnStepCompleted, child.ID, child.Revision, &step
			key = run.ID + "/step/" + child.ID + "/" + strconv.FormatInt(child.Revision, 10) + "/" + strconv.FormatInt(a.Version, 10)
			break
		}
	}
	return key, in, in.Entry != "", nil
}
