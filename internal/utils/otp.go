package utils

import (
	"fmt"
	"math/rand"
	"time"

	"home_care_backend/internal/config"
)

// GenerateOTP generates a random N-digit OTP
func GenerateOTP(length int) string {
	// In test mode, return the test OTP
	if config.AppConfig.OTPTestMode {
		return config.AppConfig.OTPTestValue
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	min := 1
	max := 9
	for i := 1; i < length; i++ {
		min *= 10
		max = max*10 + 9
	}
	otp := min + rand.Intn(max-min+1)
	return fmt.Sprintf("%0*d", length, otp)
}

// OTPExpiry returns the OTP expiration time
func OTPExpiry() time.Time {
	return time.Now().Add(time.Duration(config.AppConfig.OTPExpiryMinutes) * time.Minute)
}

// SendOTPSMS simulates sending an OTP via SMS
// In production, integrate a real SMS provider (Twilio, AWS SNS, MSG91, etc.)
func SendOTPSMS(phoneNumber, otp string) error {
	// TODO: Integrate with real SMS provider
	// Example Twilio integration:
	// client := twilio.NewRestClient()
	// params := &twilioApi.CreateMessageParams{}
	// params.SetTo(phoneNumber)
	// params.SetFrom(config.AppConfig.TwilioFromNumber)
	// params.SetBody(fmt.Sprintf("Your HomeCare OTP is: %s. Valid for %d minutes.", otp, config.AppConfig.OTPExpiryMinutes))
	// _, err := client.Api.CreateMessage(params)
	// return err

	// For development, just log the OTP
	fmt.Printf("[OTP] Phone: %s | OTP: %s\n", phoneNumber, otp)
	return nil
}
