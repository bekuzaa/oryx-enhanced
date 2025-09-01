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

// BypassTranscodeConfig represents the configuration for bypass transcoding
type BypassTranscodeConfig struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	InputType  string    `json:"inputType"` // hls, srt, rtmp
	InputURL   string    `json:"inputUrl"`
	OutputType string    `json:"outputType"` // rtmp, hls, srt
	OutputURL  string    `json:"outputUrl"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Status     string    `json:"status"` // active, inactive, error
	LastError  string    `json:"lastError,omitempty"`
	BypassMode string    `json:"bypassMode"` // passthrough, filter
	Filters    []string  `json:"filters"`    // List of filters to apply
}

// BypassTranscodeManager manages bypass transcoding tasks
type BypassTranscodeManager struct {
	mu    sync.RWMutex
	tasks map[string]*BypassTranscodeConfig
	rdb   *redis.Client
}

var bypassTranscodeManager *BypassTranscodeManager

func NewBypassTranscodeManager() *BypassTranscodeManager {
	if bypassTranscodeManager == nil {
		bypassTranscodeManager = &BypassTranscodeManager{
			tasks: make(map[string]*BypassTranscodeConfig),
			rdb:   rdb,
		}
	}
	return bypassTranscodeManager
}

func (v *BypassTranscodeManager) Handle(ctx context.Context, handler *http.ServeMux) error {
	// Query bypass transcode tasks
	ep := "/terraform/v1/bypass/transcode/query"
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

			tasks := v.GetAllTasks()
			ohttp.WriteData(ctx, w, r, tasks)
			logger.Tf(ctx, "bypass transcode query ok, count=%v, token=%vB", len(tasks), len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	// Create bypass transcode task
	ep = "/terraform/v1/bypass/transcode/create"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			var token string
			var config BypassTranscodeConfig
			if err := ParseBody(ctx, r.Body, &struct {
				Token *string `json:"token"`
				*BypassTranscodeConfig
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

			if err := v.CreateTask(ctx, &config); err != nil {
				return errors.Wrapf(err, "create bypass transcode task")
			}

			ohttp.WriteData(ctx, w, r, config)
			logger.Tf(ctx, "bypass transcode create ok, %v, token=%vB", config, len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	// Update bypass transcode task
	ep = "/terraform/v1/bypass/transcode/update"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			var token string
			var config BypassTranscodeConfig
			if err := ParseBody(ctx, r.Body, &struct {
				Token *string `json:"token"`
				*BypassTranscodeConfig
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

			if err := v.UpdateTask(ctx, &config); err != nil {
				return errors.Wrapf(err, "update bypass transcode task")
			}

			ohttp.WriteData(ctx, w, r, config)
			logger.Tf(ctx, "bypass transcode update ok, %v, token=%vB", config, len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	// Delete bypass transcode task
	ep = "/terraform/v1/bypass/transcode/delete"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			var token string
			var taskID string
			if err := ParseBody(ctx, r.Body, &struct {
				Token *string `json:"token"`
				ID    *string `json:"id"`
			}{
				Token: &token,
				ID:    &taskID,
			}); err != nil {
				return errors.Wrapf(err, "parse body")
			}

			apiSecret := envApiSecret()
			if err := Authenticate(ctx, apiSecret, token, r.Header); err != nil {
				return errors.Wrapf(err, "authenticate")
			}

			if err := v.DeleteTask(ctx, taskID); err != nil {
				return errors.Wrapf(err, "delete bypass transcode task")
			}

			ohttp.WriteData(ctx, w, r, map[string]string{"message": "Bypass transcode task deleted successfully"})
			logger.Tf(ctx, "bypass transcode delete ok, id=%v, token=%vB", taskID, len(token))
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	return nil
}

func (v *BypassTranscodeManager) CreateTask(ctx context.Context, config *BypassTranscodeConfig) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	config.Status = "inactive"

	// Validate configuration
	if err := v.validateConfig(config); err != nil {
		return errors.Wrapf(err, "validate config")
	}

	// Save to Redis
	key := fmt.Sprintf("bypass_transcode:%s", config.ID)
	if b, err := json.Marshal(config); err != nil {
		return errors.Wrapf(err, "marshal config")
	} else if err := v.rdb.Set(ctx, key, b, 0).Err(); err != nil {
		return errors.Wrapf(err, "save to redis")
	}

	v.tasks[config.ID] = config
	logger.Tf(ctx, "bypass transcode task created: %v", config)
	return nil
}

func (v *BypassTranscodeManager) UpdateTask(ctx context.Context, config *BypassTranscodeConfig) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if config.ID == "" {
		return errors.Errorf("task ID is required")
	}

	existing, exists := v.tasks[config.ID]
	if !exists {
		return errors.Errorf("task not found: %v", config.ID)
	}

	config.UpdatedAt = time.Now()
	config.CreatedAt = existing.CreatedAt

	// Validate configuration
	if err := v.validateConfig(config); err != nil {
		return errors.Wrapf(err, "validate config")
	}

	// Save to Redis
	key := fmt.Sprintf("bypass_transcode:%s", config.ID)
	if b, err := json.Marshal(config); err != nil {
		return errors.Wrapf(err, "marshal config")
	} else if err := v.rdb.Set(ctx, key, b, 0).Err(); err != nil {
		return errors.Wrapf(err, "save to redis")
	}

	v.tasks[config.ID] = config
	logger.Tf(ctx, "bypass transcode task updated: %v", config)
	return nil
}

func (v *BypassTranscodeManager) DeleteTask(ctx context.Context, taskID string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if taskID == "" {
		return errors.Errorf("task ID is required")
	}

	// Remove from Redis
	key := fmt.Sprintf("bypass_transcode:%s", taskID)
	if err := v.rdb.Del(ctx, key).Err(); err != nil {
		return errors.Wrapf(err, "delete from redis")
	}

	delete(v.tasks, taskID)
	logger.Tf(ctx, "bypass transcode task deleted: %v", taskID)
	return nil
}

func (v *BypassTranscodeManager) GetAllTasks() []*BypassTranscodeConfig {
	v.mu.RLock()
	defer v.mu.RUnlock()

	tasks := make([]*BypassTranscodeConfig, 0, len(v.tasks))
	for _, task := range v.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

func (v *BypassTranscodeManager) GetTask(taskID string) *BypassTranscodeConfig {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.tasks[taskID]
}

func (v *BypassTranscodeManager) StartTask(ctx context.Context, taskID string) error {
	task := v.GetTask(taskID)
	if task == nil {
		return errors.Errorf("task not found: %v", taskID)
	}

	// Start bypass transcoding processing
	go v.processBypassTranscode(ctx, task)
	return nil
}

func (v *BypassTranscodeManager) validateConfig(config *BypassTranscodeConfig) error {
	// Validate input type
	switch config.InputType {
	case "hls", "srt", "rtmp":
		// Valid
	default:
		return errors.Errorf("invalid input type: %v", config.InputType)
	}

	// Validate output type
	switch config.OutputType {
	case "rtmp", "hls", "srt":
		// Valid
	default:
		return errors.Errorf("invalid output type: %v", config.OutputType)
	}

	// Validate bypass mode
	switch config.BypassMode {
	case "passthrough", "filter":
		// Valid
	default:
		return errors.Errorf("invalid bypass mode: %v", config.BypassMode)
	}

	return nil
}

func (v *BypassTranscodeManager) processBypassTranscode(ctx context.Context, task *BypassTranscodeConfig) {
	logger.Tf(ctx, "start processing bypass transcode: %v -> %v", task.InputURL, task.OutputURL)

	// Update status to active
	task.Status = "active"
	task.UpdatedAt = time.Now()
	v.UpdateTask(ctx, task)

	// TODO: Implement bypass transcoding logic
	// This would involve:
	// 1. Reading from input source (HLS/SRT/RTMP)
	// 2. Applying filters if needed (e.g., SCTE-35 removal)
	// 3. Forwarding to output destination without re-encoding
	// 4. Monitoring stream health and performance

	logger.Tf(ctx, "Bypass transcode processing started: %v", task.Name)
}

// SCTE35Filter represents a filter for SCTE-35 data
type SCTE35Filter struct {
	Enabled     bool     `json:"enabled"`
	RemoveTypes []string `json:"removeTypes"` // Types of SCTE-35 data to remove
	Passthrough bool     `json:"passthrough"` // Whether to pass through filtered data
}

// VideoStreamFilter represents a filter for video stream data
type VideoStreamFilter struct {
	Enabled     bool     `json:"enabled"`
	RemoveTypes []string `json:"removeTypes"` // Types of video data to remove
	Passthrough bool     `json:"passthrough"` // Whether to pass through filtered data
}

// ApplyFilters applies configured filters to the stream data
func (v *BypassTranscodeManager) ApplyFilters(data []byte, filters []string) ([]byte, error) {
	// TODO: Implement filter application logic
	// This would involve:
	// 1. Parsing the stream data
	// 2. Identifying SCTE-35, video metadata, etc.
	// 3. Applying configured filters
	// 4. Returning filtered data

	return data, nil
}
