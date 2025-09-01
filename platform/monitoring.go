// Copyright (c) 2022-2024 Winlin
//
// SPDX-License-Identifier: MIT
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ossrs/go-oryx-lib/errors"
	ohttp "github.com/ossrs/go-oryx-lib/http"
	"github.com/ossrs/go-oryx-lib/logger"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// MonitoringData represents monitoring data points
type MonitoringData struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`        // bandwidth, concurrent_streams
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`        // Mbps, count
	StreamID    string    `json:"streamId,omitempty"`
	InputType   string    `json:"inputType,omitempty"`   // hls, srt, rtmp
	OutputType  string    `json:"outputType,omitempty"`  // hls, srt, rtmp
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	ID              string    `json:"id"`
	Enabled         bool      `json:"enabled"`
	SamplingRate   int       `json:"samplingRate"`   // Seconds between samples
	RetentionDays  int       `json:"retentionDays"`  // Days to keep data
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// MonitoringManager manages monitoring data collection and retrieval
type MonitoringManager struct {
	mu      sync.RWMutex
	config  *MonitoringConfig
	rdb     *redis.Client
	metrics map[string]*MonitoringData
}

var monitoringManager *MonitoringManager

func NewMonitoringManager() *MonitoringManager {
	if monitoringManager == nil {
		monitoringManager = &MonitoringManager{
			config: &MonitoringConfig{
				ID:             uuid.New().String(),
				Enabled:        true,
				SamplingRate:   5,  // 5 seconds
				RetentionDays:  30, // 30 days
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			rdb:     rdb,
			metrics: make(map[string]*MonitoringData),
		}
	}
	return monitoringManager
}

func (v *MonitoringManager) Handle(ctx context.Context, handler *http.ServeMux) error {
	// Query monitoring data
	ep := "/terraform/v1/monitoring/query"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			var token string
			var query struct {
				Type      string    `json:"type"`
				StartTime time.Time `json:"startTime"`
				EndTime   time.Time `json:"endTime"`
				StreamID  string    `json:"streamId,omitempty"`
				Period    string    `json:"period"` // daily, weekly, monthly
			}
			if err := ParseBody(ctx, r.Body, &struct {
				Token *string `json:"token"`
				*struct {
					Type      string    `json:"type"`
					StartTime time.Time `json:"startTime"`
					EndTime   time.Time `json:"endTime"`
					StreamID  string    `json:"streamId,omitempty"`
					Period    string    `json:"period"`
				}
			}{
				Token: &token,
				*struct {
					Type      string    `json:"type"`
					StartTime time.Time `json:"startTime"`
					EndTime   time.Time `json:"endTime"`
					StreamID  string    `json:"streamId,omitempty"`
					Period    string    `json:"period"`
				}: &query,
			}); err != nil {
				return errors.Wrapf(err, "parse body")
			}

			apiSecret := envApiSecret()
			if err := Authenticate(ctx, apiSecret, token, r.Header); err != nil {
				return errors.Wrapf(err, "authenticate")
			}

			data := v.QueryData(ctx, query.Type, query.StartTime, query.EndTime, query.StreamID, query.Period)
			ohttp.WriteData(ctx, w, r, data)
			logger.Tf(ctx, "monitoring query ok, type=%v, period=%v, count=%v, token=%vB", 
				query.Type, query.Period, len(data), len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	// Get monitoring configuration
	ep = "/terraform/v1/monitoring/config/query"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			var token string
			if err := ParseBody(ctx, r.Body, &struct {
				Token *string `json:"token"`
			}{
				Token: &token,
			}); err != nil {
				return errors.Wrapf(err, "parse body")
			}

			apiSecret := envApiSecret()
			if err := Authenticate(ctx, apiSecret, token, r.Header); err != nil {
				return errors.Wrapf(err, "authenticate")
			}

			ohttp.WriteData(ctx, w, r, v.config)
			logger.Tf(ctx, "monitoring config query ok, token=%vB", len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	// Update monitoring configuration
	ep = "/terraform/v1/monitoring/config/update"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			var token string
			var config MonitoringConfig
			if err := ParseBody(ctx, r.Body, &struct {
				Token *string `json:"token"`
				*MonitoringConfig
			}{
				Token:           &token,
				TranscodeConfig: &config,
			}); err != nil {
				return errors.Wrapf(err, "parse body")
			}

			apiSecret := envApiSecret()
			if err := Authenticate(ctx, apiSecret, token, r.Header); err != nil {
				return errors.Wrapf(err, "authenticate")
			}

			if err := v.UpdateConfig(ctx, &config); err != nil {
				return errors.Wrapf(err, "update monitoring config")
			}

			ohttp.WriteData(ctx, w, r, v.config)
			logger.Tf(ctx, "monitoring config update ok, %v, token=%vB", v.config, len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	// Get real-time metrics
	ep = "/terraform/v1/monitoring/realtime"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			var token string
			if err := ParseBody(ctx, r.Body, &struct {
				Token *string `json:"token"`
			}{
				Token: &token,
			}); err != nil {
				return errors.Wrapf(err, "parse body")
			}

			apiSecret := envApiSecret()
			if err := Authenticate(ctx, apiSecret, token, r.Header); err != nil {
				return errors.Wrapf(err, "authenticate")
			}

			metrics := v.GetRealTimeMetrics()
			ohttp.WriteData(ctx, w, r, metrics)
			logger.Tf(ctx, "monitoring realtime ok, count=%v, token=%vB", len(metrics), len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	return nil
}

func (v *MonitoringManager) StartMonitoring(ctx context.Context) {
	if !v.config.Enabled {
		return
	}

	ticker := time.NewTicker(time.Duration(v.config.SamplingRate) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			v.collectMetrics(ctx)
		}
	}
}

func (v *MonitoringManager) collectMetrics(ctx context.Context) {
	// Collect bandwidth metrics
	v.collectBandwidthMetrics(ctx)
	
	// Collect concurrent stream metrics
	v.collectConcurrentStreamMetrics(ctx)
	
	// Clean up old data
	v.cleanupOldData(ctx)
}

func (v *MonitoringManager) collectBandwidthMetrics(ctx context.Context) {
	// TODO: Implement bandwidth collection logic
	// This would involve:
	// 1. Querying SRS for current bandwidth usage
	// 2. Collecting data from network interfaces
	// 3. Aggregating per-stream bandwidth
	// 4. Storing metrics in Redis

	// Example metric
	metric := &MonitoringData{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Type:      "bandwidth",
		Value:     100.5, // Mbps
		Unit:      "Mbps",
		Metadata: map[string]interface{}{
			"total_bandwidth": 100.5,
			"active_streams":  5,
		},
	}

	v.storeMetric(ctx, metric)
}

func (v *MonitoringManager) collectConcurrentStreamMetrics(ctx context.Context) {
	// TODO: Implement concurrent stream collection logic
	// This would involve:
	// 1. Querying SRS for active streams
	// 2. Counting streams by type (HLS, SRT, RTMP)
	// 3. Tracking stream lifecycle events
	// 4. Storing metrics in Redis

	// Example metric
	metric := &MonitoringData{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Type:      "concurrent_streams",
		Value:     5, // Count
		Unit:      "count",
		Metadata: map[string]interface{}{
			"hls_streams":   2,
			"srt_streams":   2,
			"rtmp_streams":  1,
			"total_streams": 5,
		},
	}

	v.storeMetric(ctx, metric)
}

func (v *MonitoringManager) storeMetric(ctx context.Context, metric *MonitoringData) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Store in memory
	v.metrics[metric.ID] = metric

	// Store in Redis
	key := fmt.Sprintf("monitoring:%s:%s", metric.Type, metric.ID)
	if b, err := json.Marshal(metric); err != nil {
		logger.Wf(ctx, "failed to marshal metric: %v", err)
		return
	} else if err := v.rdb.Set(ctx, key, b, time.Duration(v.config.RetentionDays)*24*time.Hour).Err(); err != nil {
		logger.Wf(ctx, "failed to store metric in redis: %v", err)
		return
	}

	// Store timestamp index for querying
	timestampKey := fmt.Sprintf("monitoring:timestamp:%s:%d", metric.Type, metric.Timestamp.Unix())
	if err := v.rdb.Set(ctx, timestampKey, metric.ID, time.Duration(v.config.RetentionDays)*24*time.Hour).Err(); err != nil {
		logger.Wf(ctx, "failed to store timestamp index: %v", err)
	}
}

func (v *MonitoringManager) QueryData(ctx context.Context, dataType, startTime, endTime, streamID, period string) []*MonitoringData {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var results []*MonitoringData

	// Query from Redis based on parameters
	pattern := fmt.Sprintf("monitoring:%s:*", dataType)
	keys, err := v.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Wf(ctx, "failed to query redis keys: %v", err)
		return results
	}

	for _, key := range keys {
		data, err := v.rdb.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var metric MonitoringData
		if err := json.Unmarshal([]byte(data), &metric); err != nil {
			continue
		}

		// Apply filters
		if !startTime.IsZero() && metric.Timestamp.Before(startTime) {
			continue
		}
		if !endTime.IsZero() && metric.Timestamp.After(endTime) {
			continue
		}
		if streamID != "" && metric.StreamID != streamID {
			continue
		}

		results = append(results, &metric)
	}

	// Apply period aggregation if specified
	if period != "" {
		results = v.aggregateByPeriod(results, period)
	}

	return results
}

func (v *MonitoringManager) aggregateByPeriod(data []*MonitoringData, period string) []*MonitoringData {
	if len(data) == 0 {
		return data
	}

	aggregated := make(map[string]*MonitoringData)
	var timeFormat string

	switch period {
	case "daily":
		timeFormat = "2006-01-02"
	case "weekly":
		timeFormat = "2006-W01"
	case "monthly":
		timeFormat = "2006-01"
	default:
		return data
	}

	for _, metric := range data {
		periodKey := metric.Timestamp.Format(timeFormat)
		
		if existing, exists := aggregated[periodKey]; exists {
			existing.Value += metric.Value
			// Update metadata
			if existing.Metadata == nil {
				existing.Metadata = make(map[string]interface{})
			}
			if count, ok := existing.Metadata["sample_count"].(int); ok {
				existing.Metadata["sample_count"] = count + 1
			} else {
				existing.Metadata["sample_count"] = 1
			}
		} else {
			// Create new aggregated metric
			aggregatedMetric := *metric
			aggregatedMetric.ID = uuid.New().String()
			aggregatedMetric.Timestamp = metric.Timestamp
			aggregatedMetric.Metadata = map[string]interface{}{
				"sample_count": 1,
				"period":       period,
			}
			aggregated[periodKey] = &aggregatedMetric
		}
	}

	// Convert map to slice
	results := make([]*MonitoringData, 0, len(aggregated))
	for _, metric := range aggregated {
		results = append(results, metric)
	}

	return results
}

func (v *MonitoringManager) GetRealTimeMetrics() map[string]interface{} {
	v.mu.RLock()
	defer v.mu.RUnlock()

	metrics := make(map[string]interface{})
	
	// Get latest bandwidth metric
	var latestBandwidth *MonitoringData
	for _, metric := range v.metrics {
		if metric.Type == "bandwidth" {
			if latestBandwidth == nil || metric.Timestamp.After(latestBandwidth.Timestamp) {
				latestBandwidth = metric
			}
		}
	}

	// Get latest concurrent streams metric
	var latestStreams *MonitoringData
	for _, metric := range v.metrics {
		if metric.Type == "concurrent_streams" {
			if latestStreams == nil || metric.Timestamp.After(latestStreams.Timestamp) {
				latestStreams = metric
			}
		}
	}

	if latestBandwidth != nil {
		metrics["bandwidth"] = latestBandwidth
	}
	if latestStreams != nil {
		metrics["concurrent_streams"] = latestStreams
	}

	metrics["timestamp"] = time.Now()
	return metrics
}

func (v *MonitoringManager) UpdateConfig(ctx context.Context, config *MonitoringConfig) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	config.UpdatedAt = time.Now()
	config.CreatedAt = v.config.CreatedAt
	v.config = config

	// Save to Redis
	key := "monitoring:config"
	if b, err := json.Marshal(config); err != nil {
		return errors.Wrapf(err, "marshal config")
	} else if err := v.rdb.Set(ctx, key, b, 0).Err(); err != nil {
		return errors.Wrapf(err, "save to redis")
	}

	return nil
}

func (v *MonitoringManager) cleanupOldData(ctx context.Context) {
	// Remove data older than retention period
	retentionDuration := time.Duration(v.config.RetentionDays) * 24 * time.Hour
	cutoffTime := time.Now().Add(-retentionDuration)

	// Clean up memory
	v.mu.Lock()
	for id, metric := range v.metrics {
		if metric.Timestamp.Before(cutoffTime) {
			delete(v.metrics, id)
		}
	}
	v.mu.Unlock()

	// Clean up Redis (this would be done by TTL, but we can also clean up manually)
	// TODO: Implement Redis cleanup logic
} 