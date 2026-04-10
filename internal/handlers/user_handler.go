package handlers

import (
	"context"
	"strings"
	"time"

	"home_care_backend/internal/database"
	"home_care_backend/internal/middleware"
	"home_care_backend/internal/models"
	"home_care_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetUserProfile returns a user's profile by ID
func GetUserProfile(c *gin.Context) {
	userID := c.Param("userId")
	authUserID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	if userID != authUserID && role != string(models.RoleAdmin) && role != string(models.RoleSuperAdmin) {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}

	var user models.User
	err := database.Col(database.ColUsers).FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		utils.NotFoundResponse(c, "User not found")
		return
	}
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch user")
		return
	}

	utils.SuccessResponse(c, "User profile fetched", buildUserResponse(&user))
}

// UpdateUserProfile updates name, email, gender
func UpdateUserProfile(c *gin.Context) {
	userID := c.Param("userId")
	if userID != middleware.GetUserID(c) {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	setFields := bson.M{"updated_at": time.Now()}
	if v := strings.TrimSpace(req.Name); v != "" {
		setFields["name"] = v
	}
	if v := strings.TrimSpace(req.Email); v != "" {
		setFields["email"] = v
	}
	if req.Gender != "" {
		setFields["gender"] = req.Gender
	}

	ctx := context.Background()
	_, err := database.Col(database.ColUsers).UpdateOne(ctx,
		bson.M{"_id": userID},
		bson.M{"$set": setFields},
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update profile")
		return
	}

	var user models.User
	database.Col(database.ColUsers).FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	utils.SuccessResponse(c, "Profile updated", buildUserResponse(&user))
}

// UpdateUserAddress adds or updates address1/address2 embedded in the user doc
func UpdateUserAddress(c *gin.Context) {
	userID := c.Param("userId")
	if userID != middleware.GetUserID(c) {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}

	var req models.UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	ctx := context.Background()
	var user models.User
	if err := database.Col(database.ColUsers).FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		utils.NotFoundResponse(c, "User not found")
		return
	}

	isPrimary := strings.ToLower(req.AddressType) == "address1"
	newAddr := models.Address{
		AddressID:   uuid.New().String(),
		HouseNumber: req.HouseNumber,
		Street:      req.Street,
		Landmark:    req.Landmark,
		PinCode:     req.PinCode,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		IsPrimary:   isPrimary,
		AddressType: req.AddressType,
		CreatedAt:   time.Now(),
	}

	// Replace if existing address with same type, else append
	updated := false
	for i, a := range user.Addresses {
		if a.AddressType == req.AddressType {
			newAddr.AddressID = a.AddressID // preserve ID
			user.Addresses[i] = newAddr
			updated = true
			break
		}
	}
	if !updated {
		user.Addresses = append(user.Addresses, newAddr)
	}

	database.Col(database.ColUsers).UpdateOne(ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"addresses": user.Addresses, "updated_at": time.Now()}},
	)

	database.Col(database.ColUsers).FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	utils.SuccessResponse(c, "Address updated", buildUserResponse(&user))
}

// DeleteUserAccount soft-deletes (sets status=INACTIVE) a user
func DeleteUserAccount(c *gin.Context) {
	userID := c.Param("userId")
	if userID != middleware.GetUserID(c) {
		utils.ForbiddenResponse(c, "Access denied")
		return
	}

	var req models.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID != userID {
		utils.BadRequestResponse(c, "Invalid request")
		return
	}

	database.Col(database.ColUsers).UpdateOne(context.Background(),
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"status": models.UserStatusInactive, "updated_at": time.Now()}},
	)
	utils.SuccessResponse(c, "Account deleted", nil)
}

