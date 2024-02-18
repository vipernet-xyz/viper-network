package rpc

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"

	"github.com/vipernet-xyz/viper-network/app"
	"github.com/vipernet-xyz/viper-network/x/viper-main/types"

	"github.com/julienschmidt/httprouter"
)

// Dispatch supports CORS functionality
func Dispatch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if cors(&w, r) {
		return
	}
	d := types.SessionHeader{}
	if err := PopModel(w, r, ps, &d); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	res, err := app.VCA.HandleDispatch(d)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, er := json.Marshal(res)
	if er != nil {
		WriteErrorResponse(w, 400, er.Error())
		return
	}
	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

type RPCRelayResponse struct {
	Signature string `json:"signature"`
	Response  string `json:"response"`
	// remove proof object because client already knows about it
}

type RPCRelayErrorResponse struct {
	Error    error                   `json:"error"`
	Dispatch *types.DispatchResponse `json:"dispatch"`
}

// Relay supports CORS functionality
func Relay(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var relay = types.Relay{}
	if cors(&w, r) {
		return
	}
	if err := PopModel(w, r, ps, &relay); err != nil {
		response := RPCRelayErrorResponse{
			Error: err,
		}
		j, _ := json.Marshal(response)
		WriteJSONResponseWithCode(w, string(j), r.URL.Path, r.Host, 400)
		return
	}
	res, dispatch, err := app.VCA.HandleRelay(relay)
	if err != nil {
		response := RPCRelayErrorResponse{
			Error:    err,
			Dispatch: dispatch,
		}
		j, _ := json.Marshal(response)
		WriteJSONResponseWithCode(w, string(j), r.URL.Path, r.Host, 400)
		return
	}
	response := RPCRelayResponse{
		Signature: res.Signature,
		Response:  res.Response,
	}
	j, er := json.Marshal(response)
	if er != nil {
		WriteErrorResponse(w, 400, er.Error())
		return
	}
	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

// UpdateChains
func UpdateChains(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	value := r.URL.Query().Get("authtoken")
	if value == app.AuthToken.Value {
		var hostedChainsSlice []types.HostedBlockchain
		if err := PopModel(w, r, ps, &hostedChainsSlice); err != nil {
			WriteErrorResponse(w, 400, err.Error())
			return
		}
		m := make(map[string]types.HostedBlockchain)
		for _, chain := range hostedChainsSlice {
			if err := servicersTypes.ValidateNetworkIdentifier(chain.ID); err != nil {
				WriteErrorResponse(w, 400, fmt.Sprintf("invalid ID: %s in network identifier in json", chain.ID))
				return
			}
			m[chain.ID] = chain
		}
		result, err := app.VCA.SetHostedChains(m)
		if err != nil {
			WriteErrorResponse(w, 400, err.Error())
		} else {
			j, er := json.Marshal(result)
			if er != nil {
				WriteErrorResponse(w, 400, er.Error())
				return
			}
			WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
		}
	} else {
		WriteErrorResponse(w, 401, "wrong authtoken "+value)
	}
}

// Update Geozones
func UpdateGeoZones(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	value := r.URL.Query().Get("authtoken")
	if value == app.AuthToken.Value {
		var hostedGeoZoneSlice []types.GeoZone
		if err := PopModel(w, r, ps, &hostedGeoZoneSlice); err != nil {
			WriteErrorResponse(w, 400, err.Error())
			return
		}
		m := make(map[string]types.GeoZone)
		for _, zone := range hostedGeoZoneSlice {
			if err := servicersTypes.ValidateGeoZone(zone.ID); err != nil {
				WriteErrorResponse(w, 400, fmt.Sprintf("invalid ID: %s in geo zone identifier in json", zone.ID))
				return
			}
			m[zone.ID] = zone
		}
		result, err := app.VCA.SetHostedGeoZone(m)
		if err != nil {
			WriteErrorResponse(w, 400, err.Error())
		} else {
			j, er := json.Marshal(result)
			if er != nil {
				WriteErrorResponse(w, 400, er.Error())
				return
			}
			WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
		}
	} else {
		WriteErrorResponse(w, 401, "wrong authtoken "+value)
	}
}

// Stop
func Stop(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	value := r.URL.Query().Get("authtoken")
	if value == app.AuthToken.Value {
		app.ShutdownViperCore()
		err := app.VCA.TMNode().Stop()
		if err != nil {
			fmt.Println(err)
			WriteErrorResponse(w, 400, err.Error())
			fmt.Println("Force Stop , PID:" + fmt.Sprint(os.Getpid()))
			os.Exit(1)
		}
		fmt.Println("Stop Successful, PID:" + fmt.Sprint(os.Getpid()))
		os.Exit(0)
	} else {
		WriteErrorResponse(w, 401, "wrong authtoken "+value)
	}
}

// Challenge supports CORS functionality
func Challenge(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var challenge = types.ChallengeProofInvalidData{}
	if cors(&w, r) {
		return
	}
	if err := PopModel(w, r, ps, &challenge); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	res, err := app.VCA.HandleChallenge(challenge)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, er := json.Marshal(res)
	if er != nil {
		WriteErrorResponse(w, 400, er.Error())
		return
	}
	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

type SendRawTxParams struct {
	Addr        string `json:"address"`
	RawHexBytes string `json:"raw_hex_bytes"`
}

func SendRawTx(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = SendRawTxParams{}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	bz, err := hex.DecodeString(params.RawHexBytes)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	res, err := app.VCA.SendRawTx(params.Addr, bz)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, er := app.Codec().MarshalJSON(res)
	if er != nil {
		WriteErrorResponse(w, 400, er.Error())
		return
	}
	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

type simRelayParams struct {
	RelayNetworkID string        `json:"relay_network_id"` // RelayNetworkID
	Payload        types.Payload `json:"payload"`          // the data payload of the request
}

func SimRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = simRelayParams{}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}

	hostedChains := app.NewHostedChains(false)
	chain, err := hostedChains.GetChain(params.RelayNetworkID)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}

	if chain.WebSocketURL != "" {
		// If WebSocket URL is provided, execute WebSocket request
		res, er := executeWebSocketRequest(params.Payload.Data, chain.WebSocketURL)
		if er != nil {
			WriteErrorResponse(w, 400, er.Error())
			return
		}

		WriteResponse(w, res, r.URL.Path, r.Host)
	} else {
		// If no WebSocket URL, perform basic HTTP request
		url := strings.Trim(chain.HTTPURL, `/`)
		if len(params.Payload.Path) > 0 {
			url = url + "/" + strings.Trim(params.Payload.Path, `/`)
		}

		// Do basic HTTP request on the relay
		res, er := executeHTTPRequest(params.Payload.Data, url, types.GlobalViperConfig.UserAgent, chain.BasicAuth, params.Payload.Method, params.Payload.Headers)
		if er != nil {
			WriteErrorResponse(w, 400, er.Error())
			return
		}

		WriteResponse(w, string(res), r.URL.Path, r.Host)
	}
}

