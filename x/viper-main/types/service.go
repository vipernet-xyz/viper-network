package types

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"
)

const DEFAULTHTTPMETHOD = "POST"

// "Relay" - A read / write API request from a hosted (non native) external blockchain
type Relay struct {
	Payload Payload    `json:"payload"` // the data payload of the request
	Meta    RelayMeta  `json:"meta"`    // metadata for the relay request
	Proof   RelayProof `json:"proof"`   // the authentication scheme needed for work
}

// "Validate" - Checks the validity of a relay request using store data
func (r *Relay) Validate(ctx sdk.Ctx, posKeeper PosKeeper, requestorsKeeper RequestorsKeeper, viperKeeper ViperKeeper, hb *HostedBlockchains, sessionBlockHeight int64, node *ViperNode) (maxPossibleRelays sdk.BigInt, err sdk.Error) {
	// validate payload
	if err := r.Payload.Validate(); err != nil {
		return sdk.ZeroInt(), NewEmptyPayloadDataError(ModuleName)
	}
	// validate the metadata
	if err := r.Meta.Validate(ctx); err != nil {
		return sdk.ZeroInt(), err
	}
	// validate the relay merkleHash = request merkleHash
	if r.Proof.RequestHash != r.RequestHashString() {
		return sdk.ZeroInt(), NewRequestHashError(ModuleName)
	}
	// ensure the blockchain is supported locally
	if !hb.Contains(r.Proof.Blockchain) {
		return sdk.ZeroInt(), NewUnsupportedBlockchainNodeError(ModuleName)
	}
	// ensure session block height == one in the relay proof
	if r.Proof.SessionBlockHeight != sessionBlockHeight {
		return sdk.ZeroInt(), NewInvalidBlockHeightError(ModuleName)
	}
	// get the session context
	sessionCtx, er := ctx.PrevCtx(sessionBlockHeight)
	if er != nil {
		return sdk.ZeroInt(), sdk.ErrInternal(er.Error())
	}
	// get the application that staked on behalf of the client
	app, found := GetRequestorFromPublicKey(sessionCtx, requestorsKeeper, r.Proof.Token.RequestorPublicKey)
	if !found {
		return sdk.ZeroInt(), NewRequestorNotFoundError(ModuleName)
	}
	// get session node count from that session height
	sessionNodeCount := int64(r.Proof.NumServicers)
	// get max possible relays
	maxPossibleRelays = MaxPossibleRelays(app, sessionNodeCount)
	// generate the session header
	header := SessionHeader{
		RequestorPubKey:    r.Proof.Token.RequestorPublicKey,
		Chain:              r.Proof.Blockchain,
		GeoZone:            r.Proof.GeoZone,
		NumServicers:       r.Proof.NumServicers,
		SessionBlockHeight: r.Proof.SessionBlockHeight,
	}
	// validate unique relay
	evidence, totalRelays := GetTotalProofs(header, RelayEvidence, maxPossibleRelays, node.EvidenceStore)
	if node.EvidenceStore.IsSealed(evidence) {
		return sdk.ZeroInt(), NewSealedEvidenceError(ModuleName)
	}
	// get evidence key by proof
	if !IsUniqueProof(r.Proof, evidence) {
		return sdk.ZeroInt(), NewDuplicateProofError(ModuleName)
	}
	// validate not over service
	if sdk.NewInt(totalRelays).GTE(maxPossibleRelays) {
		return sdk.ZeroInt(), NewOverServiceError(ModuleName)
	}
	// validate the Proof
	nodeAddr := node.GetAddress()
	if err := r.Proof.ValidateLocal(app.GetChains(), int(sessionNodeCount), sessionBlockHeight, nodeAddr); err != nil {
		return sdk.ZeroInt(), err
	}
	// check cache
	session, found := GetSession(header, node.SessionStore)
	// if not found generate the session
	if !found {
		bh, err := sessionCtx.BlockHash(viperKeeper.Codec(), sessionCtx.BlockHeight())
		if err != nil {
			return sdk.ZeroInt(), sdk.ErrInternal(err.Error())
		}
		var er sdk.Error
		session, er = NewSession(sessionCtx, ctx, posKeeper, header, hex.EncodeToString(bh))
		if er != nil {
			return sdk.ZeroInt(), er
		}
		// add to cache
		SetSession(session, node.SessionStore)
	}
	// validate the session
	err = session.Validate(nodeAddr, app, int(sessionNodeCount))
	if err != nil {
		return sdk.ZeroInt(), err
	}
	// if the payload method is empty, set it to the default
	if r.Payload.Method == "" {
		r.Payload.Method = DEFAULTHTTPMETHOD
	}
	return maxPossibleRelays, nil
}

