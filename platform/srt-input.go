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

// SRTInputConfig represents the configuration for SRT input
type SRTInputConfig struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Port            int       `json:"port"`            // Port for SRT with StreamID (default: 10080)
	PortNoStreamId1 int       `json:"portNoStreamId1"` // Port for SRT without StreamID, stream 1 (default: 10081)
	PortNoStreamId2 int       `json:"portNoStreamId2"` // Port for SRT without StreamID, stream 2 (default: 10082)
	Enabled         bool      `json:"enabled"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	Status          string    `json:"status"` // active, inactive, error
	LastError       string    `json:"lastError,omitempty"`
	StreamCount     int       `json:"streamCount"`
	MaxStreams      int       `json:"maxStreams"` // Maximum 2 streams
}

// SRTStream represents an individual SRT stream
type SRTStream struct {
	ID        string    `json:"id"`
	InputID   string    `json:"inputId"`
	StreamID  string    `json:"streamId"`
	Status    string    `json:"status"` // connected, disconnected
	Connected time.Time `json:"connected"`
	IP        string    `json:"ip"`
	Port      int       `json:"port"`
}

// SRTInputManager manages SRT input streams
type SRTInputManager struct {
	mu      sync.RWMutex
	inputs  map[string]*SRTInputConfig
	streams map[string]*SRTStream
	rdb     *redis.Client
}

var srtInputManager *SRTInputManager

func NewSRTInputManager() *SRTInputManager {
	if srtInputManager == nil {
		srtInputManager = &SRTInputManager{
			inputs:  make(map[string]*SRTInputConfig),
			streams: make(map[string]*SRTStream),
			rdb:     rdb,
		}
	}
	return srtInputManager
}

func (v *SRTInputManager) Handle(ctx context.Context, handler *http.ServeMux) error {
	// Query SRT inputs
	ep := "/terraform/v1/srt/input/query"
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

			inputs := v.GetAllInputs()
			ohttp.WriteData(ctx, w, r, inputs)
			logger.Tf(ctx, "srt input query ok, count=%v, token=%vB", len(inputs), len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	// Create SRT input
	ep = "/terraform/v1/srt/input/create"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			var token string
			var config SRTInputConfig
			if err := ParseBody(ctx, r.Body, &struct {
				Token *string `json:"token"`
				*SRTInputConfig
			}{
				Token:          &token,
				SRTInputConfig: &config,
			}); err != nil {
				return errors.Wrapf(err, "parse body")
			}

			apiSecret := envApiSecret()
			if err := Authenticate(ctx, apiSecret, token, r.Header); err != nil {
				return errors.Wrapf(err, "authenticate")
			}

			if err := v.CreateInput(ctx, &config); err != nil {
				return errors.Wrapf(err, "create srt input")
			}

			ohttp.WriteData(ctx, w, r, config)
			logger.Tf(ctx, "srt input create ok, %v, token=%vB", config, len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	// Update SRT input
	ep = "/terraform/v1/srt/input/update"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			var token string
			var config SRTInputConfig
			if err := ParseBody(ctx, r.Body, &struct {
				Token *string `json:"token"`
				*SRTInputConfig
			}{
				Token:          &token,
				SRTInputConfig: &config,
			}); err != nil {
				return errors.Wrapf(err, "parse body")
			}

			apiSecret := envApiSecret()
			if err := Authenticate(ctx, apiSecret, token, r.Header); err != nil {
				return errors.Wrapf(err, "authenticate")
			}

			if err := v.UpdateInput(ctx, &config); err != nil {
				return errors.Wrapf(err, "update srt input")
			}

			ohttp.WriteData(ctx, w, r, config)
			logger.Tf(ctx, "srt input update ok, %v, token=%vB", config, len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	// Delete SRT input
	ep = "/terraform/v1/srt/input/delete"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			var token string
			var inputID string
			if err := ParseBody(ctx, r.Body, &struct {
				Token *string `json:"token"`
				ID    *string `json:"id"`
			}{
				Token: &token,
				ID:    &inputID,
			}); err != nil {
				return errors.Wrapf(err, "parse body")
			}

			apiSecret := envApiSecret()
			if err := Authenticate(ctx, apiSecret, token, r.Header); err != nil {
				return errors.Wrapf(err, "authenticate")
			}

			if err := v.DeleteInput(ctx, inputID); err != nil {
				return errors.Wrapf(err, "delete srt input")
			}

			ohttp.WriteData(ctx, w, r, map[string]string{"message": "SRT input deleted successfully"})
			logger.Tf(ctx, "srt input delete ok, id=%v, token=%vB", inputID, len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	// Query SRT streams
	ep = "/terraform/v1/srt/stream/query"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			var token string
			var inputID string
			if err := ParseBody(ctx, r.Body, &struct {
				Token *string `json:"token"`
				ID    *string `json:"inputId"`
			}{
				Token: &token,
				ID:    &inputID,
			}); err != nil {
				return errors.Wrapf(err, "parse body")
			}

			apiSecret := envApiSecret()
			if err := Authenticate(ctx, apiSecret, token, r.Header); err != nil {
				return errors.Wrapf(err, "authenticate")
			}

			streams := v.GetStreamsByInput(inputID)
			ohttp.WriteData(ctx, w, r, streams)
			logger.Tf(ctx, "srt stream query ok, inputId=%v, count=%v, token=%vB", inputID, len(streams), len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	return nil
}

func (v *SRTInputManager) CreateInput(ctx context.Context, config *SRTInputConfig) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	config.Status = "inactive"
	config.MaxStreams = 2 // Fixed to 2 streams

	// Set default ports if not specified
	if config.Port <= 0 {
		config.Port = 10080 // Default port for SRT with StreamID
	}
	if config.PortNoStreamId1 <= 0 {
		config.PortNoStreamId1 = 10081 // Default port for SRT without StreamID, stream 1
	}
	if config.PortNoStreamId2 <= 0 {
		config.PortNoStreamId2 = 10082 // Default port for SRT without StreamID, stream 2
	}

	// Validate ports
	if config.Port <= 0 || config.Port > 65535 {
		return errors.Errorf("invalid port: %v", config.Port)
	}
	if config.PortNoStreamId1 <= 0 || config.PortNoStreamId1 > 65535 {
		return errors.Errorf("invalid portNoStreamId1: %v", config.PortNoStreamId1)
	}
	if config.PortNoStreamId2 <= 0 || config.PortNoStreamId2 > 65535 {
		return errors.Errorf("invalid portNoStreamId2: %v", config.PortNoStreamId2)
	}

	// Save to Redis
	key := fmt.Sprintf("srt_input:%s", config.ID)
	if b, err := json.Marshal(config); err != nil {
		return errors.Wrapf(err, "marshal config")
	} else if err := v.rdb.Set(ctx, key, b, 0).Err(); err != nil {
		return errors.Wrapf(err, "save to redis")
	}

	v.inputs[config.ID] = config
	logger.Tf(ctx, "srt input created: %v", config)
	return nil
}

func (v *SRTInputManager) UpdateInput(ctx context.Context, config *SRTInputConfig) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if config.ID == "" {
		return errors.Errorf("input ID is required")
	}

	existing, exists := v.inputs[config.ID]
	if !exists {
		return errors.Errorf("input not found: %v", config.ID)
	}

	config.UpdatedAt = time.Now()
	config.CreatedAt = existing.CreatedAt
	config.MaxStreams = 2 // Always enforce 2 streams

	// Save to Redis
	key := fmt.Sprintf("srt_input:%s", config.ID)
	if b, err := json.Marshal(config); err != nil {
		return errors.Wrapf(err, "marshal config")
	} else if err := v.rdb.Set(ctx, key, b, 0).Err(); err != nil {
		return errors.Wrapf(err, "save to redis")
	}

	v.inputs[config.ID] = config
	logger.Tf(ctx, "srt input updated: %v", config)
	return nil
}

func (v *SRTInputManager) DeleteInput(ctx context.Context, inputID string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if inputID == "" {
		return errors.Errorf("input ID is required")
	}

	// Remove from Redis
	key := fmt.Sprintf("srt_input:%s", inputID)
	if err := v.rdb.Del(ctx, key).Err(); err != nil {
		return errors.Wrapf(err, "delete from redis")
	}

	// Remove all associated streams
	for streamID, stream := range v.streams {
		if stream.InputID == inputID {
			delete(v.streams, streamID)
		}
	}

	delete(v.inputs, inputID)
	logger.Tf(ctx, "srt input deleted: %v", inputID)
	return nil
}

func (v *SRTInputManager) GetAllInputs() []*SRTInputConfig {
	v.mu.RLock()
	defer v.mu.RUnlock()

	inputs := make([]*SRTInputConfig, 0, len(v.inputs))
	for _, input := range v.inputs {
		inputs = append(inputs, input)
	}
	return inputs
}

func (v *SRTInputManager) GetInput(inputID string) *SRTInputConfig {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.inputs[inputID]
}

func (v *SRTInputManager) GetStreamsByInput(inputID string) []*SRTStream {
	v.mu.RLock()
	defer v.mu.RUnlock()

	streams := make([]*SRTStream, 0)
	for _, stream := range v.streams {
		if stream.InputID == inputID {
			streams = append(streams, stream)
		}
	}
	return streams
}

func (v *SRTInputManager) AddStream(inputID, streamID, ip string, port int) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	input := v.inputs[inputID]
	if input == nil {
		return errors.Errorf("input not found: %v", inputID)
	}

	// Check if we already have 2 streams
	streamCount := 0
	for _, stream := range v.streams {
		if stream.InputID == inputID {
			streamCount++
		}
	}

	if streamCount >= input.MaxStreams {
		return errors.Errorf("maximum streams reached for input %v", inputID)
	}

	stream := &SRTStream{
		ID:        uuid.New().String(),
		InputID:   inputID,
		StreamID:  streamID,
		Status:    "connected",
		Connected: time.Now(),
		IP:        ip,
		Port:      port,
	}

	v.streams[stream.ID] = stream
	input.StreamCount = streamCount + 1

	logger.Tf(context.Background(), "srt stream added: %v", stream)
	return nil
}

func (v *SRTInputManager) RemoveStream(streamID string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	stream, exists := v.streams[streamID]
	if !exists {
		return errors.Errorf("stream not found: %v", streamID)
	}

	input := v.inputs[stream.InputID]
	if input != nil {
		input.StreamCount--
	}

	delete(v.streams, streamID)
	logger.Tf(context.Background(), "srt stream removed: %v", streamID)
	return nil
}

func (v *SRTInputManager) StartInput(ctx context.Context, inputID string) error {
	input := v.GetInput(inputID)
	if input == nil {
		return errors.Errorf("input not found: %v", inputID)
	}

	// Start SRT input processing
	go v.processSRTInput(ctx, input)
	return nil
}

func (v *SRTInputManager) processSRTInput(ctx context.Context, input *SRTInputConfig) {
	logger.Tf(ctx, "start processing SRT input on ports: %v (StreamID), %v (No StreamID 1), %v (No StreamID 2)",
		input.Port, input.PortNoStreamId1, input.PortNoStreamId2)

	// Update status to active
	input.Status = "active"
	input.UpdatedAt = time.Now()
	v.UpdateInput(ctx, input)

	// TODO: Implement SRT stream processing logic
	// This would involve:
	// 1. Listening on the specified ports:
	//    - Port for SRT with StreamID (default: 10080)
	//    - Port for SRT without StreamID, stream 1 (default: 10081)
	//    - Port for SRT without StreamID, stream 2 (default: 10082)
	// 2. Accepting SRT connections (max 2 per port type)
	// 3. Forwarding to SRS without re-encoding
	// 4. Monitoring stream health and connection status

	logger.Tf(ctx, "SRT input processing started on ports: %v (StreamID), %v (No StreamID 1), %v (No StreamID 2)",
		input.Port, input.PortNoStreamId1, input.PortNoStreamId2)
}
