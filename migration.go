package concept

type Migration struct {
	Version       string
	Description   string
	AppliedBy     string
	AppliedAt     uint64
	ExecutionTime uint32
	State         state

	AdvanceScript *Script
	ReverseScript *Script
}
