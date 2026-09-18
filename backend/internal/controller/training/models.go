package training

import "encoding/json"

type startRequest struct {
	ScenarioID string `json:"scenario_id"`
	Mode       string `json:"mode"`
	Channel    string `json:"channel"`
	Preview    bool   `json:"preview"`
}
type actionRequest struct {
	ActionID      string `json:"action_id"`
	ClientEventID string `json:"client_event_id"`
}
type messageRequest struct {
	Text      string `json:"text"`
	MessageID string `json:"message_id"`
}
type nodeResponse struct {
	NodeID           string `json:"node_id"`
	Speaker          string `json:"speaker"`
	Text             string `json:"text"`
	AvailableActions any    `json:"available_actions"`
}
type messageResponse struct {
	MessageID string `json:"message_id"`
	Author    string `json:"author"`
	Text      string `json:"text"`
	CreatedAt any    `json:"created_at"`
}
type sessionResponse struct {
	SessionID       string            `json:"session_id"`
	ScenarioID      string            `json:"scenario_id"`
	ScenarioVersion string            `json:"scenario_version"`
	Mode            string            `json:"mode"`
	Channel         string            `json:"channel"`
	Status          string            `json:"status"`
	CurrentNode     any               `json:"current_node"`
	Messages        []messageResponse `json:"messages"`
}
type actionResponse struct {
	EventID  string `json:"event_id"`
	NextNode any    `json:"next_node"`
	Status   string `json:"status"`
}
type trainingMessageResponse struct {
	LearnerMessage          messageResponse `json:"learner_message"`
	PartnerMessage          messageResponse `json:"partner_message"`
	DetectedSemanticActions []string        `json:"detected_semantic_actions"`
}
type completeResponse struct {
	Outcome json.RawMessage `json:"outcome"`
}
