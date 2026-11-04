# TestStepFailureSources on a fake clock (#375)

2026-11-04. `TestStepFailureSources` runs its controller on `clock.Fake`
(`runRootLease` takes a clock). The 2s lease cannot expire during steps
1 and 2 because time only moves when step 3 advances the clock past it;
supervisor passes then sweep the expired lease. Under `-race -count=5`
the old wall-clock lease could still expire mid-step.
