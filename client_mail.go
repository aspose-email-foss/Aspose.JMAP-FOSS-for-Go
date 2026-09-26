package jmap

import (
	"encoding/json"
	"fmt"
)

// MailboxGetResponse represents the response to a Mailbox/get call.
type MailboxGetResponse struct {
	AccountId string   `json:"accountId"`
	State     string   `json:"state"`
	List      []Mailbox `json:"list"`
	NotFound []string `json:"notFound,omitempty"`
}

// MailboxSetResponse represents the response to a Mailbox/set call.
type MailboxSetResponse struct {
	AccountId    string               `json:"accountId"`
	OldState     *string              `json:"oldState,omitempty"`
	NewState     string               `json:"newState"`
	Created      map[string]Mailbox   `json:"created,omitempty"`
	Updated      map[string]*Mailbox  `json:"updated,omitempty"`
	Destroyed    []string             `json:"destroyed,omitempty"`
	NotCreated   map[string]SetError  `json:"notCreated,omitempty"`
	NotUpdated   map[string]SetError  `json:"notUpdated,omitempty"`
	NotDestroyed map[string]SetError  `json:"notDestroyed,omitempty"`
}

// EmailQueryResponse represents the response to an Email/query call.
type EmailQueryResponse struct {
	AccountId            string   `json:"accountId"`
	QueryState           string   `json:"queryState"`
	CanCalculateChanges bool     `json:"canCalculateChanges"`
	Position             int      `json:"position"`
	Ids                  []string `json:"ids"`
	Total                *int     `json:"total,omitempty"`
	Limit                *int     `json:"limit,omitempty"`
}

// EmailGetResponse represents the response to an Email/get call.
type EmailGetResponse struct {
	AccountId string  `json:"accountId"`
	State     string  `json:"state"`
	List      []Email `json:"list"`
	NotFound  []string `json:"notFound,omitempty"`
}

// EmailSetResponse represents the response to an Email/set call.
type EmailSetResponse struct {
	AccountId    string               `json:"accountId"`
	OldState     *string              `json:"oldState,omitempty"`
	NewState     string               `json:"newState"`
	Created      map[string]Email     `json:"created,omitempty"`
	Updated      map[string]*Email    `json:"updated,omitempty"`
	Destroyed    []string             `json:"destroyed,omitempty"`
	NotCreated   map[string]SetError  `json:"notCreated,omitempty"`
	NotUpdated   map[string]SetError  `json:"notUpdated,omitempty"`
	NotDestroyed map[string]SetError  `json:"notDestroyed,omitempty"`
}

// SearchSnippetGetResponse represents the response to a SearchSnippet/get call.
type SearchSnippetGetResponse struct {
	AccountId string          `json:"accountId"`
	List      []SearchSnippet `json:"list"`
	NotFound  []string        `json:"notFound,omitempty"`
}

// ThreadGetResponse represents the response to a Thread/get call.
type ThreadGetResponse struct {
	AccountId string   `json:"accountId"`
	State     string   `json:"state"`
	List      []Thread `json:"list"`
	NotFound  []string `json:"notFound,omitempty"`
}

// IdentityGetResponse represents the response to an Identity/get call.
type IdentityGetResponse struct {
	AccountId string     `json:"accountId"`
	State     string     `json:"state"`
	List      []Identity `json:"list"`
	NotFound  []string   `json:"notFound,omitempty"`
}

// IdentitySetResponse represents the response to an Identity/set call.
type IdentitySetResponse struct {
	AccountId    string               `json:"accountId"`
	OldState     *string              `json:"oldState,omitempty"`
	NewState     string               `json:"newState"`
	Created      map[string]Identity  `json:"created,omitempty"`
	Updated      map[string]*Identity `json:"updated,omitempty"`
	Destroyed    []string             `json:"destroyed,omitempty"`
	NotCreated   map[string]SetError  `json:"notCreated,omitempty"`
	NotUpdated   map[string]SetError  `json:"notUpdated,omitempty"`
	NotDestroyed map[string]SetError  `json:"notDestroyed,omitempty"`
}

