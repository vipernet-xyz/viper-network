package baseapp

import (
	"fmt"
	"os"

	abci "github.com/tendermint/tendermint/abci/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

type queryRouter struct {
	routes map[string]sdk.Querier
}

// GRPCQueryRouter routes ABCI Query requests to GRPC handlers
type GRPCQueryRouter struct {
	routes      map[string]GRPCQueryHandler
	cdc         encoding.Codec
	serviceData []serviceData
}

// serviceData represents a gRPC service, along with its handler.
type serviceData struct {
	serviceDesc *grpc.ServiceDesc
	handler     interface{}
}

// NewGRPCQueryRouter creates a new GRPCQueryRouter
func NewGRPCQueryRouter() *GRPCQueryRouter {
	return &GRPCQueryRouter{
		routes: map[string]GRPCQueryHandler{},
	}
}

// GRPCQueryHandler defines a function type which handles ABCI Query requests
// using gRPC
type GRPCQueryHandler = func(ctx sdk.Context, req abci.RequestQuery) (abci.ResponseQuery, error)

var _ sdk.QueryRouter = NewQueryRouter()

// NewQueryRouter returns a reference to a new queryRouter.
//
// TODO: Either make the function private or make return type (queryRouter) public.
func NewQueryRouter() *queryRouter { // nolint: golint
	return &queryRouter{
		routes: map[string]sdk.Querier{},
	}
}

// AddRoute adds a query path to the router with a given Querier. It will panic
// if a duplicate route is given. The route must be alphanumeric.
func (qrt *queryRouter) AddRoute(path string, q sdk.Querier) sdk.QueryRouter {
	if !isAlphaNumeric(path) {
		fmt.Println("route expressions can only contain alphanumeric characters")
		os.Exit(1)
	}
	if qrt.routes[path] != nil {
		fmt.Println(fmt.Errorf("route %s has already been initialized", path))
		os.Exit(1)
	}

	qrt.routes[path] = q
	return qrt
}

// Route returns the Querier for a given query route path.
func (qrt *queryRouter) Route(path string) sdk.Querier {
	return qrt.routes[path]
}
