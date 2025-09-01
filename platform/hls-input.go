// Copyright (c) 2022-2024 Winlin
//
// SPDX-License-Identifier: MIT
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/ossrs/go-oryx-lib/errors"
	ohttp "github.com/ossrs/go-oryx-lib/http"
	"github.com/ossrs/go-oryx-lib/logger"
)

// HLSInputConfig represents the configuration for HLS input
type HLSInputConfig struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Status      string    `json:"status"` // active, inactive, error
	LastError   string    `json:"lastError,omitempty"`
	StreamCount int       `json:"streamCount"`
}

// HLSInputManager manages HLS input streams
type HLSInputManager struct {
	mu     sync.RWMutex
	inputs map[string]*HLSInputConfig
	rdb    *redis.Client
}

var hlsInputManager *HLSInputManager

func NewHLSInputManager() *HLSInputManager {
	if hlsInputManager == nil {
		hlsInputManager = &HLSInputManager{
			inputs: make(map[string]*HLSInputConfig),
			rdb:    rdb,
		}
	}
	return hlsInputManager
}

func (v *HLSInputManager) Handle(ctx context.Context, handler *http.ServeMux) error {
	// Query HLS inputs
	ep := "/terraform/v1/hls/input/query"
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
			logger.Tf(ctx, "hls input query ok, count=%v, token=%vB", len(inputs), len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	// Create HLS input
	ep = "/terraform/v1/hls/input/create"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			var token string
			var config HLSInputConfig
			if err := ParseBody(ctx, r.Body, &struct {
				Token *string `json:"token"`
				*HLSInputConfig
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

			if err := v.CreateInput(ctx, &config); err != nil {
				return errors.Wrapf(err, "create hls input")
			}

			ohttp.WriteData(ctx, w, r, config)
			logger.Tf(ctx, "hls input create ok, %v, token=%vB", config, len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	// Update HLS input
	ep = "/terraform/v1/hls/input/update"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			var token string
			var config HLSInputConfig
			if err := ParseBody(ctx, r.Body, &struct {
				Token *string `json:"token"`
				*HLSInputConfig
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

			if err := v.UpdateInput(ctx, &config); err != nil {
				return errors.Wrapf(err, "update hls input")
			}

			ohttp.WriteData(ctx, w, r, config)
			logger.Tf(ctx, "hls input update ok, %v, token=%vB", config, len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	// Delete HLS input
	ep = "/terraform/v1/hls/input/delete"
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
				return errors.Wrapf(err, "delete hls input")
			}

			ohttp.WriteData(ctx, w, r, map[string]string{"message": "HLS input deleted successfully"})
			logger.Tf(ctx, "hls input delete ok, id=%v, token=%vB", inputID, len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	return nil
}

func (v *HLSInputManager) CreateInput(ctx context.Context, config *HLSInputConfig) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	config.Status = "inactive"

	// Validate HLS URL
	if !strings.HasPrefix(config.URL, "http://") && !strings.HasPrefix(config.URL, "https://") {
		return errors.Errorf("invalid HLS URL: %v", config.URL)
	}

	// Save to Redis
	key := fmt.Sprintf("hls_input:%s", config.ID)
	if b, err := json.Marshal(config); err != nil {
		return errors.Wrapf(err, "marshal config")
	} else if err := v.rdb.Set(ctx, key, b, 0).Err(); err != nil {
		return errors.Wrapf(err, "save to redis")
	}

	v.inputs[config.ID] = config
	logger.Tf(ctx, "hls input created: %v", config)
	return nil
}

func (v *HLSInputManager) UpdateInput(ctx context.Context, config *HLSInputConfig) error {
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

	// Save to Redis
	key := fmt.Sprintf("hls_input:%s", config.ID)
	if b, err := json.Marshal(config); err != nil {
		return errors.Wrapf(err, "marshal config")
	} else if err := v.rdb.Set(ctx, key, b, 0).Err(); err != nil {
		return errors.Wrapf(err, "save to redis")
	}

	v.inputs[config.ID] = config
	logger.Tf(ctx, "hls input updated: %v", config)
	return nil
}

func (v *HLSInputManager) DeleteInput(ctx context.Context, inputID string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if inputID == "" {
		return errors.Errorf("input ID is required")
	}

	// Remove from Redis
	key := fmt.Sprintf("hls_input:%s", inputID)
	if err := v.rdb.Del(ctx, key).Err(); err != nil {
		return errors.Wrapf(err, "delete from redis")
	}

	delete(v.inputs, inputID)
	logger.Tf(ctx, "hls input deleted: %v", inputID)
	return nil
}

func (v *HLSInputManager) GetAllInputs() []*HLSInputConfig {
	v.mu.RLock()
	defer v.mu.RUnlock()

	inputs := make([]*HLSInputConfig, 0, len(v.inputs))
	for _, input := range v.inputs {
		inputs = append(inputs, input)
	}
	return inputs
}

func (v *HLSInputManager) GetInput(inputID string) *HLSInputConfig {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.inputs[inputID]
}

func (v *HLSInputManager) StartInput(ctx context.Context, inputID string) error {
	input := v.GetInput(inputID)
	if input == nil {
		return errors.Errorf("input not found: %v", inputID)
	}

	// Start HLS input processing
	go v.processHLSInput(ctx, input)
	return nil
}

func (v *HLSInputManager) processHLSInput(ctx context.Context, input *HLSInputConfig) {
	logger.Tf(ctx, "start processing HLS input: %v", input.URL)

	// Update status to active
	input.Status = "active"
	input.UpdatedAt = time.Now()
	v.UpdateInput(ctx, input)

	// TODO: Implement HLS stream processing logic
	// This would involve:
	// 1. Fetching the HLS playlist
	// 2. Parsing segments
	// 3. Forwarding to SRS without re-encoding
	// 4. Monitoring stream health

	logger.Tf(ctx, "HLS input processing started: %v", input.URL)
}
