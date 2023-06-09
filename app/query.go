package app

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"

	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication/exported"
	"github.com/vipernet-xyz/viper-network/x/authentication/util"
	"github.com/vipernet-xyz/viper-network/x/governance/types"
	providersTypes "github.com/vipernet-xyz/viper-network/x/providers/types"
	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
	viperTypes "github.com/vipernet-xyz/viper-network/x/vipernet/types"

	core_types "github.com/tendermint/tendermint/rpc/core/types"
)

const (
	messageSenderQuery     = "tx.signer='%s'"
	transferRecipientQuery = "tx.recipient='%s'"
	txHeightQuery          = "tx.height=%d"
)

// zero for height = latest
func (app ViperCoreApp) QueryBlock(height *int64) (blockJSON []byte, err error) {
	tmClient := app.GetClient()
	defer func() { _ = tmClient.Stop() }()
	b, err := tmClient.Block(height)
	if err != nil {
		return nil, err
	}
	return Codec().MarshalJSONIndent(b, "", "  ")
}

func (app ViperCoreApp) QueryTx(hash string, prove bool) (res *core_types.ResultTx, err error) {
	tmClient := app.GetClient()
	defer func() { _ = tmClient.Stop() }()
	h, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}
	res, err = tmClient.Tx(h, prove)
	return
}

func (app ViperCoreApp) QueryAccountTxs(addr string, page, perPage int, prove bool, sort string) (res *core_types.ResultTxSearch, err error) {
	tmClient := app.GetClient()
	defer func() { _ = tmClient.Stop() }()
	_, err = hex.DecodeString(addr)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(messageSenderQuery, addr)
	page, perPage = checkPagination(page, perPage)
	res, err = tmClient.TxSearch(query, prove, page, perPage, checkSort(sort))
	return
}
func (app ViperCoreApp) QueryRecipientTxs(addr string, page, perPage int, prove bool, sort string) (res *core_types.ResultTxSearch, err error) {
	tmClient := app.GetClient()
	defer func() { _ = tmClient.Stop() }()
	_, err = hex.DecodeString(addr)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(transferRecipientQuery, addr)
	page, perPage = checkPagination(page, perPage)
	res, err = tmClient.TxSearch(query, prove, page, perPage, checkSort(sort))
	return
}

func (app ViperCoreApp) QueryBlockTxs(height int64, page, perPage int, prove bool, sort string) (res *core_types.ResultTxSearch, err error) {
	tmClient := app.GetClient()
	defer func() { _ = tmClient.Stop() }()
	query := fmt.Sprintf(txHeightQuery, height)
	page, perPage = checkPagination(page, perPage)
	res, err = tmClient.TxSearch(query, prove, page, perPage, checkSort(sort))
	return
}

func (app ViperCoreApp) QueryAllBlockTxs(height int64, page, perPage int) (res *core_types.ResultTxSearch, err error) {
	res = &core_types.ResultTxSearch{}
	tmClient := app.GetClient()
	defer func() { _ = tmClient.Stop() }()
	page, perPage = checkPagination(page, perPage)
	b, err := tmClient.Block(&height)
	if err != nil {
		return nil, err
	}
	skip := (page - 1) * perPage
	b1, err := tmClient.BlockResults(&height)
	if err != nil {
		return nil, err
	}
	res.TotalCount = len(b1.TxsResults) // this
	for i, t := range b1.TxsResults {
		if i < skip {
			continue
		}
		tx := b.Block.Txs[i]
		res.Txs = append(res.Txs, &core_types.ResultTx{
			Hash:     tx.Hash(),
			Height:   height,
			Index:    uint32(i),
			TxResult: *t,
			Tx:       tx,
			Proof:    tmtypes.TxProof{},
		})
		if len(res.Txs) >= perPage {
			break
		}
	}
	return
}

