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

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/ossrs/go-oryx-lib/errors"
	ohttp "github.com/ossrs/go-oryx-lib/http"
	"github.com/ossrs/go-oryx-lib/logger"
)

// StreamConfig represents the configuration for a stream
type StreamConfig struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`        // rtmp, srt, hls, webrtc
	Direction   string    `json:"direction"`   // input, output
	URL         string    `json:"url"`
	Port        int       `json:"port"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Status      string    `json:"status"`      // active, inactive, error, connecting
	Connected   bool      `json:"connected"`
	LastError   string    `json:"lastError,omitempty"`
}

// StreamManager manages all streams (input and output)
type StreamManager struct {
	mu      sync.RWMutex
	streams map[string]*StreamConfig
	rdb     *redis.Client
}

var streamManager *StreamManager

func NewStreamManager() *StreamManager {
	if streamManager == nil {
		streamManager = &StreamManager{
			streams: make(map[string]*StreamConfig),
			rdb:     rdb,
		}
	}
	return streamManager
}

func (v *StreamManager) Handle(ctx context.Context, handler *http.ServeMux) error {
	ep := "/terraform/v1/streams/inputs"
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

			switch r.Method {
			case "GET":
				streams := v.GetStreamsByDirection("input")
				ohttp.WriteData(ctx, w, r, streams)
				logger.Tf(ctx, "stream inputs query ok, count=%v, token=%vB", len(streams), len(token))
			case "POST":
				var config StreamConfig
				if err := ParseBody(ctx, r.Body, &config); err != nil {
					return errors.Wrapf(err, "parse config")
				}
				config.Direction = "input"
				if err := v.CreateStream(ctx, &config); err != nil {
					return errors.Wrapf(err, "create stream")
				}
				ohttp.WriteData(ctx, w, r, config)
				logger.Tf(ctx, "stream input created: %v", config)
			default:
				return errors.Errorf("method %v not allowed", r.Method)
			}
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	ep = "/terraform/v1/streams/outputs"
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

			switch r.Method {
			case "GET":
				streams := v.GetStreamsByDirection("output")
				ohttp.WriteData(ctx, w, r, streams)
				logger.Tf(ctx, "stream outputs query ok, count=%v, token=%vB", len(streams), len(token))
			case "POST":
				var config StreamConfig
				if err := ParseBody(ctx, r.Body, &config); err != nil {
					return errors.Wrapf(err, "parse config")
				}
				config.Direction = "output"
				if err := v.CreateStream(ctx, &config); err != nil {
					return errors.Wrapf(err, "create stream")
				}
				ohttp.WriteData(ctx, w, r, config)
				logger.Tf(ctx, "stream output created: %v", config)
			default:
				return errors.Errorf("method %v not allowed", r.Method)
			}
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	ep = "/terraform/v1/streams/all"
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

			streams := v.GetAllStreams()
			ohttp.WriteData(ctx, w, r, streams)
			logger.Tf(ctx, "all streams query ok, count=%v, token=%vB", len(streams), len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	ep = "/terraform/v1/streams/{direction}/{id}"
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

			// Extract direction and id from URL path
			path := r.URL.Path
			var direction, id string
			if _, err := fmt.Sscanf(path, "/terraform/v1/streams/%s/%s", &direction, &id); err != nil {
				return errors.Wrapf(err, "parse path")
			}

			switch r.Method {
			case "PUT":
				var config StreamConfig
				if err := ParseBody(ctx, r.Body, &config); err != nil {
					return errors.Wrapf(err, "parse config")
				}
				config.ID = id
				config.Direction = direction
				if err := v.UpdateStream(ctx, &config); err != nil {
					return errors.Wrapf(err, "update stream")
				}
				ohttp.WriteData(ctx, w, r, config)
				logger.Tf(ctx, "stream updated: %v", config)
			case "DELETE":
				if err := v.DeleteStream(ctx, id); err != nil {
					return errors.Wrapf(err, "delete stream")
				}
				ohttp.WriteData(ctx, w, r, map[string]string{"message": "stream deleted"})
				logger.Tf(ctx, "stream deleted: %v", id)
			default:
				return errors.Errorf("method %v not allowed", r.Method)
			}
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	return nil
}

func (v *StreamManager) CreateStream(ctx context.Context, config *StreamConfig) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	config.Status = "inactive"
	config.Connected = false

	// Set default values
	if config.Type == "" {
		config.Type = "rtmp"
	}
	if config.Port <= 0 {
		switch config.Type {
		case "rtmp":
			config.Port = 1935
		case "srt":
			config.Port = 10080
		case "hls":
			config.Port = 8080
		case "webrtc":
			config.Port = 8000
		}
	}

	// Validate configuration
	if config.Name == "" {
		return errors.Errorf("stream name is required")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return errors.Errorf("invalid port: %v", config.Port)
	}

	// Save to Redis
	key := fmt.Sprintf("stream:%s", config.ID)
	if b, err := json.Marshal(config); err != nil {
		return errors.Wrapf(err, "marshal config")
	} else if err := v.rdb.Set(ctx, key, b, 0).Err(); err != nil {
		return errors.Wrapf(err, "save to redis")
	}

	v.streams[config.ID] = config
	logger.Tf(ctx, "stream created: %v", config)
	return nil
}

func (v *StreamManager) UpdateStream(ctx context.Context, config *StreamConfig) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	config.UpdatedAt = time.Now()

	// Save to Redis
	key := fmt.Sprintf("stream:%s", config.ID)
	if b, err := json.Marshal(config); err != nil {
		return errors.Wrapf(err, "marshal config")
	} else if err := v.rdb.Set(ctx, key, b, 0).Err(); err != nil {
		return errors.Wrapf(err, "save to redis")
	}

	v.streams[config.ID] = config
	logger.Tf(ctx, "stream updated: %v", config)
	return nil
}

func (v *StreamManager) DeleteStream(ctx context.Context, id string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Remove from Redis
	key := fmt.Sprintf("stream:%s", id)
	if err := v.rdb.Del(ctx, key).Err(); err != nil {
		return errors.Wrapf(err, "delete from redis")
	}

	delete(v.streams, id)
	logger.Tf(ctx, "stream deleted: %v", id)
	return nil
}

func (v *StreamManager) GetStreamsByDirection(direction string) []*StreamConfig {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var streams []*StreamConfig
	for _, stream := range v.streams {
		if stream.Direction == direction {
			streams = append(streams, stream)
		}
	}
	return streams
}

func (v *StreamManager) GetAllStreams() []*StreamConfig {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var streams []*StreamConfig
	for _, stream := range v.streams {
		streams = append(streams, stream)
	}
	return streams
}

func (v *StreamManager) GetStream(id string) *StreamConfig {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.streams[id]
}

func (v *StreamManager) LoadStreamsFromRedis(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	keys, err := v.rdb.Keys(ctx, "stream:*").Result()
	if err != nil {
		return errors.Wrapf(err, "get stream keys")
	}

	for _, key := range keys {
		val, err := v.rdb.Get(ctx, key).Result()
		if err != nil {
			logger.Wf(ctx, "failed to get stream %v: %v", key, err)
			continue
		}

		var config StreamConfig
		if err := json.Unmarshal([]byte(val), &config); err != nil {
			logger.Wf(ctx, "failed to unmarshal stream %v: %v", key, err)
			continue
		}

		v.streams[config.ID] = &config
	}

	logger.Tf(ctx, "loaded %v streams from redis", len(v.streams))
	return nil
}
