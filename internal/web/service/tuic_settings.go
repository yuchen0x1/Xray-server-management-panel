package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/tuic"
	"github.com/mhsanaei/3x-ui/v3/internal/util/common"
)

// validateTuicClientFields checks the three identity fields the TUIC server
// needs to authenticate a client: the UUID (keying material), the password
// (RFC5705 token) and the email (the accounting key).
func validateTuicClientFields(client model.Client) error {
	if client.ID == "" {
		return common.NewError("empty client ID")
	}
	if client.Password == "" {
		return common.NewError("tuic client requires a password")
	}
	if client.Email == "" {
		return common.NewError("empty client email")
	}
	return nil
}

// validateTuicClients runs validateTuicClientFields over every client of a
// TUIC inbound. Non-TUIC inbounds pass through untouched.
func (s *InboundService) validateTuicClients(inbound *model.Inbound) error {
	if inbound == nil || inbound.Protocol != model.TUIC {
		return nil
	}
	clients, err := s.GetClients(inbound)
	if err != nil {
		return err
	}
	for _, client := range clients {
		if err := validateTuicClientFields(client); err != nil {
			return err
		}
	}
	return validateTuicClientIdentities(clients)
}

// validateTuicClientIdentities rejects a TUIC client whose ID is not a UUID,
// and two clients on one inbound that share a UUID. The TUIC server derives
// its per-client accounting key from the UUID, so two clients sharing one
// would have their traffic merged into a single row.
func validateTuicClientIdentities(clients []model.Client) error {
	seen := make(map[string]string, len(clients))
	for _, client := range clients {
		if client.ID == "" {
			continue
		}
		if _, err := uuid.Parse(client.ID); err != nil {
			return common.NewError("TUIC: invalid client UUID")
		}
		key := strings.ToLower(client.ID)
		if email, exists := seen[key]; exists && email != client.Email {
			return common.NewError("TUIC: duplicate client UUID within inbound")
		}
		seen[key] = client.Email
	}
	return nil
}

func normalizeTuicSettings(inbound *model.Inbound) error {
	if inbound == nil || inbound.Protocol != model.TUIC || inbound.Settings == "" {
		return nil
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal([]byte(inbound.Settings), &root); err != nil {
		return fmt.Errorf("invalid TUIC settings: %w", err)
	}
	if err := normalizeTuicSettingsBlock(root); err != nil {
		return err
	}
	if rawServer, ok := root["server"]; ok && len(rawServer) > 0 {
		var server map[string]json.RawMessage
		if err := json.Unmarshal(rawServer, &server); err != nil {
			return fmt.Errorf("invalid TUIC server settings: %w", err)
		}
		if err := normalizeTuicSettingsBlock(server); err != nil {
			return err
		}
		encoded, err := json.Marshal(server)
		if err != nil {
			return fmt.Errorf("encode TUIC server settings: %w", err)
		}
		root["server"] = encoded
	}
	encoded, err := json.Marshal(root)
	if err != nil {
		return fmt.Errorf("encode TUIC settings: %w", err)
	}
	inbound.Settings = string(encoded)
	return nil
}

func normalizeTuicSettingsBlock(settings map[string]json.RawMessage) error {
	if raw, ok := settings["congestion_control"]; ok {
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return fmt.Errorf("TUIC congestion_control must be a string: %w", err)
		}
		normalized, err := tuic.NormalizeCongestionControl(value)
		if err != nil {
			return err
		}
		encoded, _ := json.Marshal(normalized)
		settings["congestion_control"] = encoded
	}
	if raw, ok := settings["log_level"]; ok {
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return fmt.Errorf("TUIC log_level must be a string: %w", err)
		}
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "warning" {
			value = "warn"
		}
		if value == "" {
			value = "info"
		}
		switch value {
		case "debug", "info", "warn", "error":
		default:
			return fmt.Errorf("TUIC log_level %q is unsupported", value)
		}
		encoded, _ := json.Marshal(value)
		settings["log_level"] = encoded
	}
	if raw, ok := settings["udp_relay_mode"]; ok {
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return fmt.Errorf("TUIC udp_relay_mode must be a string: %w", err)
		}
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			value = "native"
		}
		if value != "native" && value != "quic" {
			return fmt.Errorf("TUIC udp_relay_mode %q is unsupported", value)
		}
		encoded, _ := json.Marshal(value)
		settings["udp_relay_mode"] = encoded
	}
	if raw, ok := settings["max_udp_relay_packet_size"]; ok {
		var value int
		if err := json.Unmarshal(raw, &value); err != nil {
			return fmt.Errorf("TUIC max_udp_relay_packet_size must be an integer: %w", err)
		}
		if value > 65507 {
			return fmt.Errorf("TUIC max_udp_relay_packet_size must not exceed %d", 65245)
		}
		if value > 65245 {
			value = 65245
		}
		encoded, _ := json.Marshal(value)
		settings["max_udp_relay_packet_size"] = encoded
	}
	return nil
}