// "Execute" - Attempts to do a request on the non-native blockchain specified
func (r Relay) Execute(hostedBlockchains *HostedBlockchains, address *sdk.Address) (string, sdk.Error) {
	// retrieve the hosted blockchain url requested
	chain, err := hostedBlockchains.GetChain(r.Proof.Blockchain)
	if err != nil {
		// metric track
		addServiceMetricErrorFor(r.Proof.Blockchain, address)
		return "", err
	}
	url := strings.Trim(chain.URL, `/`)
	if len(r.Payload.Path) > 0 {
		url = url + "/" + strings.Trim(r.Payload.Path, `/`)
	}
	// do basic http request on the relay
	res, er := executeHTTPRequest(r.Payload.Data, url, GlobalViperConfig.UserAgent, chain.BasicAuth, r.Payload.Method, r.Payload.Headers)
	if er != nil {
		// metric track
		addServiceMetricErrorFor(r.Proof.Blockchain, address)
		return res, NewHTTPExecutionError(ModuleName, er)
	}
	return res, nil
}

// "Bytes" - Returns the bytes representation of the Relay
func (r Relay) Bytes() []byte {
	//Anonymous Struct used because of #742 empty proof object being marshalled
	relay := struct {
		Payload Payload   `json:"payload"` // the data payload of the request
		Meta    RelayMeta `json:"meta"`    // metadata for the relay request
	}{r.Payload, r.Meta}
	res, err := json.Marshal(relay)
	if err != nil {
		log.Fatal(fmt.Errorf("cannot marshal relay request hash: %s", err.Error()))
	}
	return res
}

// "Requesthash" - The cryptographic merkleHash representation of the request
func (r Relay) RequestHash() []byte {
	return Hash(r.Bytes())
}

// "RequestHashString" - The hex string representation of the request merkleHash
func (r Relay) RequestHashString() string {
	return hex.EncodeToString(r.RequestHash())
}

// "Payload" - A data being sent to the non-native chain
type Payload struct {
	Data    string            `json:"data"`              // the actual data string for the external chain
	Method  string            `json:"method"`            // the http CRUD method
	Path    string            `json:"path"`              // the REST Path
	Headers map[string]string `json:"headers,omitempty"` // http headers
}

// "Bytes" - The bytes reprentation of a payload object
func (p Payload) Bytes() []byte {
	bz, err := json.Marshal(p)
	if err != nil {
		log.Fatalf(fmt.Errorf("an error occured converting the payload to bytes:\n%v", err).Error())
	}
	return bz
}

// "Hash" - The cryptographic merkleHash representation of the payload object
func (p Payload) Hash() []byte {
	return Hash(p.Bytes())
}

// "HashString" - The hex encoded string representation of the payload object
func (p Payload) HashString() string {
	return hex.EncodeToString(p.Hash())
}

// "Validate" - Validity check for the payload object
func (p Payload) Validate() sdk.Error {
	if p.Data == "" && p.Path == "" {
		return NewEmptyPayloadDataError(ModuleName)
	}
	return nil
}

// "payload" - A structure used for custom json marshalling/unmarshalling
type payload struct {
	Data    string            `json:"data"`
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
}

// "MarshalJSON" - Overrides json marshalling
func (p Payload) MarshalJSON() ([]byte, error) {
	pay := payload{ //nolint:golint,gosimple
		Data:    p.Data,
		Method:  p.Method,
		Path:    p.Path,
		Headers: p.Headers,
	}
	return json.Marshal(pay)
}

// "RelayMeta" - Metadata that is included in the relay request
type RelayMeta struct {
	BlockHeight int64 `json:"block_height"` // the block height when the request is made
}

// "Validate" - Validates the relay meta object
func (m RelayMeta) Validate(ctx sdk.Ctx) sdk.Error {
	// ensures the block height is within the acceptable range
	if ctx.BlockHeight()+int64(GlobalViperConfig.ClientBlockSyncAllowance) < m.BlockHeight || ctx.BlockHeight()-int64(GlobalViperConfig.ClientBlockSyncAllowance) > m.BlockHeight {
		return NewOutOfSyncRequestError(ModuleName)
	}
	return nil
}

func InitClientBlockAllowance(allowance int) {
	GlobalViperConfig.ClientBlockSyncAllowance = allowance
}

// "Validate" - The servicer validates the response after signing
func (rr RelayResponse) Validate() sdk.Error {
	// cannot contain empty response
	if rr.Response == "" {
		return NewEmptyResponseError(ModuleName)
	}
	// cannot contain empty signature (servicers must be accountable)
	if rr.Signature == "" || len(rr.Signature) == crypto.Ed25519SignatureSize {
		return NewResponseSignatureError(ModuleName)
	}
	return nil
}

