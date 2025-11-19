package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
)

// SMSParser handles parsing of compressed SMS heartbeat payloads
type SMSParser struct{}

func NewSMSParser() *SMSParser {
	return &SMSParser{}
}

// ParseHeartbeatSMS parses compressed SMS format:
// uid=uuid;ts=2025-11-19T12:50Z;lat=6.5244;lng=3.3792;acc=200;cell=621,20,12345,678,-85;sig=abc123
func (sp *SMSParser) ParseHeartbeatSMS(smsBody string) (*models.Heartbeat, error) {
	parts := strings.Split(smsBody, ";")
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid SMS format: insufficient fields")
	}

	hb := &models.Heartbeat{
		ID:        uuid.New(),
		Source:    "sms",
		CreatedAt: time.Now(),
	}

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "uid":
			userID, err := uuid.Parse(value)
			if err != nil {
				return nil, fmt.Errorf("invalid user ID: %w", err)
			}
			hb.UserID = userID

		case "ts":
			timestamp, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return nil, fmt.Errorf("invalid timestamp: %w", err)
			}
			hb.Timestamp = timestamp

		case "lat":
			lat, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid latitude: %w", err)
			}
			hb.Lat = lat

		case "lng":
			lng, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid longitude: %w", err)
			}
			hb.Lng = lng

		case "acc":
			acc, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid accuracy: %w", err)
			}
			hb.AccuracyM = acc

		case "cell":
			cellInfo, err := sp.parseCellInfo(value)
			if err != nil {
				return nil, fmt.Errorf("invalid cell info: %w", err)
			}
			hb.CellInfo = cellInfo

		case "bat":
			bat, err := strconv.Atoi(value)
			if err == nil {
				hb.BatteryPct = &bat
			}

		case "spd":
			spd, err := strconv.ParseFloat(value, 64)
			if err == nil {
				hb.Speed = &spd
			}

		case "lg":
			hb.LastGasp = value == "1" || value == "true"

		case "sig":
			hb.Signature = value
		}
	}

	// Validate required fields
	if hb.UserID == uuid.Nil {
		return nil, fmt.Errorf("missing user ID")
	}
	if hb.Timestamp.IsZero() {
		return nil, fmt.Errorf("missing timestamp")
	}
	if hb.Signature == "" {
		return nil, fmt.Errorf("missing signature")
	}

	return hb, nil
}

// parseCellInfo parses cell info from CSV format: mcc,mnc,cid,lac,rssi
func (sp *SMSParser) parseCellInfo(cellStr string) (models.CellInfo, error) {
	parts := strings.Split(cellStr, ",")
	if len(parts) < 5 {
		return models.CellInfo{}, fmt.Errorf("invalid cell info format")
	}

	mcc, err := strconv.Atoi(parts[0])
	if err != nil {
		return models.CellInfo{}, fmt.Errorf("invalid MCC: %w", err)
	}

	mnc, err := strconv.Atoi(parts[1])
	if err != nil {
		return models.CellInfo{}, fmt.Errorf("invalid MNC: %w", err)
	}

	cid, err := strconv.Atoi(parts[2])
	if err != nil {
		return models.CellInfo{}, fmt.Errorf("invalid CID: %w", err)
	}

	lac, err := strconv.Atoi(parts[3])
	if err != nil {
		return models.CellInfo{}, fmt.Errorf("invalid LAC: %w", err)
	}

	rssi, err := strconv.Atoi(parts[4])
	if err != nil {
		return models.CellInfo{}, fmt.Errorf("invalid RSSI: %w", err)
	}

	return models.CellInfo{
		MCC:  mcc,
		MNC:  mnc,
		CID:  cid,
		LAC:  lac,
		RSSI: rssi,
	}, nil
}

// BuildSMSPayload creates compressed SMS payload (for mobile client reference)
func (sp *SMSParser) BuildSMSPayload(hb *models.Heartbeat) string {
	parts := []string{
		fmt.Sprintf("uid=%s", hb.UserID),
		fmt.Sprintf("ts=%s", hb.Timestamp.Format(time.RFC3339)),
		fmt.Sprintf("lat=%.6f", hb.Lat),
		fmt.Sprintf("lng=%.6f", hb.Lng),
		fmt.Sprintf("acc=%d", hb.AccuracyM),
		fmt.Sprintf("cell=%d,%d,%d,%d,%d",
			hb.CellInfo.MCC, hb.CellInfo.MNC, hb.CellInfo.CID,
			hb.CellInfo.LAC, hb.CellInfo.RSSI),
	}

	if hb.BatteryPct != nil {
		parts = append(parts, fmt.Sprintf("bat=%d", *hb.BatteryPct))
	}

	if hb.Speed != nil {
		parts = append(parts, fmt.Sprintf("spd=%.1f", *hb.Speed))
	}

	if hb.LastGasp {
		parts = append(parts, "lg=1")
	}

	parts = append(parts, fmt.Sprintf("sig=%s", hb.Signature))

	return strings.Join(parts, ";")
}
