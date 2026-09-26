package jmap

// Mailbox represents a named set of Emails (JMAP's analogue of a mail folder / IMAP mailbox). Top-level, addressable JMAP data type.
type Mailbox struct {
	Id            string          `json:"id,omitempty"`            // server-assigned, omit when constructing a create payload
	Name          string          `json:"name"`                    // required
	ParentId      *string         `json:"parentId,omitempty"`      // Id|null
	Role          *string         `json:"role,omitempty"`          // String|null
	SortOrder     int             `json:"sortOrder,omitempty"`     // UnsignedInt (default=0)
	TotalEmails   int             `json:"totalEmails,omitempty"`   // server-assigned, omit when constructing a create payload
	UnreadEmails  int             `json:"unreadEmails,omitempty"`  // server-assigned, omit when constructing a create payload
	TotalThreads  int             `json:"totalThreads,omitempty"`  // server-assigned, omit when constructing a create payload
	UnreadThreads int             `json:"unreadThreads,omitempty"` // server-assigned, omit when constructing a create payload
	MyRights      *MailboxRights `json:"myRights,omitempty"`      // server-assigned, omit when constructing a create payload
	IsSubscribed  bool            `json:"isSubscribed,omitempty"`  // default=False
}

// MailboxRights represents the rights a user has on a mailbox.
type MailboxRights struct {
	MayReadItems   bool `json:"mayReadItems"`
	MayAddItems    bool `json:"mayAddItems"`
	MayRemoveItems bool `json:"mayRemoveItems"`
	MaySetSeen     bool `json:"maySetSeen"`
	MaySetKeywords bool `json:"maySetKeywords"`
	MayCreateChild bool `json:"mayCreateChild"`
	MayRename      bool `json:"mayRename"`
	MayDelete      bool `json:"mayDelete"`
	MaySubmit      bool `json:"maySubmit"`
}
