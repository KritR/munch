package protocol

type Suggestion struct {
	Command              string   `json:"command"`
	Description          string   `json:"description"`
	Risk                 string   `json:"risk"`
	RequiresConfirmation bool     `json:"requires_confirmation"`
	ConfirmationReason   string   `json:"confirmation_reason,omitempty"`
	Assumptions          []string `json:"assumptions,omitempty"`
	UsesTools            []string `json:"uses_tools,omitempty"`
	Confidence           *float64 `json:"confidence,omitempty"`
}
