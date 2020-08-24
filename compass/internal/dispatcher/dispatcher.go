package dispatcher

import (
	"compass/internal/configuration"
	"compass/internal/metric"
	"compass/internal/metricsgroup"
	"compass/internal/util"
	"fmt"
	"sync"
	"time"
)

type UseCases interface {
	Start() error
}

type Dispatcher struct {
	metric metric.UseCases
	mux           sync.Mutex
}

func NewDispatcher(metric metric.UseCases) UseCases {
	return &Dispatcher{metric, sync.Mutex{}}
}

func (dispatcher *Dispatcher) dispatch() {
	metricExecutions, err := dispatcher.metric.FindAllActivesMetricExecutions()
	if err != nil {
		util.Panic("Cannot find active metric executions", "Dispatch", err, nil)
	}

	for _, execution := range metricExecutions {
		go dispatcher.getMetricResult(execution)
	}

	fmt.Printf("after 5 seconds... %s", time.Now().String())
}

func compareResultWithMetricTreshhold(result float64, threshold float64, condition string) bool {
	switch condition {
	case metricsgroup.EQUAL.String():
		return result == threshold
	case metricsgroup.GREATER_THEN.String():
		return result > threshold
	case metricsgroup.LOWER_THEN.String():
		return result < threshold
	default:
		return false
	}
}

func (dispatcher *Dispatcher) getMetricResult(execution metric.MetricExecution) {
	currentMetric, err := dispatcher.metric.FindMetricById(execution.MetricID.String())
	if err != nil {
		return
	}

	metricResult, err := dispatcher.metric.ResultQuery(currentMetric)
	if err != nil {
		util.Error(util.ResultByGroupMetricError, "getMetricResult", err, currentMetric)
		return
	}

	if compareResultWithMetricTreshhold(metricResult, currentMetric.Threshold, currentMetric.Condition) {
		dispatcher.mux.Lock()
		execution.Status = metric.MetricReached
		dispatcher.metric.SaveMetricExecution(execution)
		dispatcher.mux.Unlock()
	}

	dispatcher.mux.Lock()
	execution.LastValue = metricResult
	dispatcher.metric.SaveMetricExecution(execution)
	dispatcher.mux.Unlock()
}

func (dispatcher *Dispatcher) getInterval() (time.Duration, error) {
	return time.ParseDuration(configuration.GetConfiguration("DISPATCHER_INTERVAL"))
}

func (dispatcher *Dispatcher) Start() error {
	interval, err := dispatcher.getInterval()
	if err != nil {
		return err
	}

	for {
		time.Sleep(interval * time.Second)
		dispatcher.dispatch()
	}
}
