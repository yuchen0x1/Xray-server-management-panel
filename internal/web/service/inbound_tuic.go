package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/logger"
	"github.com/mhsanaei/3x-ui/v3/internal/tuic"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

// DesiredTuicInstances derives the TUIC server instances this panel should be
// running: one per enabled local tuic inbound, serving only the clients that
// are enabled in the inbound settings and not depletion-disabled in
// client_traffics. Inbounds whose every client is filtered away are omitted so
// Reconcile stops their server.
func (s *InboundService) DesiredTuicInstances() ([]tuic.Instance, error) {
	db := database.GetDB()
	var inbounds []*model.Inbound
	err := db.Model(model.Inbound{}).
		Where("protocol = ? AND enable = ? AND node_id IS NULL", model.TUIC, true).
		Find(&inbounds).Error
	if err != nil {
		return nil, err
	}
	if len(inbounds) == 0 {
		return nil, nil
	}

	instances := make([]tuic.Instance, 0, len(inbounds))
	for _, ib := range inbounds {
		built, err := s.buildRuntimeInboundForAPI(db, ib)
		if err != nil {
			return nil, err
		}
		inst, ok := tuic.InstanceFromInbound(built)
		if !ok {
			continue
		}
		instances = append(instances, inst)
	}
	emails := make([]string, 0)
	for _, inst := range instances {
		for _, e := range inst.Clients {
			emails = append(emails, e.Email)
		}
	}
	disabled, err := trafficDisabledEmails(db, emails)
	if err != nil {
		return nil, err
	}
	served := instances[:0]
	for _, inst := range instances {
		kept := make([]tuic.TuicClientSettings, 0, len(inst.Clients))
		for _, e := range inst.Clients {
			if _, off := disabled[e.Email]; !off {
				kept = append(kept, e)
			}
		}
		inst.Clients = kept
		if len(kept) > 0 {
			served = append(served, inst)
		}
	}
	return served, nil
}

// applyLocalTuic pushes a single local tuic inbound's current client set to
// its TUIC server right after a client edit commits, so an add, removal,
// re-key or enable-toggle takes effect immediately instead of waiting up to
// 10s for the reconcile job. It re-reads the inbound so it sees the committed
// settings and is a no-op for node-owned or non-tuic inbounds. Failures are
// logged and swallowed: the reconcile job is the backstop.
func (s *InboundService) applyLocalTuic(inboundId int) {
	inbound, err := s.GetInbound(inboundId)
	if err != nil || inbound == nil || inbound.Protocol != model.TUIC || inbound.NodeID != nil {
		return
	}
	rt, err := s.runtimeFor(inbound)
	if err != nil {
		return
	}
	payload := inbound
	if inbound.Enable {
		if built, bErr := s.buildRuntimeInboundForAPI(database.GetDB(), inbound); bErr == nil {
			payload = built
		}
	}
	if err := rt.UpdateInbound(context.Background(), inbound, payload); err != nil {
		logger.Debug("tuic: immediate client apply failed for inbound", inboundId, ":", err)
	}
}

// trafficDisabledEmails reports which emails have a switched-off stats row. The
// table is email-keyed and its inbound_id goes stale, so never filter on it.
func trafficDisabledEmails(db *gorm.DB, emails []string) (map[string]struct{}, error) {
	disabled := make(map[string]struct{})
	for _, batch := range chunkStrings(uniqueNonEmptyStrings(emails), sqlInChunk) {
		var page []string
		if err := db.Model(xray.ClientTraffic{}).
			Where("email IN ? AND enable = ?", batch, false).
			Pluck("email", &page).Error; err != nil {
			return nil, err
		}
		for _, e := range page {
			disabled[e] = struct{}{}
		}
	}
	return disabled, nil
}
