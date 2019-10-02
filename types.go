package gomercury

// EmailMessage represents email message sending requests
type EmailMessage struct {
	Message  Message     `json:"message"`
	Options  SendOptions `json:"send_options,omitempty"`
	Provider Provider    `json:"provider,omitempty"`
}

// Message holds the message content
type Message struct {
	HTML             string            `json:"html,omitempty"`
	Text             string            `json:"text,omitempty"`
	TemplateName     string            `json:"template_name,omitempty"`
	TemplateData     map[string]string `json:"template_data,omitempty"`
	TemplateLanguage string            `json:"template_language,omitempty"`
	Subject          string            `json:"subject,omitempty"`
	FromEmail        string            `json:"from_email,omitempty"`
	FromName         string            `json:"from_name,omitempty"`
	To               []ToAddress       `json:"to,omitempty"`
	Attachments      []Attachment      `json:"attachments,omitempty"`
}

// ToAddress represent an recipient to send an email message
type ToAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"` // values: 'to', 'cc', 'bcc' (default: 'to')
}

// Attachment represents an email attachment file
type Attachment struct {
	Content string `json:"content,omitempty"`
	Type    string `json:"type,omitempty"`
	Name    string `json:"name,omitempty"`
}

// SendOptions has the email sending options
type SendOptions *struct {
	Method   string `json:"method,omitempty"` // values: 'sync', 'async', 'schedule' (default: 'sync')
	Schedule struct {
		At int64 `json:"at,omitempty"`
	}
}

// Provider has the provider ID and credentials
type Provider struct {
	Name        string                 `json:"name,omitempty"`
	Credentials map[string]interface{} `json:"credentials,omitempty"`
}

// SendMessageResponse holds the response sent to the client for a message send sync operation
type SendMessageResponse struct {
	OperationID string         `json:"operation_id"`
	Result      []SendResponse `json:"result"`
}

// SendResponse returns the result of an email send operation
type SendResponse struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	Status         string `json:"status"`
	RejectedReason string `json:"reject_reason,omitempty"`
}
