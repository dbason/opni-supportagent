package input

import "time"

const (
	headerContentType = "Content-Type"
	jsonContentHeader = "application/json"
	batchSize         = 20
	KlogRegex         = `^[I,E,F]\d{4} \d{2}:\d{2}:\d{2}.\d{6}`
)

type LogMessage struct {
	Timestamp      time.Time `json:"timestamp,omitempty"`
	Time           time.Time `json:"time,omitempty"`
	Log            string    `json:"log,omitempty"`
	Agent          string    `json:"agent,omitempty"`
	IsControlPlane bool      `json:"is_control_plane_log"`
	Component      string    `json:"kubernetes_component,omitempty"`
	ClusterID      string    `json:"cluster_id,omitempty"`
	NodeName       string    `json:"node_name,omitempty"`
	Processed      bool      `json:"processed"`
}

type ComponentInput interface {
	Publish(parser DateParser) (time.Time, time.Time, error) // Publish should read the contents of component logs and publish them to the payload endpoint.
	ComponentName() string
}

type DateParser interface {
	ParseTimestamp(log string) (time.Time, string, bool) // Parse timestamp should have the implementation for parsing the timestamp from a log line
}
