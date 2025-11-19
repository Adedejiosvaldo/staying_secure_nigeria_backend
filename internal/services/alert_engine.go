package services

import (
	"context"
	"fmt"
	"time"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
	"firebase.google.com/go/v4/messaging"
	"github.com/adedejiosvaldo/safetrace/backend/internal/config"
	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
)

type AlertEngine struct {
	cfg          *config.Config
	twilioClient *twilio.RestClient
	fcmClient    *messaging.Client
}

func NewAlertEngine(cfg *config.Config, fcmClient *messaging.Client) *AlertEngine {
	twilioClient := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: cfg.TwilioAccountSID,
		Password: cfg.TwilioAuthToken,
	})

	return &AlertEngine{
		cfg:          cfg,
		twilioClient: twilioClient,
		fcmClient:    fcmClient,
	}
}

// SendAlertToContacts sends alerts to all trusted contacts
func (ae *AlertEngine) SendAlertToContacts(
	ctx context.Context,
	user *models.User,
	heartbeat *models.Heartbeat,
	score int,
	reason string,
) error {
	if len(user.TrustedContacts) == 0 {
		return fmt.Errorf("no trusted contacts configured")
	}

	// Generate map link
	mapLink := ae.generateMapLink(heartbeat.Lat, heartbeat.Lng)

	// Build message
	message := ae.buildAlertMessage(user, heartbeat, score, reason, mapLink)

	// Send to each contact
	var errors []error
	for _, contact := range user.TrustedContacts {
		// Send SMS
		if err := ae.SendSMS(contact.Phone, message); err != nil {
			errors = append(errors, fmt.Errorf("failed to send SMS to %s: %w", contact.Phone, err))
		}

		// Try WhatsApp as well (if number supports it)
		// WhatsApp requires "whatsapp:" prefix
		if err := ae.SendWhatsApp(contact.Phone, message); err != nil {
			// Log but don't fail - WhatsApp is optional
			fmt.Printf("WhatsApp failed for %s: %v\n", contact.Phone, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some alerts failed: %v", errors)
	}

	return nil
}

// SendSMS sends an SMS via Twilio
func (ae *AlertEngine) SendSMS(to, message string) error {
	params := &twilioApi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(ae.cfg.TwilioPhoneNumber)
	params.SetBody(message)

	resp, err := ae.twilioClient.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("twilio SMS error: %w", err)
	}

	if resp.ErrorCode != nil {
		return fmt.Errorf("twilio error code: %d, message: %s", *resp.ErrorCode, *resp.ErrorMessage)
	}

	return nil
}

// SendWhatsApp sends a WhatsApp message via Twilio
func (ae *AlertEngine) SendWhatsApp(to, message string) error {
	params := &twilioApi.CreateMessageParams{}
	params.SetTo("whatsapp:" + to)
	params.SetFrom("whatsapp:" + ae.cfg.TwilioPhoneNumber)
	params.SetBody(message)

	resp, err := ae.twilioClient.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("twilio WhatsApp error: %w", err)
	}

	if resp.ErrorCode != nil {
		return fmt.Errorf("twilio error code: %d, message: %s", *resp.ErrorCode, *resp.ErrorMessage)
	}

	return nil
}

// SendPushNotification sends a push notification via FCM
func (ae *AlertEngine) SendPushNotification(ctx context.Context, fcmToken, title, body string) error {
	if ae.fcmClient == nil {
		return fmt.Errorf("FCM client not initialized")
	}

	message := &messaging.Message{
		Token: fcmToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Priority: messaging.PriorityHigh,
				Sound:    "default",
			},
		},
	}

	_, err := ae.fcmClient.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("FCM error: %w", err)
	}

	return nil
}

// SendSilentPing sends a silent check notification to user
func (ae *AlertEngine) SendSilentPing(ctx context.Context, fcmToken string) error {
	return ae.SendPushNotification(
		ctx,
		fcmToken,
		"Are you safe?",
		"Tap to confirm you're okay",
	)
}

// buildAlertMessage constructs the alert SMS message
func (ae *AlertEngine) buildAlertMessage(
	user *models.User,
	hb *models.Heartbeat,
	score int,
	reason string,
	mapLink string,
) string {
	timestamp := hb.Timestamp.Format("Jan 2, 3:04 PM")
	
	msg := fmt.Sprintf(
		"🚨 SAFETRACE ALERT\n\n"+
			"%s may be in danger.\n\n"+
			"Last seen: %s\n"+
			"Location: %.6f, %.6f (±%dm)\n"+
			"Confidence: %d%%\n"+
			"Reason: %s\n\n"+
			"Map: %s\n\n"+
			"Please check on them immediately.\n"+
			"Contact: %s",
		user.Name,
		timestamp,
		hb.Lat,
		hb.Lng,
		hb.AccuracyM,
		score,
		reason,
		mapLink,
		user.Phone,
	)

	return msg
}

// generateMapLink creates a link to view location on map
func (ae *AlertEngine) generateMapLink(lat, lng float64) string {
	if ae.cfg.MapboxToken != "" {
		// Mapbox static map
		return fmt.Sprintf(
			"https://api.mapbox.com/styles/v1/mapbox/streets-v11/static/pin-s+f74e4e(%.6f,%.6f)/%.6f,%.6f,15,0/600x400@2x?access_token=%s",
			lng, lat, lng, lat, ae.cfg.MapboxToken,
		)
	}
	// Fallback to Google Maps
	return fmt.Sprintf("https://www.google.com/maps?q=%.6f,%.6f", lat, lng)
}

// SendHeartbeatReceivedConfirmation sends confirmation to user (optional)
func (ae *AlertEngine) SendHeartbeatReceivedConfirmation(ctx context.Context, fcmToken string) error {
	return ae.SendPushNotification(
		ctx,
		fcmToken,
		"Protection Active",
		"Your location has been updated",
	)
}

// SendLastGaspAcknowledgment confirms LastGasp received
func (ae *AlertEngine) SendLastGaspAcknowledgment(to string) error {
	message := "SafeTrace: Your emergency location has been recorded. We're monitoring your situation."
	return ae.SendSMS(to, message)
}

// SendAlertResolved notifies contacts that user is safe
func (ae *AlertEngine) SendAlertResolved(ctx context.Context, user *models.User) error {
	message := fmt.Sprintf(
		"✅ SafeTrace Update\n\n"+
			"%s has confirmed they are safe.\n"+
			"Alert resolved at %s.",
		user.Name,
		time.Now().Format("Jan 2, 3:04 PM"),
	)

	var errors []error
	for _, contact := range user.TrustedContacts {
		if err := ae.SendSMS(contact.Phone, message); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some resolution notifications failed: %v", errors)
	}

	return nil
}
