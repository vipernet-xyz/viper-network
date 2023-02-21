package rpc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	sdk "github.com/vipernet-xyz/viper-network/types"
	types2 "github.com/vipernet-xyz/viper-network/x/authentication/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/types"

	"github.com/vipernet-xyz/viper-network/app"
	providersTypes "github.com/vipernet-xyz/viper-network/x/providers/types"
	servicerTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"

	"github.com/julienschmidt/httprouter"
	core_types "github.com/tendermint/tendermint/rpc/core/types"
)

func Version(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	WriteResponse(w, APIVersion, r.URL.Path, r.Host)
}

type HeightParams struct {
	Height int64 `json:"height"`
}
type HeightAndKeyParams struct {
	Height int64  `json:"height"`
	Key    string `json:"key"`
}

type HashAndProveParams struct {
	Hash  string `json:"hash"`
	Prove bool   `json:"prove"`
}

type HeightAndAddrParams struct {
	Height  int64  `json:"height"`
	Address string `json:"address"`
}

type HeightAndValidatorOptsParams struct {
	Height int64                               `json:"height"`
	Opts   servicerTypes.QueryValidatorsParams `json:"opts"`
}

type HeightAndProviderOptsParams struct {
	Height int64                                 `json:"height"`
	Opts   providersTypes.QueryProvidersWithOpts `json:"opts"`
}

type PaginateAddrParams struct {
	Address  string `json:"address"`
	Page     int    `json:"page,omitempty"`
	PerPage  int    `json:"per_page,omitempty"`
	Received bool   `json:"received,omitempty"`
	Prove    bool   `json:"prove,omitempty"`
	Sort     string `json:"order,omitempty"`
}

type PaginatedHeightParams struct {
	Height  int64  `json:"height"`
	Page    int    `json:"page,omitempty"`
	PerPage int    `json:"per_page,omitempty"`
	Prove   bool   `json:"prove,omitempty"`
	Sort    string `json:"order,omitempty"`
}

type PaginatedHeightAndAddrParams struct {
	Height  int64  `json:"height"`
	Addr    string `json:"address"`
	Page    int    `json:"page,omitempty"`
	PerPage int    `json:"per_page,omitempty"`
}

func Block(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryBlock(&params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(res), r.URL.Path, r.Host)
}

func Tx(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HashAndProveParams{}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	res, err := app.VCA.QueryTx(params.Hash, params.Prove)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	rpcResponse := ResultTxToRPC(res)
	s, er := json.MarshalIndent(rpcResponse, "", "  ")
	if er != nil {
		WriteErrorResponse(w, 400, er.Error())
		return
	}
	WriteJSONResponse(w, string(s), r.URL.Path, r.Host)
}

// Result of searching for txs
type RPCResultTxSearch struct {
	Txs       []*RPCResultTx `json:"txs"`
	PageCount int            `json:"page_count"`
	TotalTxs  int            `json:"total_txs"`
}

// Result of querying for a tx
type RPCResultTx struct {
	Hash     bytes.HexBytes       `json:"hash"`
	Height   int64                `json:"height"`
	Index    uint32               `json:"index"`
	TxResult RPCResponseDeliverTx `json:"tx_result"`
	Tx       types.Tx             `json:"tx"`
	Proof    types.TxProof        `json:"proof,omitempty"`
	StdTx    RPCStdTx             `json:"stdTx,omitempty"`
}

type RPCResponseDeliverTx struct {
	Code        uint32        `json:"code"`
	Data        []byte        `json:"data"`
	Log         string        `json:"log"`
	Info        string        `json:"info"`
	Events      []abci.Event  `json:"events"`
	Codespace   string        `json:"codespace"`
	Signer      types.Address `json:"signer"`
	Recipient   types.Address `json:"recipient"`
	MessageType string        `json:"message_type"`
}

type RPCStdTx types2.StdTx

type rPCStdTx struct {
	Msg       json.RawMessage `json:"msg" yaml:"msg"`
	Fee       sdk.Coins       `json:"fee" yaml:"fee"`
	Signature RPCStdSignature `json:"signature" yaml:"signature"`
	Memo      string          `json:"memo" yaml:"memo"`
	Entropy   int64           `json:"entropy" yaml:"entropy"`
}

type RPCStdSignature struct {
	PublicKey string `json:"pub_key"`
	Signature string `json:"signature"`
}

func (r RPCStdTx) MarshalJSON() ([]byte, error) {
	if r.Msg == nil {
		return json.Marshal(rPCStdTx{})
	}
	msgBz := (types2.StdTx)(r).Msg.GetSignBytes()
	sig := RPCStdSignature{
		PublicKey: r.Signature.RawString(),
		Signature: hex.EncodeToString(r.Signature.Signature),
	}
	return json.Marshal(rPCStdTx{
		Msg:       msgBz,
		Fee:       r.Fee,
		Signature: sig,
		Memo:      r.Memo,
		Entropy:   r.Entropy,
	})
}

