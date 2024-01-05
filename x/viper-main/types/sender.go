package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/vipernet-xyz/utils-go/client"
)

var (
	// Err4xxOnConnection error when RPC responds with 4xx code
	Err4xxOnConnection = errors.New("rpc responded with 4xx")
	// Err5xxOnConnection error when RPC responds with 5xx code
	Err5xxOnConnection = errors.New("rpc responded with 5xx")
	// ErrUnexpectedCodeOnConnection error when RPC responds with unexpected code
	ErrUnexpectedCodeOnConnection = errors.New("rpc responded with unexpected code")
	// ErrNoDispatchers error when dispatch call is requested with no dispatchers set
	ErrNoDispatchers = errors.New("no dispatchers")
	// ErrNonJSONResponse error when sender does not respond with a JSON
	ErrNonJSONResponse = errors.New("non JSON response")

	errOnRelayRequest = errors.New("error on relay request")
)

// Sender struct handler por JSON RPC sender
type Sender struct {
	rpcURL string
	client *client.Client
}

// NewSender returns Sender instance from input
func NewSender(rpcURL string) *Sender {
	return &Sender{
		rpcURL: rpcURL,
		client: client.NewDefaultClient(),
	}
}

// UpdateRequestConfig updates retries and timeout used for RPC requests
func (p *Sender) UpdateRequestConfig(retries int, timeout time.Duration) {
	p.client = client.NewCustomClient(retries, timeout)
}

// ResetRequestConfigToDefault resets request config to default
func (p *Sender) ResetRequestConfigToDefault() {
	p.client = client.NewDefaultClient()
}

func (p *Sender) getFinalRPCURL(rpcURL string, route V1RPCRoute) (string, error) {
	if rpcURL != "" {
		return rpcURL, nil
	}
	return p.rpcURL, nil
}

func (p *Sender) doPostRequest(rpcURL string, params any, route V1RPCRoute) (*http.Response, error) {
	finalRPCURL, err := p.getFinalRPCURL(rpcURL, route)
	if err != nil {
		return nil, err
	}

	output, err := p.client.PostWithURLJSONParams(fmt.Sprintf("%s%s", finalRPCURL, route), params, http.Header{})
	if err != nil {
		return nil, err
	}

	if output.StatusCode == http.StatusBadRequest {
		return output, returnRPCError(route, output.Body)
	}

	if string(output.Status[0]) == "4" {
		return output, Err4xxOnConnection
	}

	if string(output.Status[0]) == "5" {
		return output, Err5xxOnConnection
	}

	if string(output.Status[0]) == "2" {
		return output, nil
	}

	return nil, ErrUnexpectedCodeOnConnection
}

func returnRPCError(route V1RPCRoute, body io.ReadCloser) error {
	if route == ClientRelayRoute {
		return errOnRelayRequest
	}

	defer CloseOrLog(body)

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	output := RPCError{}

	err = json.Unmarshal(bodyBytes, &output)
	if err != nil {
		return err
	}

	return &output
}

// Relay does request to be relayed to a target blockchain
func (p *Sender) Relay(rpcURL string, input *RelayInput) (*RelayOutput, error) {
	rawOutput, reqErr := p.doPostRequest(rpcURL, input, ClientRelayRoute)

	defer closeOrLog(rawOutput)

	if reqErr != nil && !errors.Is(reqErr, errOnRelayRequest) {
		return nil, reqErr
	}

	bodyBytes, err := ioutil.ReadAll(rawOutput.Body)
	if err != nil {
		return nil, err
	}

	if errors.Is(reqErr, errOnRelayRequest) {
		return nil, parseRelayErrorOutput(bodyBytes, input.Proof.ServicerPubKey)
	}

	return parseRelaySuccesfulOutput(bodyBytes)
}

// Relay does request to be relayed to a target blockchain
func (p *Sender) localRelay(rpcURL string, input *RelayInput) (*RelayOutput, error) {
	rawOutput, reqErr := p.doPostRequest(rpcURL, input, LocalRelayRoute)

	defer closeOrLog(rawOutput)

	if reqErr != nil && !errors.Is(reqErr, errOnRelayRequest) {
		return nil, reqErr
	}

	bodyBytes, err := ioutil.ReadAll(rawOutput.Body)
	if err != nil {
		return nil, err
	}

	if errors.Is(reqErr, errOnRelayRequest) {
		return nil, parseRelayErrorOutput(bodyBytes, input.Proof.ServicerPubKey)
	}

	return parseRelaySuccesfulOutput(bodyBytes)
}

func parseRelaySuccesfulOutput(bodyBytes []byte) (*RelayOutput, error) {
	output := RelayOutput{}

	err := json.Unmarshal(bodyBytes, &output)
	if err != nil {
		return nil, err
	}

	if !json.Valid([]byte(output.Response)) {
		return nil, ErrNonJSONResponse
	}

	return &output, nil
}

func parseRelayErrorOutput(bodyBytes []byte, servicerPubKey string) error {
	output := RelayErrorOutput{}

	err := json.Unmarshal(bodyBytes, &output)
	if err != nil {
		return err
	}

	return &RelayError{
		Code:           output.Error.Code,
		Codespace:      output.Error.Codespace,
		Message:        output.Error.Message,
		ServicerPubKey: servicerPubKey,
	}
}

func closeOrLog(response *http.Response) {
	if response != nil {
		io.Copy(ioutil.Discard, response.Body)
		CloseOrLog(response.Body)
	}
}

// CloseOrLog closes closable body and logs if close fails
func CloseOrLog(closable io.Closer) {
	if closable != nil {
		err := closable.Close()
		if err != nil {
			fmt.Println("closing object failed") // TODO: make this log better
		}
	}
}

// RelayError represents the thrown error of a relay request
type RelayError struct {
	Code           RelayErrorCode
	Codespace      string
	Message        string
	ServicerPubKey string
}

type RelayErrorCode int

func (e *RelayError) Error() string {
	return fmt.Sprintf("Request failed with code: %v, codespace: %s and message: %s\nWith ServicerPubKey: %s",
		e.Code, e.Codespace, e.Message, e.ServicerPubKey)
}

// RelayErrorOutput represents error response of relay request
type RelayErrorOutput struct {
	Error struct {
		Code      RelayErrorCode `json:"code"`
		Codespace string         `json:"codespace"`
		Message   string         `json:"message"`
	} `json:"error"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error returns string representation of error
// needed to implement error interface
func (e *RPCError) Error() string {
	return fmt.Sprintf("Request failed with code: %v and message: %s", e.Code, e.Message)
}
