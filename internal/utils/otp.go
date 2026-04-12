package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"home_care_backend/internal/config"
)

// OTPExpiry returns the OTP expiration time
func OTPExpiry() time.Time {
	return time.Now().Add(time.Duration(config.AppConfig.OTPExpiryMinutes) * time.Minute)
}

// ── MessageCentral API types ───────────────────────────────────────────────

// OTPData is the response from MessageCentral when sending an OTP
type OTPData struct {
	ResponseCode int    `json:"responseCode"`
	Message      string `json:"message"`
	Data         struct {
		VerificationID string `json:"verificationId"`
		MobileNumber   string `json:"mobileNumber"`
		ResponseCode   string `json:"responseCode"`
		Timeout        string `json:"timeout"`
		TransactionID  string `json:"transactionId"`
	} `json:"data"`
}

// VerifyOTPData is the response from MessageCentral when verifying an OTP
type VerifyOTPData struct {
	ResponseCode int    `json:"responseCode"`
	Message      string `json:"message"`
	Data         struct {
		VerificationID     int     `json:"verificationId"`
		MobileNumber       string  `json:"mobileNumber"`
		VerificationStatus string  `json:"verificationStatus"`
		ResponseCode       string  `json:"responseCode"`
		ErrorMessage       *string `json:"errorMessage"`
		TransactionID      string  `json:"transactionId"`
		AuthToken          *string `json:"authToken"`
	} `json:"data"`
}

// ── SendOTP calls MessageCentral to dispatch an OTP SMS ───────────────────

// SendOTP sends an OTP to the given mobile number via MessageCentral.
// Returns OTPData containing verificationId which must be stored for later verification.
func SendOTP(phoneNumber string) (*OTPData, error) {
	url := fmt.Sprintf(
		"https://cpaas.messagecentral.com/verification/v3/send?countryCode=%s&customerId=%s&flowType=SMS&mobileNumber=%s",
		config.AppConfig.MCCountryCode,
		config.AppConfig.MCCustomerID,
		phoneNumber,
	)

	req, err := http.NewRequest("POST", url, strings.NewReader(""))
	if err != nil {
		return nil, fmt.Errorf("failed to build OTP request: %w", err)
	}
	req.Header.Add("authToken", config.AppConfig.MCAuthToken)

	// No timeout — MessageCentral can be slow; let the request complete naturally
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OTP send request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read OTP response: %w", err)
	}

	fmt.Printf("[OTP] Send HTTP %d | body: %s\n", resp.StatusCode, string(body))

	var otpData OTPData
	if err := json.Unmarshal(body, &otpData); err != nil {
		return nil, fmt.Errorf("failed to parse OTP response: %w", err)
	}

	fmt.Printf("[OTP] Sent to %s | verificationId: %s\n", phoneNumber, otpData.Data.VerificationID)
	return &otpData, nil
}

// VerifyOTP validates the OTP code with MessageCentral.
// Returns VerifyOTPData; check Data.VerificationStatus == "ACTIVE" for success.
func VerifyOTP(phoneNumber, verificationID, otp string) (*VerifyOTPData, error) {
	url := fmt.Sprintf(
		"https://cpaas.messagecentral.com/verification/v3/validateOtp?countryCode=%s&mobileNumber=%s&verificationId=%s&customerId=%s&code=%s",
		config.AppConfig.MCCountryCode,
		phoneNumber,
		verificationID,
		config.AppConfig.MCCustomerID,
		otp,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build verify request: %w", err)
	}
	req.Header.Add("authToken", config.AppConfig.MCAuthToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OTP verify request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read verify response: %w", err)
	}

	var verifyResp VerifyOTPData
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return nil, fmt.Errorf("failed to parse verify response: %w", err)
	}

	fmt.Printf("[OTP] Verify for %s | status: %s\n", phoneNumber, verifyResp.Data.VerificationStatus)
	return &verifyResp, nil
}
