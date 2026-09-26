package jmap

// Session represents the JMAP Session resource.
// It is fetched once from the well‑known URL and cached by the client.
type Session struct {
	Capabilities    map[string]any   `json:"capabilities"`    // capability URN -> capability object
	Accounts        map[string]Account `json:"accounts"`        // account ID -> account object
	PrimaryAccounts map[string]string `json:"primaryAccounts"` // capability URN -> default account ID
	Username        string            `json:"username"`
	ApiUrl          string            `json:"apiUrl"`
	DownloadUrl     string            `json:"downloadUrl"`
	UploadUrl       string            `json:"uploadUrl"`
	EventSourceUrl  string            `json:"eventSourceUrl"`
	State           string            `json:"state"`

	// internal field, not part of the wire format, holds the URL that was used to fetch the session.
	origin string `json:"-"`
}

// Account represents an account entry in the Session object.
type Account struct {
	Name                string         `json:"name,omitempty"`
	IsPersonal          bool           `json:"isPersonal,omitempty"`
	IsReadOnly          bool           `json:"isReadOnly,omitempty"`
	AccountCapabilities map[string]any `json:"accountCapabilities,omitempty"` // capability URN -> capability object
}

// CoreCapability is the value of capabilities["urn:ietf:params:jmap:core"].
type CoreCapability struct {
	MaxSizeUpload          int      `json:"maxSizeUpload"`
	MaxConcurrentUpload    int      `json:"maxConcurrentUpload"`
	MaxSizeRequest         int      `json:"maxSizeRequest"`
	MaxConcurrentRequests  int      `json:"maxConcurrentRequests"`
	MaxCallsInRequest      int      `json:"maxCallsInRequest"`
	MaxObjectsInGet        int      `json:"maxObjectsInGet"`
	MaxObjectsInSet        int      `json:"maxObjectsInSet"`
	CollationAlgorithms    []string `json:"collationAlgorithms"`
}
