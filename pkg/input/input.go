package input

import "time"

const (
	datetimeRegexISO8601  = `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{9}Z`
	datetimeRegexJournalD = `^[A-Z][a-z]{2} \d{1,2} \d{2}:\d{2}:\d{2}`
	headerContentType     = "Content-Type"
	jsonContentHeader     = "application/json"
	batchSize             = 20
)

type LogMessage struct {
	Time           time.Time `json:"time,omitempty"`
	Log            string    `json:"log,omitempty"`
	Agent          string    `json:"agent,omitempty"`
	IsControlPlane bool      `json:"is_control_plane_log,omitempty"`
	Component      string    `json:"kubernetes_component,omitempty"`
}

type ComponentInput interface {
	Publish(endpoint string) error // Publish should read the contents of component logs and publish them to the payload endpoint.
	ComponentName() string
}