// ListMailboxes returns all mailboxes for the given account.
func (c *JmapClient) ListMailboxes(accountId string) (MailboxGetResponse, error) {
	inv := Invocation{
		Name:      "Mailbox/get",
		Arguments: map[string]interface{}{"accountId": accountId},
		MethodCallId: "c1",
	}
	env, err := c.SendRequest([]Invocation{inv}, []string{"urn:ietf:params:jmap:mail"})
	if err != nil {
		return MailboxGetResponse{}, err
	}
	if len(env.MethodResponses) == 0 {
		return MailboxGetResponse{}, fmt.Errorf("no methodResponses for Mailbox/get")
	}
	b, _ := json.Marshal(env.MethodResponses[0].Arguments)
	var res MailboxGetResponse
	if err := json.Unmarshal(b, &res); err != nil {
		return MailboxGetResponse{}, err
	}
	return res, nil
}

// GetMailbox retrieves specific mailboxes by id.
func (c *JmapClient) GetMailbox(accountId string, ids []string, properties []string) (MailboxGetResponse, error) {
	args := map[string]interface{}{
		"accountId": accountId,
		"ids":       ids,
	}
	if len(properties) > 0 {
		args["properties"] = properties
	}
	inv := Invocation{
		Name:        "Mailbox/get",
		Arguments:   args,
		MethodCallId: "c1",
	}
	env, err := c.SendRequest([]Invocation{inv}, []string{"urn:ietf:params:jmap:mail"})
	if err != nil {
		return MailboxGetResponse{}, err
	}
	if len(env.MethodResponses) == 0 {
		return MailboxGetResponse{}, fmt.Errorf("no methodResponses for Mailbox/get")
	}
	b, _ := json.Marshal(env.MethodResponses[0].Arguments)
	var res MailboxGetResponse
	if err := json.Unmarshal(b, &res); err != nil {
		return MailboxGetResponse{}, err
	}
	return res, nil
}

// CreateMailbox creates new mailboxes.
func (c *JmapClient) CreateMailbox(accountId string, create map[string]Mailbox) (MailboxSetResponse, error) {
	args := map[string]interface{}{
		"accountId": accountId,
		"create":    create,
	}
	inv := Invocation{
		Name:        "Mailbox/set",
		Arguments:   args,
		MethodCallId: "c1",
	}
	env, err := c.SendRequest([]Invocation{inv}, []string{"urn:ietf:params:jmap:mail"})
	if err != nil {
		return MailboxSetResponse{}, err
	}
	if len(env.MethodResponses) == 0 {
		return MailboxSetResponse{}, fmt.Errorf("no methodResponses for Mailbox/set")
	}
	b, _ := json.Marshal(env.MethodResponses[0].Arguments)

	// Decode "created" separately as raw JSON first: per RFC 8620 section 5.3, each
	// entry is PARTIAL (the server only includes fields it assigned/defaulted,
	// typically just the new id), not the full object - it must be merged onto the
	// client-submitted object (server fields win) BEFORE unmarshaling into Mailbox,
	// since encoding/json would otherwise silently zero-fill required fields (e.g.
	// Name) that the server had no reason to repeat back.
	var rawArgs struct {
		Created map[string]json.RawMessage `json:"created,omitempty"`
	}
	if err := json.Unmarshal(b, &rawArgs); err != nil {
		return MailboxSetResponse{}, err
	}

	var res MailboxSetResponse
	if err := json.Unmarshal(b, &res); err != nil {
		return MailboxSetResponse{}, err
	}
	if res.Created != nil {
		res.Created = make(map[string]Mailbox, len(rawArgs.Created))
		for id, rawEntry := range rawArgs.Created {
			merged := create[id] // start from the client-submitted object
			if err := json.Unmarshal(rawEntry, &merged); err != nil {
				return MailboxSetResponse{}, err
			}
			// merged.Id is already set correctly by the Unmarshal above (the server's
			// "created" entry always includes the real server-assigned id) - do NOT
			// overwrite it with the map key, which is the client-chosen CREATION id
			// (e.g. "new1"), a different, unrelated identifier.
			res.Created[id] = merged
		}
	}
	return res, nil
}

