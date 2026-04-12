package database

import (
	"context"
	"log"
	"time"

	"home_care_backend/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DB *mongo.Database

// Collection names
const (
	ColUsers          = "users"
	ColOTPs           = "otps"
	ColDoctors        = "doctors"
	ColNurses         = "nurses"
	ColHospitals      = "hospitals"
	ColAmbulances     = "ambulances"
	ColServices       = "services"
	ColCategories     = "service_categories"
	ColProfessionals  = "professionals"
	ColAppointments   = "appointments"
	ColBookings       = "bookings"
	ColCartItems      = "cart_items"
	ColEmergencies    = "emergencies"
	ColMedicalRecords = "medical_records"
	ColPrescriptions  = "prescriptions"
	ColPayments       = "payments"
	ColReviews        = "reviews"
	ColNotifications  = "notifications"
	ColFamilyMembers      = "family_members"
	ColSubscriptionPlans  = "subscription_plans"
	ColUserSubscriptions  = "user_subscriptions"
	ColSupportTickets     = "support_tickets"
	ColPlatformSettings   = "platform_settings"
	ColServiceZones       = "service_zones"
)

func Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.AppConfig.MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("MongoDB connection established")
	Client = client
	DB = client.Database(config.AppConfig.MongoDBName)

	createIndexes()
}

func createIndexes() {
	ctx := context.Background()

	// Users: unique phone_number
	Col(ColUsers).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "phone_number", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	// Users: email sparse unique
	Col(ColUsers).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true),
	})

	// OTPs: TTL index — auto-delete after expiry
	Col(ColOTPs).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	Col(ColOTPs).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "phone_number", Value: 1}},
	})

	// Doctors: unique license_number
	Col(ColDoctors).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "license_number", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	Col(ColDoctors).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "specialty", Value: 1}},
	})
	Col(ColDoctors).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "approval_status", Value: 1}},
	})

	// Hospitals: unique registration_number + 2dsphere geo index
	Col(ColHospitals).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "registration_number", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	Col(ColHospitals).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}},
	})

	// Professionals: 2dsphere geo index
	Col(ColProfessionals).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}},
	})

	// Doctors: specialty search + 2dsphere
	Col(ColDoctors).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}},
	})

	// Ambulances: 2dsphere for real-time tracking
	Col(ColAmbulances).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}},
	})
	Col(ColAmbulances).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "emergency_id", Value: 1}},
	})

	// Bookings & Appointments
	Col(ColBookings).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "patient_user_id", Value: 1}, {Key: "status", Value: 1}},
	})
	Col(ColAppointments).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "patient_user_id", Value: 1}, {Key: "status", Value: 1}},
	})

	// Emergencies: status + patient geo
	Col(ColEmergencies).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}},
	})

	// Notifications
	Col(ColNotifications).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "is_read", Value: 1}},
	})

	// Cart Items: unique per user+service
	Col(ColCartItems).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "service_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// Support tickets
	Col(ColSupportTickets).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "status", Value: 1}},
	})

	// User subscriptions
	Col(ColUserSubscriptions).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "status", Value: 1}},
	})

	// Service zones
	Col(ColServiceZones).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "city", Value: 1}, {Key: "status", Value: 1}},
	})

	log.Println("MongoDB indexes created")
}

// Col returns a collection from the active database
func Col(name string) *mongo.Collection {
	return DB.Collection(name)
}

// Disconnect closes the MongoDB connection
func Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if Client != nil {
		Client.Disconnect(ctx)
		log.Println("MongoDB disconnected")
	}
}
