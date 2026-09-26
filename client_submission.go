package jmap

import (
	"encoding/json"
	"fmt"
)

// Send creates an EmailSubmission, which dispatches the referenced Email for delivery.
// The server returns the created EmailSubmission with server‑assigned fields populated.
func (c *JmapClient) Send(accountId string, sub EmailSubmission) (EmailSubmission, error) {
	// Use a fixed creation id for the single‑call request.
	const createId = "c1"

	args := map[string]interface{}{
		"accountId": accountId,
		"create": map[string]EmailSubmission{
			createId: sub,
		},
	}
	inv := Invocation{
		Name:        "EmailSubmission/set",
		Arguments:   args,
		MethodCallId: "c1",
	}
	respEnv, err := c.SendRequest([]Invocation{inv}, []string{"urn:ietf:params:jmap:submission"})
	if err != nil {
		return EmailSubmission{}, err
	}
	if len(respEnv.MethodResponses) == 0 {
		return EmailSubmission{}, fmt.Errorf("no methodResponses in Send result")
	}
	mr := respEnv.MethodResponses[0]
	createdRaw, ok := mr.Arguments["created"]
	if !ok {
		return EmailSubmission{}, fmt.Errorf("missing created field in Send response")
	}
	createdMap, ok := createdRaw.(map[string]interface{})
	if !ok {
		return EmailSubmission{}, fmt.Errorf("invalid created field type in Send response")
	}
	createdEntry, ok := createdMap[createId]
	if !ok {
		return EmailSubmission{}, fmt.Errorf("created entry for id %s not found", createId)
	}
	b, err := json.Marshal(createdEntry)
	if err != nil {
		return EmailSubmission{}, err
	}
	var created EmailSubmission
	if err := json.Unmarshal(b, &created); err != nil {
		return EmailSubmission{}, err
	}
	return created, nil
}

// CancelSend updates an existing EmailSubmission to cancel its delivery.
// It sets the "undoStatus" property to "canceled".
func (c *JmapClient) CancelSend(accountId string, submissionId string) error {
	update := map[string]interface{}{
		"undoStatus": "canceled",
	}
	args := map[string]interface{}{
		"accountId": accountId,
		"update": map[string]map[string]interface{}{
			submissionId: update,
		},
	}
	inv := Invocation{
		Name:        "EmailSubmission/set",
		Arguments:   args,
		MethodCallId: "c1",
	}
	respEnv, err := c.SendRequest([]Invocation{inv}, []string{"urn:ietf:params:jmap:submission"})
	if err != nil {
		return err
	}
	if len(respEnv.MethodResponses) == 0 {
		return fmt.Errorf("no methodResponses in CancelSend result")
	}
	// The response may contain "updated" or per‑id errors; we treat lack of error as success.
	return nil
}

// ListSubmissions retrieves EmailSubmission objects for the given account.
// If ids is non‑empty, only those submissions are returned.
// If properties is non‑empty, only the specified properties are included in each result.
func (c *JmapClient) ListSubmissions(accountId string, ids []string, properties []string) ([]EmailSubmission, error) {
	args := map[string]interface{}{
		"accountId": accountId,
	}
	if len(ids) > 0 {
		args["ids"] = ids
	}
	if len(properties) > 0 {
		args["properties"] = properties
	}
	inv := Invocation{
		Name:        "EmailSubmission/get",
		Arguments:   args,
		MethodCallId: "c1",
	}
	respEnv, err := c.SendRequest([]Invocation{inv}, []string{"urn:ietf:params:jmap:submission"})
	if err != nil {
		return nil, err
	}
	if len(respEnv.MethodResponses) == 0 {
		return nil, fmt.Errorf("no methodResponses in ListSubmissions result")
	}
	mr := respEnv.MethodResponses[0]
	listRaw, ok := mr.Arguments["list"]
	if !ok {
		return nil, fmt.Errorf("missing list field in ListSubmissions response")
	}
	listSlice, ok := listRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid list field type in ListSubmissions response")
	}
	submissions := make([]EmailSubmission, 0, len(listSlice))
	for _, item := range listSlice {
		b, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		var sub EmailSubmission
		if err := json.Unmarshal(b, &sub); err != nil {
			return nil, err
		}
		submissions = append(submissions, sub)
	}
	return submissions, nil
}