// "Hash" - The cryptographic merkleHash representation of the relay response
func (rr RelayResponse) Hash() []byte {
	seed, err := json.Marshal(relayResponse{
		Signature: "",
		Response:  rr.Response,
		Proof:     rr.Proof.HashString(),
	})
	if err != nil {
		log.Fatalf(fmt.Errorf("an error occured hashing the relay response:\n%v", err).Error())
	}
	return Hash(seed)
}

// "HashString" - The hex string representation of the merkleHash
func (rr RelayResponse) HashString() string {
	return hex.EncodeToString(rr.Hash())
}

// "relayResponse" - a structure used for custom json
type relayResponse struct {
	Signature string `json:"signature"`
	Response  string `json:"payload"`
	Proof     string `json:"Proof"`
}

// "ChallengeReponse" - The response object used in challenges
type ChallengeResponse struct {
	Response string `json:"response"`
}

// "DispatchResponse" - The response object used in dispatching
type DispatchResponse struct {
	Session     DispatchSession `json:"session"`
	BlockHeight int64           `json:"block_height"`
}

type DispatchSession struct {
	SessionHeader    `json:"header"`
	SessionKey       `json:"key"`
	SessionServicers []exported.ValidatorI `json:"servicers"`
	SessionFishermen []exported.ValidatorI `json:"fishermen"`
}

// "executeHTTPRequest" takes in the raw JSON string and forwards it to the RPC endpoint
func executeHTTPRequest(payload, url, userAgent string, basicAuth BasicAuth, method string, headers map[string]string) (string, error) {
	// Check if the payload is compressed
	isCompressed := isPayloadCompressed(payload)

	// Decompress the payload if it is compressed
	if isCompressed {
		decodedPayload, err := decompressPayload(payload)
		if err != nil {
			return "", err
		}
		payload = decodedPayload
	}

	// Generate an HTTP request
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", err
	}

	if basicAuth.Username != "" {
		req.SetBasicAuth(basicAuth.Username, basicAuth.Password)
	}

	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	// Add headers if needed
	if len(headers) == 0 {
		req.Header.Set("Content-Type", "requestor/json")
	} else {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	// Set the "Accept-Encoding" header to indicate compressed response is expected
	req.Header.Set("Accept-Encoding", "gzip")

	// Execute the request
	resp, err := (&http.Client{Timeout: globalRPCTimeout * time.Millisecond}).Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// Check if the response is compressed
	isResponseCompressed := strings.Contains(resp.Header.Get("Content-Encoding"), "gzip")

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Compress the response body if the payload was not already compressed
	if isCompressed && !isResponseCompressed {
		body, err = compressResponse(string(body))
		if err != nil {
			return "", err
		}
	}

	if GlobalViperConfig.JSONSortRelayResponses {
		body = []byte(sortJSONResponse(string(body)))
	}

	// Return the response
	return string(body), nil
}

// Check if the payload is compressed
func isPayloadCompressed(payload string) bool {
	// Check if the payload starts with the gzip magic number
	return strings.HasPrefix(payload, "\x1f\x8b")
}

// Decompresses a gzip-encoded payload
func decompressPayload(payload string) (string, error) {
	reader, err := gzip.NewReader(strings.NewReader(payload))
	if err != nil {
		return "", err
	}

	defer reader.Close()

	decoded, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

// Compresses a response using gzip
func compressResponse(response string) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	_, err := writer.Write([]byte(response))
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// "sortJSONResponse" - sorts json from a relay response
func sortJSONResponse(response string) string {
	var rawJSON map[string]interface{}
	// unmarshal into json
	if err := json.Unmarshal([]byte(response), &rawJSON); err != nil {
		return response
	}
	// marshal into json
	bz, err := json.Marshal(rawJSON)
	if err != nil {
		return response
	}
	return string(bz)
}

func ErrorWarrantsDispatch(err error) bool {
	cErr := err.(sdk.Error)
	if cErr.Code() == NewOverServiceError(ModuleName).Code() ||
		cErr.Code() == NewInvalidBlockHeightError(ModuleName).Code() ||
		cErr.Code() == NewInvalidSessionError(ModuleName).Code() ||
		cErr.Code() == NewOutOfSyncRequestError(ModuleName).Code() {
		return true
	}
	return false
}

func addServiceMetricErrorFor(blockchain string, address *sdk.Address) {
	if GlobalViperConfig.LeanViper {
		go GlobalServiceMetric().AddErrorFor(blockchain, address)
	} else {
		GlobalServiceMetric().AddErrorFor(blockchain, address)
	}
}