func ResultTxSearchToRPC(res *core_types.ResultTxSearch) RPCResultTxSearch {
	if res == nil {
		return RPCResultTxSearch{}
	}
	pageCount := len(res.Txs)
	rpcTxSearch := RPCResultTxSearch{
		Txs:       make([]*RPCResultTx, 0, res.TotalCount),
		PageCount: pageCount,
		TotalTxs:  res.TotalCount,
	}
	for _, result := range res.Txs {
		rpcTxSearch.Txs = append(rpcTxSearch.Txs, ResultTxToRPC(result))
	}
	return rpcTxSearch
}

func ResultTxToRPC(res *core_types.ResultTx) *RPCResultTx {
	if res == nil {
		return nil
	}
	if app.GlobalConfig.ViperConfig.DisableTxEvents {
		res.TxResult.Events = nil
	}
	rpcDeliverTx := RPCResponseDeliverTx{
		Code:        res.TxResult.Code,
		Data:        res.TxResult.Data,
		Log:         res.TxResult.Log,
		Info:        res.TxResult.Info,
		Events:      res.TxResult.Events,
		Codespace:   res.TxResult.Codespace,
		Signer:      res.TxResult.Signer,
		Recipient:   res.TxResult.Recipient,
		MessageType: res.TxResult.MessageType,
	}

	r := &RPCResultTx{
		Hash:     res.Hash,
		Height:   res.Height,
		Index:    res.Index,
		TxResult: rpcDeliverTx,
		Tx:       res.Tx,
		Proof:    res.Proof,
	}
	tx, err := app.UnmarshalTx(res.Tx, res.Height)
	if err != nil {
		fmt.Println("an error occurred unmarshalling the transaction", err.Error())
		return r
	}
	r.StdTx = RPCStdTx(tx)
	return r
}

func AccountTxs(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = PaginateAddrParams{}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	var res *core_types.ResultTxSearch
	var err error
	if !params.Received {
		res, err = app.VCA.QueryAccountTxs(params.Address, params.Page, params.PerPage, params.Prove, params.Sort)
	} else {
		res, err = app.VCA.QueryRecipientTxs(params.Address, params.Page, params.PerPage, params.Prove, params.Sort)
	}
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	rpcResponse := ResultTxSearchToRPC(res)
	s, er := json.MarshalIndent(rpcResponse, "", "  ")
	if er != nil {
		WriteErrorResponse(w, 400, er.Error())
		return
	}
	WriteJSONResponse(w, string(s), r.URL.Path, r.Host)
}

func BlockTxs(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = PaginatedHeightParams{}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryBlockTxs(params.Height, params.Page, params.PerPage, params.Prove, params.Sort)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
	}
	rpcResponse := ResultTxSearchToRPC(res)
	s, er := json.MarshalIndent(rpcResponse, "", "  ")
	if er != nil {
		WriteErrorResponse(w, 400, er.Error())
		return
	}
	WriteJSONResponse(w, string(s), r.URL.Path, r.Host)
}

func AllBlockTxs(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = PaginatedHeightParams{}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryAllBlockTxs(params.Height, params.Page, params.PerPage)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
	}
	rpcResponse := ResultTxSearchToRPC(res)
	s, er := json.MarshalIndent(rpcResponse, "", "  ")
	if er != nil {
		WriteErrorResponse(w, 400, er.Error())
		return
	}
	WriteJSONResponse(w, string(s), r.URL.Path, r.Host)
}

type queryHeightResponse struct {
	Height int64 `json:"height"`
}

func Height(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	res, err := app.VCA.QueryHeight()
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	height, err := json.Marshal(&queryHeightResponse{Height: res})
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(height), r.URL.Path, r.Host)
}

type queryBalanceResponse struct {
	Balance *big.Int `json:"balance"`
}

func Balance(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightAndAddrParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	balance, err := app.VCA.QueryBalance(params.Address, params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	s, err := json.MarshalIndent(&queryBalanceResponse{Balance: balance.BigInt()}, "", "")
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(s), r.URL.Path, r.Host)
}

func Account(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightAndAddrParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryAccount(params.Address, params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	s, err := json.Marshal(res)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(s), r.URL.Path, r.Host)
}

func Accounts(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = PaginatedHeightParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryAccounts(params.Height, params.Page, params.PerPage)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	s, err := json.Marshal(res)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(s), r.URL.Path, r.Host)
}

func Servicers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightAndValidatorOptsParams{}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	if params.Opts.Page == 0 {
		params.Opts.Page = 1
	}
	if params.Opts.Limit == 0 {
		params.Opts.Limit = 1000
	}
	res, err := app.VCA.QueryServicers(params.Height, params.Opts)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, err := res.JSON()
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	w.Header().Set("Content-Type", "provider/json; charset=UTF-8")
	_, err = w.Write(j)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
	}
}