func (app ViperCoreApp) QueryHeight() (res int64, err error) {
	tmClient := app.GetClient()
	defer func() { _ = tmClient.Stop() }()
	status, err := tmClient.Status()
	if err != nil {
		return -1, err
	}

	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

func (app ViperCoreApp) QueryServicerStatus() (res *core_types.ResultStatus, err error) {
	tmClient := app.GetClient()
	defer func() { _ = tmClient.Stop() }()
	return tmClient.Status()
}

func (app ViperCoreApp) QueryBalance(addr string, height int64) (res sdk.BigInt, err error) {
	acc, err := app.QueryAccount(addr, height)
	if err != nil {
		return
	}
	if (*acc) == nil {
		return sdk.NewInt(0), nil
	}
	return (*acc).GetCoins().AmountOf(sdk.DefaultStakeDenom), nil
}

func (app ViperCoreApp) QueryAccount(addr string, height int64) (res *exported.Account, err error) {
	a, err := sdk.AddressFromHex(addr)
	if err != nil {
		return nil, err
	}
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	acc := app.accountKeeper.GetAccount(ctx, a)
	return &acc, nil
}

func (app ViperCoreApp) QueryAccounts(height int64, page, perPage int) (res Page, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	page, perPage = checkPagination(page, perPage)
	accs := app.accountKeeper.GetAllAccountsExport(ctx)
	return paginate(page, perPage, accs, 10000)
}

func (app ViperCoreApp) QueryServicers(height int64, opts servicersTypes.QueryValidatorsParams) (res Page, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	opts.Page, opts.Limit = checkPagination(opts.Page, opts.Limit)
	servicers := app.servicersKeeper.GetAllValidatorsWithOpts(ctx, opts)
	return paginate(opts.Page, opts.Limit, servicers, int(app.servicersKeeper.MaxValidators(ctx)))
}

func (app ViperCoreApp) QueryServicer(addr string, height int64) (res servicersTypes.Validator, err error) {
	a, err := sdk.AddressFromHex(addr)
	if err != nil {
		return res, err
	}
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	res, found := app.servicersKeeper.GetValidator(ctx, a)
	if !found {
		err = fmt.Errorf("validator not found for %s", a.String())
	}
	return
}

func (app ViperCoreApp) QueryServicerParams(height int64) (res servicersTypes.Params, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	return app.servicersKeeper.GetParams(ctx), nil
}

func (app ViperCoreApp) QueryHostedChains() (res map[string]viperTypes.HostedBlockchain, err error) {
	return app.viperKeeper.GetHostedBlockchains().M, nil
}

func (app ViperCoreApp) SetHostedChains(req map[string]viperTypes.HostedBlockchain) (res map[string]viperTypes.HostedBlockchain, err error) {
	return app.viperKeeper.SetHostedBlockchains(req).M, nil
}

func (app ViperCoreApp) QuerySigningInfo(height int64, addr string) (res servicersTypes.ValidatorSigningInfo, err error) {
	a, err := sdk.AddressFromHex(addr)
	if err != nil {
		return servicersTypes.ValidatorSigningInfo{}, err
	}
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	res, found := app.servicersKeeper.GetValidatorSigningInfo(ctx, a)
	if !found {
		err = fmt.Errorf("signing info not found for %s", a.String())
	}
	return
}

func (app ViperCoreApp) QuerySigningInfos(address string, height int64, page, perPage int) (res Page, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	signingInfos := make([]servicersTypes.ValidatorSigningInfo, 0)

	page, perPage = checkPagination(page, perPage)
	if address != "" {
		addr, err := sdk.AddressFromHex(address)
		if err != nil {
			return Page{}, err
		}
		sinfo, found := app.servicersKeeper.GetValidatorSigningInfo(ctx, addr)
		if !found {
			return Page{}, err
		}
		signingInfos = append(signingInfos, sinfo)
	} else {
		app.servicersKeeper.IterateAndExecuteOverValSigningInfo(ctx, func(address sdk.Address, info servicersTypes.ValidatorSigningInfo) (stop bool) {
			signingInfos = append(signingInfos, info)
			return false
		})
	}
	return paginate(page, perPage, signingInfos, int(app.servicersKeeper.MaxValidators(ctx)))
}

func (app ViperCoreApp) QueryTotalServicerCoins(height int64) (stakedTokens sdk.BigInt, totalTokens sdk.BigInt, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	stakedTokens = app.servicersKeeper.GetStakedTokens(ctx)
	totalTokens = app.servicersKeeper.TotalTokens(ctx)
	return
}

func (app ViperCoreApp) QueryDaoBalance(height int64) (res sdk.BigInt, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	return app.governanceKeeper.GetDAOTokens(ctx), nil
}

func (app ViperCoreApp) QueryDaoOwner(height int64) (res sdk.Address, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	return app.governanceKeeper.GetDAOOwner(ctx), nil
}

func (app ViperCoreApp) QueryUpgrade(height int64) (res types.Upgrade, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	return app.governanceKeeper.GetUpgrade(ctx), nil
}

func (app ViperCoreApp) QueryACL(height int64) (res types.ACL, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	return app.governanceKeeper.GetACL(ctx), nil
}

type AllParamsReturn struct {
	AppParams   []SingleParamReturn `json:"app_params"`
	NodeParams  []SingleParamReturn `json:"node_params"`
	ViperParams []SingleParamReturn `json:"viper_params"`
	GovParams   []SingleParamReturn `json:"governance_params"`
	AuthParams  []SingleParamReturn `json:"auth_params"`
}

type SingleParamReturn struct {
	Key   string `json:"param_key"`
	Value string `json:"param_value"`
}

func (app ViperCoreApp) QueryAllParams(height int64) (r AllParamsReturn, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	//get all the parameters from governance module
	allmap := app.governanceKeeper.GetAllParamNameValue(ctx)

	//transform for easy handling
	for k, v := range allmap {
		sub, _ := types.SplitACLKey(k)
		s, err2 := strconv.Unquote(v)
		if err2 != nil {
			//ignoring this error as content is a json object
			s = v
		}
		switch sub {
		case "pos":
			r.NodeParams = append(r.NodeParams, SingleParamReturn{
				Key:   k,
				Value: s,
			})
		case "application":
			r.AppParams = append(r.AppParams, SingleParamReturn{
				Key:   k,
				Value: s,
			})
		case "vipernet":
			r.ViperParams = append(r.ViperParams, SingleParamReturn{
				Key:   k,
				Value: s,
			})
		case "governance":
			r.GovParams = append(r.GovParams, SingleParamReturn{
				Key:   k,
				Value: s,
			})
		case "authentication":
			r.AuthParams = append(r.AuthParams, SingleParamReturn{
				Key:   k,
				Value: s,
			})
		default:
		}
	}

	return r, nil
}

func (app ViperCoreApp) QueryParam(height int64, paramkey string) (r SingleParamReturn, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	//get all the parameters from governance module
	allmap := app.governanceKeeper.GetAllParamNameValue(ctx)

	if val, ok := allmap[paramkey]; ok {
		r.Key = paramkey
		s, err2 := strconv.Unquote(val)
		if err2 != nil {
			//ignoring this error as content is a json object
			r.Value = val
			return r, err
		}
		r.Value = s
	}
	return
}

func (app ViperCoreApp) QueryProviders(height int64, opts providersTypes.QueryProvidersWithOpts) (res Page, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	opts.Page, opts.Limit = checkPagination(opts.Page, opts.Limit)
	applications := app.providersKeeper.GetAllProvidersWithOpts(ctx, opts)
	return paginate(opts.Page, opts.Limit, applications, int(app.providersKeeper.GetParams(ctx).MaxProviders))
}

func (app ViperCoreApp) QueryProvider(addr string, height int64) (res providersTypes.Provider, err error) {
	a, err := sdk.AddressFromHex(addr)
	if err != nil {
		return res, err
	}
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	res, found := app.providersKeeper.GetProvider(ctx, a)
	if !found {
		err = providersTypes.ErrNoProviderFound(providersTypes.ModuleName)
		return
	}
	return
}

func (app ViperCoreApp) QueryTotalProviderCoins(height int64) (staked sdk.BigInt, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	return app.providersKeeper.GetStakedTokens(ctx), nil
}

func (app ViperCoreApp) QueryProviderParams(height int64) (res providersTypes.Params, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	return app.providersKeeper.GetParams(ctx), nil
}

func (app ViperCoreApp) QueryValidatorByChain(height int64, chain string) (amount int64, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	_, count := app.servicersKeeper.GetValidatorsByChain(ctx, chain)
	return int64(count), nil
}

func (app ViperCoreApp) QueryViperSupportedBlockchains(height int64) (res []string, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	sb := app.viperKeeper.SupportedBlockchains(ctx)
	return sb, nil
}

func (app ViperCoreApp) QueryClaim(address, providerPubkey, chain, evidenceType string, sessionBlockHeight int64, height int64) (res *viperTypes.MsgClaim, err error) {
	a, err := sdk.AddressFromHex(address)
	if err != nil {
		return nil, err
	}
	header := viperTypes.SessionHeader{
		ProviderPubKey:     providerPubkey,
		Chain:              chain,
		SessionBlockHeight: sessionBlockHeight,
	}
	err = header.ValidateHeader()
	if err != nil {
		return nil, err
	}
	et, err := viperTypes.EvidenceTypeFromString(evidenceType)
	if err != nil {
		return nil, err
	}
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	claim, found := app.viperKeeper.GetClaim(ctx, a, header, et)
	if !found {
		return nil, viperTypes.NewClaimNotFoundError(viperTypes.ModuleName)
	}
	return &claim, nil
}

func (app ViperCoreApp) QueryClaims(address string, height int64, page, perPage int) (res Page, err error) {
	var a sdk.Address
	var claims []viperTypes.MsgClaim
	ctx, err := app.NewContext(height)
	page, perPage = checkPagination(page, perPage)
	if err != nil {
		return
	}
	if address != "" {
		a, err = sdk.AddressFromHex(address)
		if err != nil {
			return Page{}, err
		}
		claims, err = app.viperKeeper.GetClaims(ctx, a)
		if err != nil {
			return Page{}, err
		}
	} else {
		claims = app.viperKeeper.GetAllClaims(ctx)
	}
	p, err := paginate(page, perPage, claims, 10000)
	if err != nil {
		return Page{}, err
	}
	return p, nil
}

func (app ViperCoreApp) QueryViperParams(height int64) (res viperTypes.Params, err error) {
	ctx, err := app.NewContext(height)
	if err != nil {
		return
	}
	p := app.viperKeeper.GetParams(ctx)
	return p, nil
}

func (app ViperCoreApp) HandleChallenge(c viperTypes.ChallengeProofInvalidData) (res *viperTypes.ChallengeResponse, err error) {
	ctx, err := app.NewContext(app.LastBlockHeight())
	if err != nil {
		return nil, err
	}
	return app.viperKeeper.HandleChallenge(ctx, c)
}

func (app ViperCoreApp) HandleDispatch(header viperTypes.SessionHeader) (res *viperTypes.DispatchResponse, err error) {
	ctx, err := app.NewContext(app.LastBlockHeight())
	if err != nil {
		return nil, err
	}
	return app.viperKeeper.HandleDispatch(ctx, header)
}

func (app ViperCoreApp) HandleRelay(r viperTypes.Relay) (res *viperTypes.RelayResponse, dispatch *viperTypes.DispatchResponse, err error) {
	ctx, err := app.NewContext(app.LastBlockHeight())

	if err != nil {
		return nil, nil, err
	}

	status, sErr := app.viperKeeper.TmNode.ConsensusReactorStatus()
	if sErr != nil {
		return nil, nil, fmt.Errorf("viper node is unable to retrieve synced status from tendermint node, cannot service in this state")
	}

	if status.IsCatchingUp {
		return nil, nil, fmt.Errorf("viper node is currently syncing to the blockchain, cannot service in this state")
	}
	res, err = app.viperKeeper.HandleRelay(ctx, r)
	var err1 error
	if err != nil && viperTypes.ErrorWarrantsDispatch(err) {
		dispatch, err1 = app.HandleDispatch(r.Proof.SessionHeader())
		if err1 != nil {
			return
		}
	}
	return
}

func checkPagination(page, limit int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 30
	}
	return page, limit
}
func checkSort(s string) string {
	switch s {
	case "asc":
		return s
	case "desc":
		return s
	default:
		return "desc"
	}
}

