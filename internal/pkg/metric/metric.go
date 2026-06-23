package metric

import (
	"cake/env"
	"cake/internal/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	reg = prometheus.NewRegistry()

	// RPC总请求量
	RpcReqTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rpc_request_total",
			Help: "gensvc total request count",
		},
		[]string{"service", "cmd", "call_type"},
	)

	// RPC错误次数
	RpcErrTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rpc_error_total",
			Help: "gensvc error count",
		},
		[]string{"service", "cmd", "err_type"},
	)

	// RPC耗时直方图
	RpcDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rpc_handle_duration_ms",
			Help:    "gensvc handle cost ms",
			Buckets: []float64{1, 5, 10, 20, 50, 100, 200, 500},
		},
		[]string{"service", "cmd"},
	)

	// 慢调用次数
	RpcSlowTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rpc_slow_call_total",
			Help: "gensvc slow call count",
		},
		[]string{"service", "cmd"},
	)

	SendQueueLen = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rpc_send_queue_length",
			Help: "service send channel current length",
		},
		[]string{"service"},
	)
)

func Init() {
	reg.MustRegister(RpcReqTotal)
	reg.MustRegister(RpcErrTotal)
	reg.MustRegister(RpcDuration)
	reg.MustRegister(RpcSlowTotal)
	reg.MustRegister(SendQueueLen)
	addr := env.GetString("metric.addr")
	StartHTTP(addr)
}

// StartHTTP 启动监控接口
func StartHTTP(addr string) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Errorf("prometheus metrics start failed: %v", err)
		}
	}()
}
