package presence

import "time"

type UserPresence struct {
	InstanceID     string     `json:"instance_id"`
	Status         string     `json:"status"` // "online" or "offline"
	UpdatedAt      time.Time  `json:"updated_at"`
	DisconnectedAt *time.Time `json:"disconnected_at,omitempty"`
}

type TargetedEvent struct {
	UserID       string            `json:"user_id"`
	Type         string            `json:"type"`
	Payload      string            `json:"payload"`
	TraceContext map[string]string `json:"trace_context,omitempty"`
}
