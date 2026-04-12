package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"home_care_backend/internal/config"
	"home_care_backend/internal/database"
	"home_care_backend/internal/models"
	"home_care_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const devBypassOTP = "000000"
const devBypassVerificationID = "dev-bypass"

type LoginRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

type VerifyOTPRequest struct {
	PhoneNumber    string `json:"phone_number"     binding:"required"`
	OTP            string `json:"otp"              binding:"required"`
	VerificationID string `json:"verification_id"  binding:"required"`
}

// Login — sends OTP to the given phone number via MessageCentral.
// In dev bypass mode (DEV_OTP_BYPASS=true) skips MessageCentral and
// returns a fixed verification_id so you can test without SMS credits.
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "phone_number is required")
		return
	}

	phone := strings.TrimSpace(req.PhoneNumber)
	if len(phone) < 10 {
		utils.BadRequestResponse(c, "Invalid phone number")
		return
	}

	ctx := context.Background()
	expiry := utils.OTPExpiry()

	// ── Dev bypass mode ────────────────────────────────────────────────────
	if config.AppConfig.DevOTPBypass {
		fmt.Printf("[OTP-DEV] Bypass active for %s — use OTP \"%s\"\n", phone, devBypassOTP)

		// Invalidate prior sessions
		database.Col(database.ColOTPs).UpdateMany(ctx,
			bson.M{"phone_number": phone, "is_used": false},
			bson.M{"$set": bson.M{"is_used": true}},
		)

		otpDoc := models.OTP{
			ID:             uuid.New().String(),
			PhoneNumber:    phone,
			VerificationID: devBypassVerificationID,
			IsUsed:         false,
			ExpiresAt:      expiry,
			CreatedAt:      time.Now(),
		}
		database.Col(database.ColOTPs).InsertOne(ctx, otpDoc)

		c.JSON(http.StatusOK, gin.H{
			"success":         true,
			"message":         fmt.Sprintf("DEV MODE: use OTP \"%s\"", devBypassOTP),
			"verification_id": devBypassVerificationID,
		})
		return
	}

	// ── Production: call MessageCentral ────────────────────────────────────
	otpResp, err := utils.SendOTP(phone)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to send OTP: "+err.Error())
		return
	}

	// MessageCentral returned an error code (e.g. 508 = insufficient credits)
	if otpResp.ResponseCode != 200 {
		utils.ErrorResponse(c, http.StatusServiceUnavailable,
			fmt.Sprintf("OTP service error: %s", otpResp.Message),
			"otp_service_error",
		)
		return
	}

	verificationID := otpResp.Data.VerificationID
	if verificationID == "" {
		utils.InternalServerErrorResponse(c, "OTP service did not return a verification ID")
		return
	}

	// Invalidate prior pending OTPs for this phone
	database.Col(database.ColOTPs).UpdateMany(ctx,
		bson.M{"phone_number": phone, "is_used": false},
		bson.M{"$set": bson.M{"is_used": true}},
	)

	otpDoc := models.OTP{
		ID:             uuid.New().String(),
		PhoneNumber:    phone,
		VerificationID: verificationID,
		IsUsed:         false,
		ExpiresAt:      expiry,
		CreatedAt:      time.Now(),
	}
	if _, err := database.Col(database.ColOTPs).InsertOne(ctx, otpDoc); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to save OTP session")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "OTP sent successfully",
		"verification_id": verificationID,
	})
}

// VerifyOTP — validates OTP (or dev bypass), then issues a JWT.
func VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "phone_number, otp and verification_id are required")
		return
	}

	phone := strings.TrimSpace(req.PhoneNumber)
	otp := strings.TrimSpace(req.OTP)
	verificationID := strings.TrimSpace(req.VerificationID)
	ctx := context.Background()

	// 1. Look up the OTP session in our DB
	var otpDoc models.OTP
	err := database.Col(database.ColOTPs).FindOne(ctx, bson.M{
		"phone_number":    phone,
		"verification_id": verificationID,
		"is_used":         false,
		"expires_at":      bson.M{"$gt": time.Now()},
	}).Decode(&otpDoc)

	if err == mongo.ErrNoDocuments {
		utils.ErrorResponse(c, http.StatusUnauthorized,
			"Session expired or invalid. Please request a new OTP.", "session_invalid")
		return
	}
	if err != nil {
		utils.InternalServerErrorResponse(c, "OTP lookup error")
		return
	}

	// 2. Validate OTP
	if config.AppConfig.DevOTPBypass && verificationID == devBypassVerificationID {
		// Dev mode: accept the fixed bypass OTP
		if otp != devBypassOTP {
			utils.ErrorResponse(c, http.StatusUnauthorized,
				fmt.Sprintf("Invalid OTP. In DEV mode use \"%s\"", devBypassOTP), "otp_invalid")
			return
		}
		fmt.Printf("[OTP-DEV] Bypass accepted for %s\n", phone)
	} else {
		// Production: verify with MessageCentral
		verifyResp, err := utils.VerifyOTP(phone, verificationID, otp)
		if err != nil {
			utils.InternalServerErrorResponse(c, "OTP verification service error: "+err.Error())
			return
		}
		if verifyResp.ResponseCode != 200 || verifyResp.Data.VerificationStatus != "VERIFICATION_COMPLETED" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired OTP", "otp_invalid")
			return
		}
	}

	// 3. Mark session used
	database.Col(database.ColOTPs).UpdateOne(ctx,
		bson.M{"_id": otpDoc.ID},
		bson.M{"$set": bson.M{"is_used": true}},
	)

	// 4. Find or create user
	var user models.User
	err = database.Col(database.ColUsers).FindOne(ctx, bson.M{"phone_number": phone}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		now := time.Now()
		user = models.User{
			ID:            uuid.New().String(),
			PhoneNumber:   phone,
			Role:          models.RolePatient,
			Status:        models.Active,
			BlockStatus:   models.Unblocked,
			UserService:   models.Unsubscribed,
			ServiceStatus: models.New,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if _, err := database.Col(database.ColUsers).InsertOne(ctx, user); err != nil {
			utils.InternalServerErrorResponse(c, "Failed to create user account")
			return
		}
	} else if err != nil {
		utils.InternalServerErrorResponse(c, "Database error")
		return
	}

	// 5. Check account status
	if user.BlockStatus == models.Blocked {
		utils.ForbiddenResponse(c, "Account blocked. Please contact support.")
		return
	}
	if user.Status == models.Suspended {
		utils.ForbiddenResponse(c, "Account suspended. Please contact support.")
		return
	}

	// 6. Generate JWT
	token, err := utils.GenerateToken(user.ID, user.PhoneNumber, string(user.Role))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "Login successful",
		"token":        token,
		"access_token": token,
		"user_id":      user.ID,
		"role":         string(user.Role),
		"name":         user.Name,
		"phone_number": user.PhoneNumber,
		"is_new_user":  user.Name == "",
	})
}
