package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"home_care_backend/internal/config"
	"home_care_backend/internal/database"
	"home_care_backend/internal/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	config.Load()
	database.Connect()
	defer database.Disconnect()

	superAdmins := []struct {
		Name  string
		Phone string
	}{
		{"Super Admin 1", "7571928881"},
		{"Super Admin 2", "6386098744"},
		{"Super Admin 3", "8368701991"},
	}

	col := database.Col(database.ColUsers)
	ctx := context.Background()

	for _, sa := range superAdmins {
		// Check if already exists
		var existing models.User
		err := col.FindOne(ctx, bson.M{"phone_number": sa.Phone}).Decode(&existing)
		if err == nil {
			// User exists — just update role to SUPER_ADMIN
			col.UpdateOne(ctx,
				bson.M{"phone_number": sa.Phone},
				bson.M{"$set": bson.M{
					"role":       string(models.RoleSuperAdmin),
					"status":     string(models.Active),
					"block_status": string(models.Unblocked),
					"updated_at": time.Now(),
				}},
			)
			fmt.Printf("✅ Updated existing user %s (%s) → SUPER_ADMIN\n", sa.Name, sa.Phone)
			continue
		}
		if err != mongo.ErrNoDocuments {
			log.Printf("❌ DB error for %s: %v", sa.Phone, err)
			continue
		}

		// Insert new super admin
		now := time.Now()
		user := models.User{
			ID:            uuid.New().String(),
			Name:          sa.Name,
			PhoneNumber:   sa.Phone,
			Role:          models.RoleSuperAdmin,
			Status:        models.Active,
			BlockStatus:   models.Unblocked,
			UserService:   models.Unsubscribed,
			ServiceStatus: models.New,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		_, err = col.InsertOne(ctx, user,
			options.InsertOne(),
		)
		if err != nil {
			log.Printf("❌ Failed to insert %s (%s): %v", sa.Name, sa.Phone, err)
			continue
		}
		fmt.Printf("✅ Created SUPER_ADMIN: %s (%s)\n", sa.Name, sa.Phone)
	}

	fmt.Println("\nDone. All 3 super admins can now log in via OTP.")
}