func executeHTTPRequest(payload, url, userAgent string, basicAuth types.BasicAuth, method string, headers map[string]string) (string, error) {
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
	resp, err := (&http.Client{Timeout: types.GetRPCTimeout() * time.Millisecond}).Do(req)
	if err != nil {
		return payload, err
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

	// Return the response
	return string(body), nil
}

func executeWebSocketRequest(payload, url string) (string, error) {
	// Use the gorilla websocket Dialer
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return "", fmt.Errorf("Error connecting to WebSocket: %s", err)
	}
	defer conn.Close()

	// Write the data to the WebSocket connection
	err = conn.WriteMessage(websocket.TextMessage, []byte(payload))
	if err != nil {
		return "", fmt.Errorf("Error writing to WebSocket: %s", err)
	}

	// Read and handle WebSocket responses
	_, response, err := conn.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("Error reading from WebSocket: %s", err)
	}

	// Return the WebSocket response
	return string(response), nil
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

func FishermanTrigger(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var trigger = types.FishermenTrigger{}
	if cors(&w, r) {
		return
	}
	if err := PopModel(w, r, ps, &trigger); err != nil {
		response := RPCRelayErrorResponse{
			Error: err,
		}
		j, _ := json.Marshal(response)
		WriteJSONResponseWithCode(w, string(j), r.URL.Path, r.Host, 400)
		return
	}
	res, dispatch, err := app.VCA.HandleFishermanTrigger(trigger)
	if err != nil {
		response := RPCRelayErrorResponse{
			Error:    err,
			Dispatch: dispatch,
		}
		j, _ := json.Marshal(response)
		WriteJSONResponseWithCode(w, string(j), r.URL.Path, r.Host, 400)
		return
	}
	response := RPCRelayResponse{
		Signature: res.Signature,
		Response:  res.Response,
	}
	j, er := json.Marshal(response)
	if er != nil {
		WriteErrorResponse(w, 400, er.Error())
		return
	}
	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

type RelayWebsocketResponse struct {
	Signature string `json:"signature"`
	Response  string `json:"response"`
}

type RelayWebsocketErrorResponse struct {
	Error    error                   `json:"error"`
	Dispatch *types.DispatchResponse `json:"dispatch"`
}

func RelayWebsocket(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return cors(&w, r)
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	var relay = types.Relay{}
	if err := PopModel(w, r, ps, &relay); err != nil {
		response := RelayWebsocketErrorResponse{
			Error: err,
		}
		j, _ := json.Marshal(response)
		conn.WriteMessage(websocket.TextMessage, j)
		return
	}

	dispatch, err := app.VCA.HandleWebsocketRelay(relay, conn)
	if err != nil {
		response := RelayWebsocketErrorResponse{
			Error:    err,
			Dispatch: dispatch,
		}
		j, _ := json.Marshal(response)
		conn.WriteMessage(websocket.TextMessage, j)
		return
	}
}
