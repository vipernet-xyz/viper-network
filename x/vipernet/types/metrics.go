package types

import (
	"context"
	"net/http"
	"sync"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdPrometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tendermint/tendermint/libs/log"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

var (
	globalServiceMetrics *ServiceMetrics
)

const (
	ServiceMetricsKey       = "service"
	ServiceMetricsNamespace = ServiceMetricsKey
	RelayCountName          = "relay_count_for_"
	RelayCountHelp          = "the number of relays executed against: "
	ChallengeCountName      = "challenge_count_for_"
	ChallengeCountHelp      = "the number of challenges executed against: "
	ErrCountName            = "err_count_for_"
	ErrCountHelp            = "the number of errors resulting from relays executed against: "
	AvgRelayHistName        = "avg_relay_time_for_"
	AvgrelayHistHelp        = "the average relay time in ms executed against: "
	SessionsCountName       = "sessions_count_for_"
	SessionsCountHelp       = "the number of unique sessions generated for: "
	uviprCountName          = "tokens_earned_for_"
	uviprCountHelp          = "the number of tokens earned in uvipr for : "
)

type ServiceMetrics struct {
	l               sync.Mutex
	tmLogger        log.Logger
	ServiceMetric   `json:"accumulated_service_metrics"` // total metrics
	NonNativeChains map[string]ServiceMetric             `json:"individual_service_metrics"` // metrics per chain
	GeoZone         `json:"geo_zone"`
	prometheusSrv   *http.Server
}

type ServiceMetricsEncodable struct {
	ServiceMetric   `json:"accumulated_service_metrics"` // total metrics
	NonNativeChains []ServiceMetric                      `json:"individual_service_metrics"` // metrics per chain
}

func GlobalServiceMetric() *ServiceMetrics {
	return globalServiceMetrics
}

func InitGlobalServiceMetric(hostedBlockchains *HostedBlockchains, hostedGeozone *HostedGeoZones, logger log.Logger, addr string, maxOpenConn int) {
	// create a new service metric
	serviceMetric := NewServiceMetrics(hostedBlockchains, hostedGeozone, logger)
	// set the service metrics
	globalServiceMetrics = serviceMetric
	// start metrics server
	globalServiceMetrics.prometheusSrv = globalServiceMetrics.StartPrometheusServer(addr, maxOpenConn)
}

func StopServiceMetrics() {
	if err := GlobalServiceMetric().prometheusSrv.Shutdown(context.Background()); err != nil {
		GlobalServiceMetric().tmLogger.Error("unable to shutdown service metrics server: ", err.Error())
	}
}

// startPrometheusServer starts a Prometheus HTTP server, listening for metrics
// collectors on addr.
func (sm *ServiceMetrics) StartPrometheusServer(addr string, maxOpenConn int) *http.Server {
	srv := &http.Server{
		Addr: ":" + addr,
		Handler: promhttp.InstrumentMetricHandler(
			stdPrometheus.DefaultRegisterer, promhttp.HandlerFor(
				stdPrometheus.DefaultGatherer,
				promhttp.HandlerOpts{MaxRequestsInFlight: maxOpenConn},
			),
		),
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// Error starting or closing listener:
			sm.tmLogger.Error("Prometheus HTTP server ListenAndServe", "err", err)
		}
	}()
	return srv
}

func (sm *ServiceMetrics) AddRelayFor(networkID string, nodeAddress *sdk.Address) {
	sm.l.Lock()
	defer sm.l.Unlock()
	// attempt to locate nn chain
	nnc, ok := sm.NonNativeChains[networkID]
	if !ok {
		sm.tmLogger.Error("unable to find corresponding networkID in service metrics: ", networkID)
		sm.NonNativeChains[networkID] = NewServiceMetricsFor(networkID)
		return
	}
	// add relay to accumulated count
	labels := sm.getValidatorLabel(nodeAddress)
	sm.RelayCount.With(labels...).Add(1)
	// add to individual relay count
	nnc.RelayCount.With(labels...).Add(1)
	// update nnc
	sm.NonNativeChains[networkID] = nnc
}

func (sm *ServiceMetrics) AddChallengeFor(networkID string, nodeAddress *sdk.Address) {
	sm.l.Lock()
	defer sm.l.Unlock()
	// attempt to locate nn chain
	nnc, ok := sm.NonNativeChains[networkID]
	if !ok {
		sm.tmLogger.Error("unable to find corresponding networkID in service metrics: ", networkID)
		sm.NonNativeChains[networkID] = NewServiceMetricsFor(networkID)
		return
	}
	labels := sm.getValidatorLabel(nodeAddress)
	// add to accumulated count
	sm.ChallengeCount.With(labels...).Add(1)
	// add to individual count
	nnc.ChallengeCount.With(labels...).Add(1)
	// update nnc
	sm.NonNativeChains[networkID] = nnc
}