// DeleteMailbox removes mailboxes.
func (c *JmapClient) DeleteMailbox(accountId string, destroy []string) (MailboxSetResponse, error) {
	args := map[string]interface{}{
		"accountId": accountId,
		"destroy":   destroy,
	}
	inv := Invocation{
		Name:        "Mailbox/set",
		Arguments:   args,
		MethodCallId: "c1",
	}
	env, err := c.SendRequest([]Invocation{inv}, []string{"urn:ietf:params:jmap:mail"})
	if err != nil {
		return MailboxSetResponse{}, err
	}
	if len(env.MethodResponses) == 0 {
		return MailboxSetResponse{}, fmt.Errorf("no methodResponses for Mailbox/set")
	}
	b, _ := json.Marshal(env.MethodResponses[0].Arguments)
	var res MailboxSetResponse
	if err := json.Unmarshal(b, &res); err != nil {
		return MailboxSetResponse{}, err
	}
	return res, nil
}

// ListMessages performs an Email/query to list messages.
func (c *JmapClient) ListMessages(accountId string, filter map[string]interface{}, sort []Comparator, limit *int) (EmailQueryResponse, error) {
	args := map[string]interface{}{
		"accountId": accountId,
	}
	if filter != nil {
		args["filter"] = filter
	}
	if len(sort) > 0 {
		args["sort"] = sort
	}
	if limit != nil {
		args["limit"] = *limit
	}
	inv := Invocation{
		Name:        "Email/query",
		Arguments:   args,
		MethodCallId: "c1",
	}
	env, err := c.SendRequest([]Invocation{inv}, []string{"urn:ietf:params:jmap:mail"})
	if err != nil {
		return EmailQueryResponse{}, err
	}
	if len(env.MethodResponses) == 0 {
		return EmailQueryResponse{}, fmt.Errorf("no methodResponses for Email/query")
	}
	b, _ := json.Marshal(env.MethodResponses[0].Arguments)
	var res EmailQueryResponse
	if err := json.Unmarshal(b, &res); err != nil {
		return EmailQueryResponse{}, err
	}
	return res, nil
}

// FetchMessage retrieves full messages by id.
func (c *JmapClient) FetchMessage(accountId string, ids []string, properties []string) (EmailGetResponse, error) {
	args := map[string]interface{}{
		"accountId": accountId,
		"ids":       ids,
	}
	if len(properties) > 0 {
		args["properties"] = properties
	}
	inv := Invocation{
		Name:        "Email/get",
		Arguments:   args,
		MethodCallId: "c1",
	}
	env, err := c.SendRequest([]Invocation{inv}, []string{"urn:ietf:params:jmap:mail"})
	if err != nil {
		return EmailGetResponse{}, err
	}
	if len(env.MethodResponses) == 0 {
		return EmailGetResponse{}, fmt.Errorf("no methodResponses for Email/get")
	}
	b, _ := json.Marshal(env.MethodResponses[0].Arguments)
	var res EmailGetResponse
	if err := json.Unmarshal(b, &res); err != nil {
		return EmailGetResponse{}, err
	}
	return res, nil
}

