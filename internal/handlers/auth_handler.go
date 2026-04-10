package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"home_care_backend/internal/database"
	"home_care_backend/internal/models"
	"home_care_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type LoginRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

type VerifyOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	OTP         string `json:"otp"          binding:"required"`
}

// Login — sends OTP to the given phone number
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

	otp := utils.GenerateOTP(4)
	expiry := utils.OTPExpiry()
	ctx := context.Background()

	// Invalidate previous unused OTPs for this phone
	database.Col(database.ColOTPs).UpdateMany(ctx,
		bson.M{"phone_number": phone, "is_used": false},
		bson.M{"$set": bson.M{"is_used": true}},
	)

	otpDoc := models.OTP{
		ID:          uuid.New().String(),
		PhoneNumber: phone,
		OTPCode:     otp,
		IsUsed:      false,
		ExpiresAt:   expiry,
		CreatedAt:   time.Now(),
	}
	if _, err := database.Col(database.ColOTPs).InsertOne(ctx, otpDoc); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate OTP")
		return
	}

	if err := utils.SendOTPSMS(phone, otp); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to send OTP")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent successfully",
		"status":  http.StatusOK,
	})
}

// VerifyOTP — verifies OTP and returns JWT token
func VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "phone_number and otp are required")
		return
	}

	phone := strings.TrimSpace(req.PhoneNumber)
	otpCode := strings.TrimSpace(req.OTP)
	ctx := context.Background()

	// Find valid OTP
	var otpDoc models.OTP
	err := database.Col(database.ColOTPs).FindOne(ctx, bson.M{
		"phone_number": phone,
		"otp_code":     otpCode,
		"is_used":      false,
		"expires_at":   bson.M{"$gt": time.Now()},
	}).Decode(&otpDoc)

	if err == mongo.ErrNoDocuments {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired OTP", "otp_invalid")
		return
	}
	if err != nil {
		utils.InternalServerErrorResponse(c, "OTP verification error")
		return
	}

	// Mark OTP used
	database.Col(database.ColOTPs).UpdateOne(ctx,
		bson.M{"_id": otpDoc.ID},
		bson.M{"$set": bson.M{"is_used": true}},
	)

	// Find or create user
	var user models.User
	err = database.Col(database.ColUsers).FindOne(ctx, bson.M{"phone_number": phone}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		now := time.Now()
		user = models.User{
			ID:            uuid.New().String(),
			PhoneNumber:   phone,
			Role:          models.RolePatient,
			Status:        models.UserStatusActive,
			BlockStatus:   models.BlockStatusUnblocked,
			UserService:   models.UserServiceUnsubscribed,
			ServiceStatus: models.ServiceStatusNew,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if _, err := database.Col(database.ColUsers).InsertOne(ctx, user); err != nil {
			utils.InternalServerErrorResponse(c, "Failed to create user")
			return
		}
	} else if err != nil {
		utils.InternalServerErrorResponse(c, "Database error")
		return
	}

	if user.BlockStatus == models.BlockStatusBlocked {
		utils.ForbiddenResponse(c, "Account blocked. Contact support.")
		return
	}

	token, err := utils.GenerateToken(user.ID, user.PhoneNumber, string(user.Role))
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "OTP verified successfully",
		"token":        token,
		"access_token": token,
		"user_id":      user.ID,
	})
}