func paginate(page, limit int, items interface{}, max int) (res Page, error error) {
	slice, success := takeArg(items, reflect.Slice)
	if !success {
		return Page{}, fmt.Errorf("invalid argument, non slice input to paginate")
	}
	l := slice.Len()
	start, end := util.Paginate(l, page, limit, max)
	if start == -1 && end == -1 {
		return Page{}, nil
	}
	if start < 0 || end < 0 {
		return Page{}, fmt.Errorf("invalid bounds error: start %d finish %d", start, end)
	} else {
		items = slice.Slice(start, end).Interface()
	}
	totalPages := int(math.Ceil(float64(l) / float64(end-start)))
	if totalPages < 1 {
		totalPages = 1
	}
	return Page{Result: items, Total: totalPages, Page: page}, nil
}

func takeArg(arg interface{}, kind reflect.Kind) (val reflect.Value, ok bool) {
	val = reflect.ValueOf(arg)
	if val.Kind() == kind {
		ok = true
	}
	return
}

type Page struct {
	Result interface{} `json:"result"`
	Total  int         `json:"total_pages"`
	Page   int         `json:"page"`
}

// Marshals struct into JSON
func (p Page) JSON() (out []byte, err error) {
	// each element should be a JSON
	return json.Marshal(p)
}

// String returns a human readable string representation of a validator page
func (p Page) String() string {
	return fmt.Sprintf("Total:\t\t%d\nPage:\t\t%d\nResult:\t\t\n====\n%v\n====\n", p.Total, p.Page, p.Result)
}
