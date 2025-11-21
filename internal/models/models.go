package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User represents a SafeTrace user
type User struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	Phone           string          `json:"phone" db:"phone"`
	Name            string          `json:"name" db:"name"`
	TrustedContacts TrustedContacts `json:"trusted_contacts" db:"trusted_contacts"`
	Settings        UserSettings    `json:"settings" db:"settings"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// Contact represents a trusted contact
type Contact struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// TrustedContacts is a slice of contacts stored as JSONB
type TrustedContacts []Contact

func (t TrustedContacts) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *TrustedContacts) Scan(value interface{}) error {
	if value == nil {
		*t = TrustedContacts{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, t)
}

// UserSettings represents user preferences
type UserSettings struct {
	HeartbeatInterval   int  `json:"heartbeat_interval"`    // seconds
	SilentPromptTimeout int  `json:"silent_prompt_timeout"` // seconds
	AutoEscalatePolice  bool `json:"auto_escalate_police"`
	ShareAudio          bool `json:"share_audio"`
	PanicGesture        string `json:"panic_gesture"` // "power_button_3x" | "shake"
}

func (s UserSettings) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *UserSettings) Scan(value interface{}) error {
	if value == nil {
		*s = UserSettings{
			HeartbeatInterval:   180,
			SilentPromptTimeout: 10,
			AutoEscalatePolice:  false,
			ShareAudio:          false,
			PanicGesture:        "power_button_3x",
		}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, s)
}

// Heartbeat represents a location/sensor update
type Heartbeat struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	Source     string    `json:"source" db:"source"` // "http" | "sms"
	Lat        float64   `json:"lat" db:"lat"`
	Lng        float64   `json:"lng" db:"lng"`
	AccuracyM  int       `json:"accuracy_m" db:"accuracy_m"`
	CellInfo   CellInfo  `json:"cell_info" db:"cell_info"`
	BatteryPct *int      `json:"battery_pct,omitempty" db:"battery_pct"`
	Speed      *float64  `json:"speed,omitempty" db:"speed"` // km/h
	LastGasp   bool      `json:"last_gasp" db:"last_gasp"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
	Signature  string    `json:"signature" db:"signature"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// CellInfo represents cellular network information
type CellInfo struct {
	MCC       int    `json:"mcc"`        // Mobile Country Code
	MNC       int    `json:"mnc"`        // Mobile Network Code
	CID       int    `json:"cid"`        // Cell ID
	LAC       int    `json:"lac"`        // Location Area Code
	RSSI      int    `json:"rssi"`       // Signal strength
	NetworkType string `json:"network_type"` // "2G" | "3G" | "4G" | "5G"
	Neighbors []NeighborCell `json:"neighbors,omitempty"`
}

type NeighborCell struct {
	CID  int `json:"cid"`
	RSSI int `json:"rssi"`
}

func (c CellInfo) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CellInfo) Scan(value interface{}) error {
	if value == nil {
		*c = CellInfo{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// LastGasp represents a final known location before connectivity loss
type LastGasp struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Lat       float64   `json:"lat" db:"lat"`
	Lng       float64   `json:"lng" db:"lng"`
	AccuracyM int       `json:"accuracy_m" db:"accuracy_m"`
	CellInfo  CellInfo  `json:"cell_info" db:"cell_info"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiryTs  time.Time `json:"expiry_ts" db:"expiry_ts"`
}

// Alert represents a safety alert
type Alert struct {
	ID         uuid.UUID    `json:"id" db:"id"`
	UserID     uuid.UUID    `json:"user_id" db:"user_id"`
	State      AlertState   `json:"state" db:"state"`
	Score      int          `json:"score" db:"score"`
	Reason     string       `json:"reason" db:"reason"`
	SentTo     []string     `json:"sent_to" db:"sent_to"`
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
	ResolvedAt *time.Time   `json:"resolved_at,omitempty" db:"resolved_at"`
}

type AlertState string

const (
	AlertStateCaution AlertState = "CAUTION"
	AlertStateAtRisk  AlertState = "AT_RISK"
	AlertStateAlert   AlertState = "ALERT"
)

func (s AlertState) Value() (driver.Value, error) {
	return string(s), nil
}

func (s *AlertState) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if sv, ok := value.(string); ok {
		*s = AlertState(sv)
	}
	return nil
}

type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = StringArray{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, s)
}

// BlackboxTrail represents uploaded sensor trail
type BlackboxTrail struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	StartTs    time.Time `json:"start_ts" db:"start_ts"`
	EndTs      time.Time `json:"end_ts" db:"end_ts"`
	DataPoints int       `json:"data_points" db:"data_points"`
	FileURL    string    `json:"file_url" db:"file_url"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}

// BlackboxEntry represents a single trail data point
type BlackboxEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Lat        float64   `json:"lat"`
	Lng        float64   `json:"lng"`
	AccuracyM  int       `json:"accuracy_m"`
	CellInfo   CellInfo  `json:"cell_info"`
	SensorData SensorData `json:"sensor_data,omitempty"`
}

type SensorData struct {
	AccelX float64 `json:"accel_x"`
	AccelY float64 `json:"accel_y"`
	AccelZ float64 `json:"accel_z"`
	GyroX  float64 `json:"gyro_x"`
	GyroY  float64 `json:"gyro_y"`
	GyroZ  float64 `json:"gyro_z"`
}

// UserState represents current safety state (stored in Redis)
type UserState struct {
	UserID         uuid.UUID  `json:"user_id"`
	State          string     `json:"state"` // SAFE | CAUTION | AT_RISK | ALERT | WAIT_LASTGASP
	Score          int        `json:"score"`
	LastHeartbeat  time.Time  `json:"last_heartbeat"`
	LastGaspActive bool       `json:"last_gasp_active"`
	LastGaspExpiry *time.Time `json:"last_gasp_expiry,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
