package gomercury

// Generic error definitions
var (
	ErrorUnknown             = NewError(ErrcCodeUnknown, "unknown error")
	ErrorNoAuthConfiguration = NewError(ErrCodeNoAuthService, "auth service not configured properly")
)

// Generic error codes
const (
	ErrcCodeUnknown      = "unknown"
	ErrCodeInternal      = "internal_error"
	ErrCodeNoAuthService = "no_auth"
)

// Error represents a Mercury client error
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() (err string) {
	if e != nil {
		if e.Code != "" {
			err = e.Code
		}
		if e.Message != "" {
			if err != "" {
				err += ": " + e.Message
			} else {
				err = e.Message
			}
		}
		if err == "" {
			err = "mercury client error"
		}
	}

	return
}

// NewError returns a new error instance
func NewError(code, message string) (err *Error) {
	return &Error{
		Code:    code,
		Message: message,
	}
}