func (sm *ServiceMetrics) AddErrorFor(networkID string, nodeAddress *sdk.Address) {
	sm.l.Lock()
	defer sm.l.Unlock()
	// attempt to locate nn chain
	nnc, ok := sm.NonNativeChains[networkID]
	if !ok {
		sm.tmLogger.Error("unable to find corresponding networkID in service metrics: ", networkID)
		sm.NonNativeChains[networkID] = NewServiceMetricsFor(networkID)
		return
	}
	labels := sm.getValidatorLabel(nodeAddress)
	sm.ErrCount.With(labels...).Add(1)
	// add to individual count
	nnc.ErrCount.With(labels...).Add(1)
	// update nnc
	sm.NonNativeChains[networkID] = nnc
}

func (sm *ServiceMetrics) AddRelayTimingFor(networkID string, relayTime float64, nodeAddress *sdk.Address) {
	sm.l.Lock()
	defer sm.l.Unlock()
	// attempt to locate nn chain
	nnc, ok := sm.NonNativeChains[networkID]
	if !ok {
		sm.tmLogger.Error("unable to find corresponding networkID in service metrics: ", networkID)
		sm.NonNativeChains[networkID] = NewServiceMetricsFor(networkID)
		return
	}
	// add to accumulated hist
	labels := sm.getValidatorLabel(nodeAddress)
	sm.AverageRelayTime.With(labels...).Observe(relayTime)
	// add to individual hist
	nnc.AverageRelayTime.With(labels...).Observe(relayTime)
	// update nnc
	sm.NonNativeChains[networkID] = nnc
}

func (sm *ServiceMetrics) AddSessionFor(networkID string, nodeAddress *sdk.Address) {
	sm.l.Lock()
	defer sm.l.Unlock()
	// attempt to locate nn chain
	nnc, ok := sm.NonNativeChains[networkID]
	if !ok {
		sm.tmLogger.Error("unable to find corresponding networkID in service metrics: ", networkID)
		sm.NonNativeChains[networkID] = NewServiceMetricsFor(networkID)

		return
	}

	if nodeAddress == nil {
		// this implies that user is not running in lean viper
		node := GetViperNode()
		if node == nil {
			sm.tmLogger.Error("unable to load privateKey", networkID)
			return
		}
		addr := sdk.GetAddress(node.PrivateKey.PublicKey())
		nodeAddress = &addr
	}
	labels := sm.getValidatorLabel(nodeAddress)

	// add to accumulated count
	sm.TotalSessions.With(labels...).Add(1)
	// add to individual count
	nnc.TotalSessions.With(labels...).Add(1)

	// update nnc
	sm.NonNativeChains[networkID] = nnc
}

func (sm *ServiceMetrics) AdduviprEarnedFor(networkID string, uviprEarned float64, nodeAddress *sdk.Address) {
	sm.l.Lock()
	defer sm.l.Unlock()
	// attempt to locate nn chain
	nnc, ok := sm.NonNativeChains[networkID]
	if !ok {
		sm.tmLogger.Error("unable to find corresponding networkID in service metrics: ", networkID)
		sm.NonNativeChains[networkID] = NewServiceMetricsFor(networkID)
		return
	}
	labels := sm.getValidatorLabel(nodeAddress)
	// add to accumulated count
	sm.uviprEarned.With(labels...).Add(uviprEarned)
	// add to individual count
	nnc.uviprEarned.With(labels...).Add(uviprEarned)
	// update nnc
	sm.NonNativeChains[networkID] = nnc
}
func KeyForServiceMetrics() []byte {
	return []byte(ServiceMetricsKey)
}

func NewServiceMetrics(hostedBlockchains *HostedBlockchains, hostedGeozone *HostedGeoZones, logger log.Logger) *ServiceMetrics {
	serviceMetrics := ServiceMetrics{
		ServiceMetric:   NewServiceMetricsFor("all"),
		NonNativeChains: make(map[string]ServiceMetric),
	}
	if hostedBlockchains != nil {
		for _, hb := range hostedBlockchains.M {
			serviceMetrics.NonNativeChains[hb.ID] = NewServiceMetricsFor(hb.ID)
		}
	}
	// add the logger
	serviceMetrics.tmLogger = logger
	// return the metrics
	return &serviceMetrics
}

