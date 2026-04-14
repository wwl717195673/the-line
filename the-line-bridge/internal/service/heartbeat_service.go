package service

import (
	"context"
	"log"
	"sync"
	"time"

	"the-line-bridge/internal/client"
	"the-line-bridge/internal/runtime"
)

type HeartbeatService struct {
	client        *client.TheLineClient
	rt            runtime.OpenClawRuntime
	integrationID uint64
	interval      time.Duration
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

func NewHeartbeatService(c *client.TheLineClient, rt runtime.OpenClawRuntime, integrationID uint64, intervalSec int) *HeartbeatService {
	if intervalSec <= 0 {
		intervalSec = 60
	}
	return &HeartbeatService{
		client:        c,
		rt:            rt,
		integrationID: integrationID,
		interval:      time.Duration(intervalSec) * time.Second,
	}
}

func (s *HeartbeatService) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.sendHeartbeat()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.sendHeartbeat()
			}
		}
	}()
}

func (s *HeartbeatService) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

func (s *HeartbeatService) sendHeartbeat() {
	health, err := s.rt.Health(context.Background())
	status := "healthy"
	lastError := ""
	if err != nil {
		status = "degraded"
		lastError = err.Error()
	} else {
		status = health.Status
	}

	err = s.client.Heartbeat(client.HeartbeatRequest{
		IntegrationID: s.integrationID,
		BridgeVersion: "0.1.0",
		Status:        status,
		LastError:     lastError,
	})
	if err != nil {
		log.Printf("heartbeat failed: %v", err)
	}
}
