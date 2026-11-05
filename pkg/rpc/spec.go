package rpc

import (
	"maps"
	"slices"

	"github.com/elee1766/jobshop/pkg/store"
)

// Spec is one admin method's parameter and result types, as zero values.
// Desk's GET /api/rpc turns them into JSON schemas and cmd/deskgen into
// desk-plugin.d.ts (#542). Nil Result means the method answers {}.
type Spec struct {
	Params any
	Result any
}

type empty struct{}

// Specs is every admin method. Methods is its keys; a control test keeps
// them equal to the controller's handler table, so a new method without
// an entry here fails the tests.
var Specs = map[string]Spec{
	MethodTicketCreate:      {CreateParams{}, store.Ticket{}},
	MethodTicketShow:        {IDParams{}, ShowResult{}},
	MethodTicketTrace:       {TraceParams{}, TraceNode{}},
	MethodTicketList:        {ListParams{}, []store.Ticket{}},
	MethodTicketUpdate:      {UpdateParams{}, store.Ticket{}},
	MethodTicketClose:       {CloseParams{}, store.Ticket{}},
	MethodTicketReopen:      {IDParams{}, store.Ticket{}},
	MethodReady:             {ReadyParams{}, []store.Ticket{}},
	MethodCommentAdd:        {CommentParams{}, store.Comment{}},
	MethodEdgeAdd:           {EdgeParams{}, nil},
	MethodEdgeRemove:        {EdgeParams{}, nil},
	MethodLabelAdd:          {LabelParams{}, nil},
	MethodLabelRemove:       {LabelParams{}, nil},
	MethodClaim:             {ClaimParams{}, ClaimResult{}},
	MethodHeartbeat:         {IDParams{}, store.Claim{}},
	MethodRelease:           {ReleaseParams{}, store.Ticket{}},
	MethodHandoff:           {HandoffParams{}, store.Ticket{}},
	MethodSessionNew:        {SessionNewParams{}, SessionNewResult{}},
	MethodSessionList:       {SessionListParams{}, []SessionInfo{}},
	MethodSessionReap:       {SessionReapParams{}, SessionReapResult{}},
	MethodSessionRestart:    {SessionRestartParams{}, SessionRestartResult{}},
	MethodSessionPrompt:     {SessionPromptParams{}, SessionPromptResult{}},
	MethodSessionSandbox:    {SessionSandboxParams{}, SessionSandboxResult{}},
	MethodSessionSend:       {SessionSendParams{}, SessionSendResult{}},
	MethodSessionSubscribe:  {SessionSubscribeParams{}, SessionSubscribeResult{}},
	MethodSessionJournal:    {SessionJournalParams{}, SessionJournalResult{}},
	MethodSessionCordon:     {SessionCordonParams{}, SessionInfo{}},
	MethodSessionConfig:     {SessionConfigParams{}, SessionInfo{}},
	MethodWhoami:            {empty{}, Whoami{}},
	MethodVersion:           {empty{}, VersionResult{}},
	MethodShutdown:          {ShutdownParams{}, nil},
	MethodDrain:             {DrainParams{}, DrainResult{}},
	MethodReload:            {empty{}, ReloadResult{}},
	MethodTriggerList:       {empty{}, []TriggerInfo{}},
	MethodConfigShow:        {empty{}, ConfigShowResult{}},
	MethodRunStart:          {RunStartParams{}, RunResult{}},
	MethodRunShow:           {IDParams{}, RunResult{}},
	MethodRunList:           {RunListParams{}, []store.Ticket{}},
	MethodRunCancel:         {IDParams{}, RunResult{}},
	MethodRunResume:         {RunResumeParams{}, RunResult{}},
	MethodBatchStart:        {BatchStartParams{}, BatchResult{}},
	MethodBatchShow:         {IDParams{}, BatchResult{}},
	MethodBatchList:         {BatchListParams{}, []store.Ticket{}},
	MethodPoolList:          {empty{}, PoolStatusResult{}},
	MethodPoolScale:         {PoolScaleParams{}, PoolStatusResult{}},
	MethodPoolUnquarantine:  {PoolParams{}, PoolStatusResult{}},
	MethodReportIncidents:   {IncidentsParams{}, IncidentsResult{}},
	MethodGuard:             {GuardParams{}, GuardResult{}},
	MethodRecall:            {RecallParams{}, []store.RecallHit{}},
	MethodAttention:         {empty{}, AttentionResult{}},
	MethodIdentityList:      {empty{}, []IdentityInfo{}},
	MethodIdentityCheck:     {IdentityParams{}, IdentityCheckResult{}},
	MethodIdentityToken:     {IdentityParams{}, IdentityTokenResult{}},
	MethodIngestDrain:       {IngestDrainParams{}, IngestDrainResult{}},
	MethodIngestClaim:       {IngestClaimParams{}, IngestClaimResult{}},
	MethodIngestAdmit:       {IngestAdmitParams{}, store.InboxItem{}},
	MethodIngestWorkers:     {empty{}, IngestWorkersResult{}},
	MethodSubscribe:         {SubscribeParams{}, store.Ticket{}},
	MethodReport:            {ReportParams{}, ReportResult{}},
	MethodQuestionAnswer:    {AnswerParams{}, AnswerResult{}},
	MethodActivityList:      {ActivityListParams{}, ActivityListResult{}},
	MethodActivityRecord:    {ActivityRecordParams{}, nil},
	MethodLearnAdd:          {LearnParams{}, store.Learning{}},
	MethodLearnList:         {LearnListParams{}, []store.Learning{}},
	MethodLearnUpdate:       {LearnUpdateParams{}, []store.Learning{}},
	MethodLearnEvaluate:     {LearnEvaluateParams{}, []store.MemoryEdit{}},
	MethodLearnEdits:        {LearnEditsParams{}, []store.MemoryEdit{}},
	MethodWriteKeyClaim:     {WriteKeyParams{}, store.WriteKey{}},
	MethodWriteKeyFinish:    {WriteKeyParams{}, store.WriteKey{}},
	MethodEventsSubscribe:   {EventsSubscribeParams{}, EventsSubscribeResult{}},
	MethodWaitAdd:           {WaitAddParams{}, store.Ticket{}},
	MethodWaitList:          {WaitListParams{}, []WaitItem{}},
	MethodWaitClear:         {WaitClearParams{}, store.Ticket{}},
	MethodResourceAdd:       {SlotParams{}, store.Slot{}},
	MethodResourceRemove:    {SlotParams{}, store.Slot{}},
	MethodResourceUnhealthy: {SlotParams{}, SlotResult{}},
	MethodResourceList:      {empty{}, ResourceListResult{}},
	MethodArtifactDeliver:   {ArtifactDeliverParams{}, Artifact{}},
	MethodArtifactList:      {ArtifactListParams{}, []Artifact{}},
	MethodArtifactShow:      {IDParams{}, ArtifactShowResult{}},
	MethodArtifactAck:       {ArtifactAckParams{}, Artifact{}},
	MethodDeskExec:          {DeskExecParams{}, DeskExecResult{}},
}

// Methods lists every admin method name, sorted. Desk's generic RPC
// endpoint checks against it (#517).
func Methods() []string {
	return slices.Sorted(maps.Keys(Specs))
}