func Node(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightAndAddrParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryServicer(params.Address, params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, err := res.MarshalJSON()
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

func SigningInfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = PaginatedHeightAndAddrParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QuerySigningInfos(params.Addr, params.Height, params.Page, params.PerPage)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, err := res.JSON()
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

func SecondUpgrade(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	result := app.Codec().IsAfterValidatorSplitUpgrade(params.Height)
	j, _ := json.Marshal(struct {
		R bool `json:"r"`
	}{R: result})

	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

func Chains(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	value := r.URL.Query().Get("authtoken")
	if value == app.AuthToken.Value {
		res, err := app.VCA.QueryHostedChains()
		if err != nil {
			WriteErrorResponse(w, 400, err.Error())
			return
		}
		j, err := app.Codec().MarshalJSON(res)
		if err != nil {
			WriteErrorResponse(w, 400, err.Error())
			return
		}
		WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
	} else {
		WriteErrorResponse(w, 401, "wrong authtoken "+value)
	}
}

func NodeParams(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	res, err := app.VCA.QueryServicerParams(params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, err := app.Codec().MarshalJSON(res)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

func QueryValidatorsByChain(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightAndValidatorOptsParams{
		Height: 0,
		Opts: servicerTypes.QueryValidatorsParams{
			Blockchain: "0001",
		},
	}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	res, err := app.VCA.QueryValidatorByChain(params.Height, params.Opts.Blockchain)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}

	j, _ := json.Marshal(struct {
		Chain string `json:"chain"`
		Count string `json:"count"`
	}{Chain: params.Opts.Blockchain, Count: strconv.FormatInt(res, 10)})

	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

type QueryNodeReceiptParam struct {
	Address        string `json:"address"`
	Blockchain     string `json:"blockchain"`
	ProviderPubkey string `json:"provider_pubkey"`
	SBlockHeight   int64  `json:"session_block_height"`
	Height         int64  `json:"height"`
	ReceiptType    string `json:"receipt_type"`
}

func NodeClaim(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = QueryNodeReceiptParam{}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryClaim(params.Address, params.ProviderPubkey, params.Blockchain, params.ReceiptType, params.SBlockHeight, params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, err := app.Codec().MarshalJSON(res)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

func NodeClaims(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = PaginatedHeightAndAddrParams{}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryClaims(params.Addr, params.Height, params.Page, params.PerPage)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, err := res.JSON()
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

func Providers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightAndProviderOptsParams{}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	if params.Opts.Page == 0 {
		params.Opts.Page = 1
	}
	if params.Opts.Limit == 0 {
		params.Opts.Limit = 1000
	}
	res, err := app.VCA.QueryProviders(params.Height, params.Opts)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, err := res.JSON()
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	w.Header().Set("Content-Type", "provider/json; charset=UTF-8")
	_, err = w.Write(j)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
	}
}

func App(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightAndAddrParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryProvider(params.Address, params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, err := res.MarshalJSON()
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

func AppParams(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryProviderParams(params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, err := app.Codec().MarshalJSON(res)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

func ViperParams(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryViperParams(params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, err := app.Codec().MarshalJSON(res)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

func SupportedChains(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryViperSupportedBlockchains(params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, err := app.Codec().MarshalJSON(res)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteResponse(w, string(j), r.URL.Path, r.Host)
}

type querySupplyResponse struct {
	NodeStaked    string `json:"servicer_staked"`
	AppStaked     string `json:"app_staked"`
	Dao           string `json:"dao"`
	TotalStaked   string `json:"total_staked"`
	TotalUnstaked string `json:"total_unstaked"`
	Total         string `json:"total"`
}

func Supply(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	servicersStake, total, err := app.VCA.QueryTotalServicerCoins(params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	providersStaked, err := app.VCA.QueryTotalProviderCoins(params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	dao, err := app.VCA.QueryDaoBalance(params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	totalStaked := servicersStake.Add(providersStaked).Add(dao)
	totalUnstaked := total.Sub(totalStaked)
	res, err := json.MarshalIndent(&querySupplyResponse{
		NodeStaked:    servicersStake.String(),
		AppStaked:     providersStaked.String(),
		Dao:           dao.String(),
		TotalStaked:   totalStaked.BigInt().String(),
		TotalUnstaked: totalUnstaked.BigInt().String(),
		Total:         total.BigInt().String(),
	}, "", "  ")
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(res), r.URL.Path, r.Host)
}

func DAOOwner(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryDaoOwner(0)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	s, err := json.Marshal(res)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteResponse(w, string(s), r.URL.Path, r.Host)
}

func Upgrade(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryUpgrade(params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	s, err := json.Marshal(res)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteResponse(w, string(s), r.URL.Path, r.Host)
}

func ACL(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryACL(params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, err := app.Codec().MarshalJSON(res)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteResponse(w, string(j), r.URL.Path, r.Host)
}

func AllParams(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryAllParams(params.Height)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, err := app.Codec().MarshalJSON(res)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}
func Param(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightAndKeyParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.QueryParam(params.Height, params.Key)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	j, err := app.Codec().MarshalJSON(res)
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteJSONResponse(w, string(j), r.URL.Path, r.Host)
}

func State(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var params = HeightParams{Height: 0}
	if err := PopModel(w, r, ps, &params); err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	if params.Height == 0 {
		params.Height = app.VCA.BaseApp.LastBlockHeight()
	}
	res, err := app.VCA.ExportState(params.Height, "")
	if err != nil {
		WriteErrorResponse(w, 400, err.Error())
		return
	}
	WriteRaw(w, res, r.URL.Path, r.Host)
}