type ServiceMetric struct {
	RelayCount       metrics.Counter   `json:"relay_count"`
	ChallengeCount   metrics.Counter   `json:"challenge_count"`
	ErrCount         metrics.Counter   `json:"err_count"`
	AverageRelayTime metrics.Histogram `json:"avg_relay_time"`
	AverageClaimTime metrics.Histogram `json:"avg_claim_time"`
	AverageProofTime metrics.Histogram `json:"avg_proof_time"`
	TotalSessions    metrics.Counter   `json:"total_sessions"`
	uviprEarned      metrics.Counter   `json:"uvipr_earned"`
}

func NewServiceMetricsFor(networkID string) ServiceMetric {
	//labels := make([]string, 1)
	// relay counter metric
	relayCounter := prometheus.NewCounterFrom(stdPrometheus.CounterOpts{
		Namespace: ModuleName,
		Subsystem: ServiceMetricsNamespace,
		Name:      RelayCountName + networkID,
		Help:      RelayCountHelp + networkID,
	}, nil)
	// challenge counter metric
	challengeCounter := prometheus.NewCounterFrom(stdPrometheus.CounterOpts{
		Namespace: ModuleName,
		Subsystem: ServiceMetricsNamespace,
		Name:      ChallengeCountName + networkID,
		Help:      ChallengeCountHelp + networkID,
	}, nil)
	// err counter metric
	errCounter := prometheus.NewCounterFrom(stdPrometheus.CounterOpts{
		Namespace: ModuleName,
		Subsystem: ServiceMetricsNamespace,
		Name:      ErrCountName + networkID,
		Help:      ErrCountHelp + networkID,
	}, nil)
	// Avg relay time histogram metric
	avgRelayTime := prometheus.NewHistogramFrom(stdPrometheus.HistogramOpts{
		Namespace:   ModuleName,
		Subsystem:   ServiceMetricsNamespace,
		Name:        AvgRelayHistName + networkID,
		Help:        AvgrelayHistHelp + networkID,
		ConstLabels: nil,
		Buckets:     stdPrometheus.LinearBuckets(1, 20, 20),
	}, nil)
	// session counter metric
	totalSessions := prometheus.NewCounterFrom(stdPrometheus.CounterOpts{
		Namespace: ModuleName,
		Subsystem: ServiceMetricsNamespace,
		Name:      SessionsCountName + networkID,
		Help:      SessionsCountHelp + networkID,
	}, nil)
	// tokens earned metric
	uviprEarned := prometheus.NewCounterFrom(stdPrometheus.CounterOpts{
		Namespace: ModuleName,
		Subsystem: ServiceMetricsNamespace,
		Name:      uviprCountName + networkID,
		Help:      uviprCountHelp + networkID,
	}, nil)
	return ServiceMetric{
		RelayCount:       relayCounter,
		ChallengeCount:   challengeCounter,
		ErrCount:         errCounter,
		AverageRelayTime: avgRelayTime,
		TotalSessions:    totalSessions,
		uviprEarned:      uviprEarned,
	}
}

func (sm *ServiceMetrics) getValidatorLabel(nodeAddress *sdk.Address) []string {
	return []string{"validator_address", nodeAddress.String()}
}

func (sm *ServiceMetrics) AddProofTiming(networkID string, time float64, nodeAddress *sdk.Address) {

	sm.l.Lock()
	defer sm.l.Unlock()
	// attempt to locate nn chain
	nnc, ok := sm.NonNativeChains[networkID]
	if !ok {
		sm.tmLogger.Error("unable to find corresponding networkID in service metrics: ", networkID)
		sm.NonNativeChains[networkID] = NewServiceMetricsFor(networkID)
		return
	}
	labels := sm.getValidatorLabel(nodeAddress)

	// add to accumulated hist
	sm.AverageProofTime.With(labels...).Observe(time)
	// add to individual hist
	nnc.AverageProofTime.With(labels...).Observe(time)
	// update nnc
	sm.NonNativeChains[networkID] = nnc
}

func (sm *ServiceMetrics) AddClaimTiming(networkID string, time float64, nodeAddress *sdk.Address) {

	sm.l.Lock()
	defer sm.l.Unlock()
	// attempt to locate nn chain
	nnc, ok := sm.NonNativeChains[networkID]
	if !ok {
		sm.tmLogger.Error("unable to find corresponding networkID in service metrics: ", networkID)
		sm.NonNativeChains[networkID] = NewServiceMetricsFor(networkID)
		return
	}
	labels := sm.getValidatorLabel(nodeAddress)
	// add to accumulated hist
	sm.AverageClaimTime.With(labels...).Observe(time)
	// add to individual hist
	nnc.AverageClaimTime.With(labels...).Observe(time)
	// update nnc
	sm.NonNativeChains[networkID] = nnc
}