// GetFamilyMembers returns family links for the logged-in patient
func GetFamilyMembers(c *gin.Context) {
	userID := middleware.GetUserID(c)

	cursor, err := database.Col(database.ColFamilyMembers).Find(
		context.Background(), bson.M{"patient_user_id": userID},
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch family members")
		return
	}
	var members []models.FamilyMember
	cursor.All(context.Background(), &members)
	utils.SuccessResponse(c, "Family members fetched", members)
}

// AddFamilyMember links a family user to a patient
func AddFamilyMember(c *gin.Context) {
	patientUserID := middleware.GetUserID(c)

	var req struct {
		FamilyUserID  string `json:"family_user_id"  binding:"required"`
		Relation      string `json:"relation"         binding:"required"`
		AlertsEnabled bool   `json:"alerts_enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	ctx := context.Background()
	// Check family user exists
	var fu models.User
	if err := database.Col(database.ColUsers).FindOne(ctx, bson.M{"_id": req.FamilyUserID}).Decode(&fu); err != nil {
		utils.NotFoundResponse(c, "Family user not found")
		return
	}

	member := models.FamilyMember{
		ID:            uuid.New().String(),
		PatientUserID: patientUserID,
		FamilyUserID:  req.FamilyUserID,
		Relation:      req.Relation,
		AccessLevel:   "VIEW",
		AlertsEnabled: req.AlertsEnabled,
		CreatedAt:     time.Now(),
	}
	database.Col(database.ColFamilyMembers).InsertOne(ctx, member)
	utils.CreatedResponse(c, "Family member added", member)
}

// GetNotifications returns paginated notifications for the user
func GetNotifications(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, limit := parsePage(c)

	ctx := context.Background()
	total, _ := database.Col(database.ColNotifications).CountDocuments(ctx, bson.M{"user_id": userID})

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, _ := database.Col(database.ColNotifications).Find(ctx, bson.M{"user_id": userID}, opts)
	var notifications []models.Notification
	cursor.All(ctx, &notifications)

	utils.PaginatedSuccessResponse(c, "Notifications fetched", notifications, page, limit, total)
}

// MarkNotificationRead marks one notification as read
func MarkNotificationRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	notifID := c.Param("notifId")

	res, _ := database.Col(database.ColNotifications).UpdateOne(
		context.Background(),
		bson.M{"_id": notifID, "user_id": userID},
		bson.M{"$set": bson.M{"is_read": true}},
	)
	if res.MatchedCount == 0 {
		utils.NotFoundResponse(c, "Notification not found")
		return
	}
	utils.SuccessResponse(c, "Notification marked as read", nil)
}

// MarkAllNotificationsRead marks all unread notifications as read
func MarkAllNotificationsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	database.Col(database.ColNotifications).UpdateMany(
		context.Background(),
		bson.M{"user_id": userID, "is_read": false},
		bson.M{"$set": bson.M{"is_read": true}},
	)
	utils.SuccessResponse(c, "All notifications marked as read", nil)
}

// ---- helpers ----

func buildUserResponse(u *models.User) models.UserResponse {
	resp := models.UserResponse{
		UserID:        u.ID,
		Name:          u.Name,
		Email:         u.Email,
		PhoneNumber:   u.PhoneNumber,
		Gender:        u.Gender,
		Role:          u.Role,
		Status:        u.Status,
		BlockStatus:   u.BlockStatus,
		UserService:   u.UserService,
		ServiceStatus: u.ServiceStatus,
		BloodGroup:    u.BloodGroup,
		ProfileImage:  u.ProfileImage,
	}
	for _, a := range u.Addresses {
		ar := &models.AddressResponse{
			HouseNumber: a.HouseNumber,
			Street:      a.Street,
			Landmark:    a.Landmark,
			City:        a.City,
			PinCode:     a.PinCode,
			Latitude:    a.Latitude,
			Longitude:   a.Longitude,
			IsPrimary:   a.IsPrimary,
		}
		if a.AddressType == "address1" {
			resp.Address1 = ar
		} else if a.AddressType == "address2" {
			resp.Address2 = ar
		}
	}
	return resp
}
