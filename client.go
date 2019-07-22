package gomercury

import (
	"fmt"
	"net/http"

	"github.com/pquerna/ffjson/ffjson"
)

// DefaultHTTPTimeout contains the default timeout for HTTP requests
const DefaultHTTPTimeout = 30

// New creates a new instance of Mercury client
func New(mercuryURL string, timeout int, useAuth bool) *Client {

	if timeout == 0 {
		timeout = DefaultHTTPTimeout
	}

	return &Client{
		mercuryURL: mercuryURL,
		useAuth:    useAuth,
		timeout:    timeout,
	}
}

// NewMessage returns a new instance of EmailMessage
func NewMessage() EmailMessage {
	return EmailMessage{}
}

// Client defines the central structure to handle communication with Mercury
type Client struct {
	useAuth           bool
	authServiceURL    string
	authServiceKey    string
	authServiceSecret string
	mercuryURL        string
	timeout           int
	mercuryService    Service
}

// ConfigAuthService takes the auth service configuration parameters. The timeout used is the
// same as the one configured in the service
func (c *Client) ConfigAuthService(authServiceURL, authServiceKey, authServiceSecret string) {
	if c != nil {
		c.authServiceURL = authServiceURL
		c.authServiceKey = authServiceKey
		c.authServiceSecret = authServiceSecret
	}
}

func (c *Client) getMercuryService() Service {
	if c.mercuryService == nil {
		c.mercuryService = NewServiceCommunication(
			c.useAuth,
			c.authServiceURL,
			c.timeout,
			c.authServiceKey,
			c.authServiceSecret,
			c.mercuryURL, c.timeout)
	}

	return c.mercuryService
}

// SendTextEmail is a simple method to send an email in plain text format
func (c *Client) SendTextEmail(from, to, subject, text string, attachments []Attachment) (response SendMessageResponse, err error) {
	msg := NewMessage()

	m := Message{}
	m.FromEmail = from
	m.To = []ToAddress{ToAddress{Email: to}}
	m.Subject = subject
	m.Text = text

	if len(attachments) > 0 {
		m.Attachments = attachments
	}

	msg.Message = m
	return c.SendEmailMessage(msg)
}

// SendHTMLEmail is a simple method to send an email in HTML format
func (c *Client) SendHTMLEmail(from, to, subject, html string, attachments []Attachment) (response SendMessageResponse, err error) {
	msg := NewMessage()

	m := Message{}
	m.FromEmail = from
	m.To = []ToAddress{ToAddress{Email: to}}
	m.Subject = subject
	m.HTML = html

	if len(attachments) > 0 {
		m.Attachments = attachments
	}

	msg.Message = m
	return c.SendEmailMessage(msg)
}

// SendEmailMessage sends the email message with all the indicated configuration to Mercury
func (c *Client) SendEmailMessage(msg EmailMessage) (response SendMessageResponse, err error) {
	payload, errJSON := ffjson.Marshal(msg)
	if errJSON != nil {
		err = NewError(ErrCodeInternal, errJSON.Error())
		return
	}

	svc := c.getMercuryService()
	httpResponse, errRequest := svc.DoRequest("v1/smtp/email/send", http.MethodPost, "", string(payload))

	if errRequest != nil {
		// there was an error calling mercury service
		err = NewError(ErrCodeInternal, fmt.Sprintf("error calling mercury service: %s", errRequest.Error()))
		return
	}

	if httpResponse.StatusCode == http.StatusOK {
		defer httpResponse.Body.Close()
		jsonDecoder := ffjson.NewDecoder()
		if errJSON := jsonDecoder.DecodeReader(httpResponse.Body, &response); errJSON != nil {
			err = NewError(ErrCodeInternal, errJSON.Error())
			return
		}
	} else {
		// service returned other code different than 200
		err = NewError(ErrCodeInternal, fmt.Sprintf("Mercury returned status code %v", httpResponse.StatusCode))
	}

	return
}
