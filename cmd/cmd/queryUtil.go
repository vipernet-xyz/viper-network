package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"

	"github.com/vipernet-xyz/viper-network/app"
	"github.com/vipernet-xyz/viper-network/rpc"
)

var (
	SendRawTxPath,
	GetNodePath,
	GetACLPath,
	GetUpgradePath,
	GetDAOOwnerPath,
	GetHeightPath,
	GetAccountPath,
	GetProviderPath,
	GetTxPath,
	GetBlockPath,
	GetSupportedChainsPath,
	GetBalancePath,
	GetAccountTxsPath,
	GetNodeParamsPath,
	GetServicersPath,
	GetSigningInfoPath,
	GetProvidersPath,
	GetProviderParamsPath,
	GetViperParamsPath,
	GetNodeClaimsPath,
	GetNodeClaimPath,
	GetBlockTxsPath,
	GetSupplyPath,
	GetAllParamsPath,
	GetParamPath,
	GetStopPath,
	GetQueryChains,
	GetAccountsPath string
)

func init() {
	routes := rpc.GetRoutes()
	for _, route := range routes {
		switch route.Name {
		case "SendRawTx":
			SendRawTxPath = route.Path
		case "QueryNode":
			GetNodePath = route.Path
		case "QueryACL":
			GetACLPath = route.Path
		case "QueryUpgrade":
			GetUpgradePath = route.Path
		case "QueryDAOOwner":
			GetDAOOwnerPath = route.Path
		case "QueryHeight":
			GetHeightPath = route.Path
		case "QueryAccount":
			GetAccountPath = route.Path
		case "QueryAccounts":
			GetAccountsPath = route.Path
		case "QueryProvider":
			GetProviderPath = route.Path
		case "QueryTX":
			GetTxPath = route.Path
		case "QueryBlock":
			GetBlockPath = route.Path
		case "QuerySupportedChains":
			GetSupportedChainsPath = route.Path
		case "QueryBalance":
			GetBalancePath = route.Path
		case "QueryAccountTxs":
			GetAccountTxsPath = route.Path
		case "QueryNodeParams":
			GetNodeParamsPath = route.Path
		case "QueryServicers":
			GetServicersPath = route.Path
		case "QuerySigningInfo":
			GetSigningInfoPath = route.Path
		case "QueryProviders":
			GetProvidersPath = route.Path
		case "QueryProviderParams":
			GetProviderParamsPath = route.Path
		case "QueryViperParams":
			GetViperParamsPath = route.Path
		case "QueryBlockTxs":
			GetBlockTxsPath = route.Path
		case "QuerySupply":
			GetSupplyPath = route.Path
		case "QueryNodeClaim":
			GetNodeClaimPath = route.Path
		case "QueryNodeClaims":
			GetNodeClaimsPath = route.Path
		case "QueryAllParams":
			GetAllParamsPath = route.Path
		case "QueryParam":
			GetParamPath = route.Path
		case "Stop":
			GetStopPath = route.Path
		case "QueryChains":
			GetQueryChains = route.Path
		default:
			continue
		}
	}
}

func QueryRPC(path string, jsonArgs []byte) (string, error) {
	//cliURL := app.GlobalConfig.ViperConfig.RemoteCLIURL + ":" + app.GlobalConfig.ViperConfig.RPCPort + path
	cliURL := app.GlobalConfig.ViperConfig.RemoteCLIURL + path
	types.SetRPCTimeout(app.GlobalConfig.ViperConfig.RPCTimeout)
	fmt.Println(cliURL)
	req, err := http.NewRequest("POST", cliURL, bytes.NewBuffer(jsonArgs))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "provider/json")
	client := &http.Client{
		Timeout: types.GetRPCTimeout() * time.Millisecond,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	res, err := strconv.Unquote(string(bz))
	if err == nil {
		bz = []byte(res)
	}
	if resp.StatusCode == http.StatusOK {
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, bz, "", "    ")
		if err == nil {
			return prettyJSON.String(), nil
		}
		return string(bz), nil
	}
	return "", fmt.Errorf("the http status code was not okay: %d, and the status was: %s, with a response of %v", resp.StatusCode, resp.Status, string(bz))
}

func QuerySecuredRPC(path string, jsonArgs []byte, token sdk.AuthToken) (string, error) {
	//cliURL := app.GlobalConfig.ViperConfig.RemoteCLIURL + ":" + app.GlobalConfig.ViperConfig.RPCPort + path
	cliURL := app.GlobalConfig.ViperConfig.RemoteCLIURL + path
	types.SetRPCTimeout(app.GlobalConfig.ViperConfig.RPCTimeout)
	fmt.Println(cliURL)
	req, err := http.NewRequest("POST", cliURL, bytes.NewBuffer(jsonArgs))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "provider/json")
	q := req.URL.Query()
	q.Add("authtoken", token.Value)
	req.URL.RawQuery = q.Encode()
	client := &http.Client{
		Timeout: types.GetRPCTimeout() * time.Millisecond,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	res, err := strconv.Unquote(string(bz))
	if err == nil {
		bz = []byte(res)
	}
	if resp.StatusCode == http.StatusOK {
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, bz, "", "    ")
		if err == nil {
			return prettyJSON.String(), nil
		}
		return string(bz), nil
	}
	return "", fmt.Errorf("the http status code was not okay: %d, and the status was: %s, with a response of %v", resp.StatusCode, resp.Status, string(bz))
}
