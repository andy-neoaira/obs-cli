package protocol

type PlanChange struct {
	Action   string         `json:"action"`
	Resource string         `json:"resource"`
	Target   string         `json:"target"`
	Details  map[string]any `json:"details,omitempty"`
}

type Plan struct {
	Changes       []PlanChange `json:"changes"`
	Risks         []string     `json:"risks"`
	Preconditions []string     `json:"preconditions"`
}

type DryRunData struct {
	DryRun  bool `json:"dry_run"`
	Applied bool `json:"applied"`
	Changed bool `json:"changed"`
	Plan    Plan `json:"plan"`
}

func NewDryRunData(changes []PlanChange, risks, preconditions []string) DryRunData {
	if changes == nil {
		changes = []PlanChange{}
	}
	if risks == nil {
		risks = []string{}
	}
	if preconditions == nil {
		preconditions = []string{}
	}
	return DryRunData{
		DryRun:  true,
		Applied: false,
		Changed: len(changes) != 0,
		Plan: Plan{
			Changes:       changes,
			Risks:         risks,
			Preconditions: preconditions,
		},
	}
}