// MoveMessage moves the given messages to destinationMailboxId. Deliberately typed
// (accountId, ids, destinationMailboxId) rather than a raw wire-protocol
// map[string]map[string]bool: a caller should not need to know the JMAP PatchObject shape
// ({emailId: {"mailboxIds": {mailboxId: true}}}) to move a message, and every other
// language target in this project (Python, TypeScript, Rust) exposes a typed signature
// here too - a raw map leaks wire-protocol detail and breaks API parity across languages.
func (c *JmapClient) MoveMessage(accountId string, ids []string, destinationMailboxId string) (EmailSetResponse, error) {
	update := make(map[string]interface{}, len(ids))
	for _, id := range ids {
		update[id] = map[string]interface{}{
			"mailboxIds": map[string]bool{destinationMailboxId: true},
		}
	}
	args := map[string]interface{}{
		"accountId": accountId,
		"update":    update,
	}
	inv := Invocation{
		Name:        "Email/set",
		Arguments:   args,
		MethodCallId: "c1",
	}
	env, err := c.SendRequest([]Invocation{inv}, []string{"urn:ietf:params:jmap:mail"})
	if err != nil {
		return EmailSetResponse{}, err
	}
	if len(env.MethodResponses) == 0 {
		return EmailSetResponse{}, fmt.Errorf("no methodResponses for Email/set")
	}
	b, _ := json.Marshal(env.MethodResponses[0].Arguments)
	var res EmailSetResponse
	if err := json.Unmarshal(b, &res); err != nil {
		return EmailSetResponse{}, err
	}
	return res, nil
}

// SetMessageKeyword sets or clears keyword on the given messages. Deliberately typed
// (accountId, ids, keyword, value) rather than a raw wire-protocol
// map[string]map[string]bool - see the note on MoveMessage.
func (c *JmapClient) SetMessageKeyword(accountId string, ids []string, keyword string, value bool) (EmailSetResponse, error) {
	update := make(map[string]interface{}, len(ids))
	for _, id := range ids {
		update[id] = map[string]interface{}{
			"keywords": map[string]bool{keyword: value},
		}
	}
	args := map[string]interface{}{
		"accountId": accountId,
		"update":    update,
	}
	inv := Invocation{
		Name:        "Email/set",
		Arguments:   args,
		MethodCallId: "c1",
	}
	env, err := c.SendRequest([]Invocation{inv}, []string{"urn:ietf:params:jmap:mail"})
	if err != nil {
		return EmailSetResponse{}, err
	}
	if len(env.MethodResponses) == 0 {
		return EmailSetResponse{}, fmt.Errorf("no methodResponses for Email/set")
	}
	b, _ := json.Marshal(env.MethodResponses[0].Arguments)
	var res EmailSetResponse
	if err := json.Unmarshal(b, &res); err != nil {
		return EmailSetResponse{}, err
	}
	return res, nil
}

// DeleteMessage removes messages.
func (c *JmapClient) DeleteMessage(accountId string, destroy []string) (EmailSetResponse, error) {
	args := map[string]interface{}{
		"accountId": accountId,
		"destroy":   destroy,
	}
	inv := Invocation{
		Name:        "Email/set",
		Arguments:   args,
		MethodCallId: "c1",
	}
	env, err := c.SendRequest([]Invocation{inv}, []string{"urn:ietf:params:jmap:mail"})
	if err != nil {
		return EmailSetResponse{}, err
	}
	if len(env.MethodResponses) == 0 {
		return EmailSetResponse{}, fmt.Errorf("no methodResponses for Email/set")
	}
	b, _ := json.Marshal(env.MethodResponses[0].Arguments)
	var res EmailSetResponse
	if err := json.Unmarshal(b, &res); err != nil {
		return EmailSetResponse{}, err
	}
	return res, nil
}

// ListIdentities retrieves identities for the given account.
func (c *JmapClient) ListIdentities(accountId string, ids []string, properties []string) (IdentityGetResponse, error) {
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
		Name:        "Identity/get",
		Arguments:   args,
		MethodCallId: "c1",
	}
	env, err := c.SendRequest([]Invocation{inv}, []string{"urn:ietf:params:jmap:mail"})
	if err != nil {
		return IdentityGetResponse{}, err
	}
	if len(env.MethodResponses) == 0 {
		return IdentityGetResponse{}, fmt.Errorf("no methodResponses for Identity/get")
	}
	b, _ := json.Marshal(env.MethodResponses[0].Arguments)
	var res IdentityGetResponse
	if err := json.Unmarshal(b, &res); err != nil {
		return IdentityGetResponse{}, err
	}
	return res, nil
}
