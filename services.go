package gomercury

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/pquerna/ffjson/ffjson"
)

// MaxRetries contains the max numer of retries if the request fails due to invalid access token
const MaxRetries = 1

// Service interface holds the
type Service interface {
	DoRequest(path string, mehtod string, contentType string, payload string) (*http.Response, error)
}

// NewServiceCommunication returns a new configured Service Communication instance
func NewServiceCommunication(authRequired bool, authURL string, authTimeout int, key, secret string, serviceURL string, serviceTimeout int) (sc *ServiceComm) {
	sc = new(ServiceComm)
	sc.authRequired = authRequired
	sc.authServiceURL = authURL
	sc.authServiceTimeout = time.Duration(authTimeout) * time.Second
	sc.key = key
	sc.secret = secret
	sc.serviceURL = serviceURL
	sc.serviceTimeout = time.Duration(serviceTimeout) * time.Second
	return
}

// ServiceComm implements interface Service to handle communication with other
// APC Core services. It's also capable of requesting access token and retring
// requests in case the token is expired.
type ServiceComm struct {
	authRequired       bool
	serviceURL         string
	serviceTimeout     time.Duration
	authServiceURL     string
	authServiceTimeout time.Duration
	key                string
	secret             string
	storedToken        string
}

func (sc *ServiceComm) String() string {
	return "Service Communication"
}

// DoRequest invokes the service endppint with the configured parameters
func (sc *ServiceComm) DoRequest(path string, mehtod string, contentType string, payload string) (resp *http.Response, err error) {

	retryCount := 0

	for {
		url := fmt.Sprintf("%s/%s", sc.serviceURL, path)
		httpRequest := gorequest.New()

		// timeout
		httpRequest = httpRequest.Timeout(sc.authServiceTimeout)

		// http verb
		switch mehtod {
		case http.MethodPost:
			httpRequest = httpRequest.Post(url)
		case http.MethodGet:
			httpRequest = httpRequest.Get(url)
		case http.MethodPut:
			httpRequest = httpRequest.Put(url)
		case http.MethodDelete:
			httpRequest = httpRequest.Delete(url)
		case http.MethodPatch:
			httpRequest = httpRequest.Patch(url)
		default:
			err = fmt.Errorf("http method %s not supported by '%s'", mehtod, sc.String())
			return
		}

		// get access token
		if sc.authRequired {
			accessToken, errAccessToken := sc.getAccessToken()
			if errAccessToken != nil {
				err = errAccessToken
				return
			}

			httpRequest.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", accessToken)}
		}

		// set headers
		httpRequest.Header["User-Agent"] = []string{"gomercury"}

		if contentType != "" {
			httpRequest.Header["Content-Type"] = []string{contentType}
		} else {
			httpRequest.Header["Content-Type"] = []string{"application/json"}
		}

		// payload
		if len(payload) > 0 {
			httpRequest = httpRequest.Send(payload)
		}

		// make request
		if httpResp, _, errRequest := httpRequest.End(); errRequest == nil {
			resp = httpResp
			if httpResp.StatusCode != http.StatusUnauthorized {
				break
			}

			if !sc.authRequired {
				break
			}

			sc.storedToken = ""
			retryCount++
		} else {
			err = errRequest[0]
			break
		}

		// in case of 401, token might be expired, so we retry
		if retryCount == MaxRetries {
			break
		}
	}

	return
}

func (sc *ServiceComm) getAccessToken() (token string, err error) {

	// if we have a stored token, use that one
	if sc.storedToken != "" {
		token = sc.storedToken
		return
	}

	if sc.authServiceURL == "" {
		return "", errors.New("auth-service URL not configured")
	}

	url := fmt.Sprintf("%s/access_token", sc.authServiceURL)
	httpRequest := gorequest.New().Post(url)
	httpRequest.Header["Content-Type"] = []string{"application/json"}
	httpRequest.Header["User-Agent"] = []string{"purchase-service"}

	payload := fmt.Sprintf(`{ "key":"%s", "secret":"%s" }`, sc.key, sc.secret)

	if response, _, errRequest := httpRequest.Send(payload).Timeout(sc.authServiceTimeout).End(); errRequest == nil {

		if response != nil {

			decoder := ffjson.NewDecoder()
			if response.StatusCode != http.StatusOK {
				err = fmt.Errorf("error trying to get new access token: %s", response.Status)
			} else {

				// retrieve token
				type tokenResponse struct {
					Token      string `json:"token"`
					ExpiresInt int    `json:"expires_in"`
				}

				tr := new(tokenResponse)
				if errJSON := decoder.DecodeReader(response.Body, tr); errJSON == nil {
					token = tr.Token
					sc.storedToken = token
				}
			}
		} else {
			err = ErrorUnknown
		}

	} else {
		err = ErrorUnknown
	}

	return
}
