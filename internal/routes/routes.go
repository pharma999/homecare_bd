package routes

import (
	"home_care_backend/internal/handlers"
	"home_care_backend/internal/middleware"
	"home_care_backend/internal/models"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) {
	api := r.Group("/api")

	// ── WebSocket (token auth via query param) ───────────────────────────────
	api.GET("/ws", handlers.ConnectWS)

	// ── Public ──────────────────────────────────────────────────────────────
	auth := api.Group("/auth")
	{
		auth.POST("/send-otp", middleware.OTPRateLimit(), handlers.Login)
		auth.POST("/verify-otp", middleware.OTPRateLimit(), handlers.VerifyOTP)
	}
	// Legacy aliases for backward compatibility
	api.POST("/login", middleware.OTPRateLimit(), handlers.Login)
	api.POST("/verify", middleware.OTPRateLimit(), handlers.VerifyOTP)

	// ── Authenticated ────────────────────────────────────────────────────────
	protected := api.Group("")
	protected.Use(middleware.AuthRequired())
	{
		// Users
		user := protected.Group("/user")
		{
			user.GET("/:userId", handlers.GetUserProfile)
			user.POST("/:userId/update", handlers.UpdateUserProfile)
			user.POST("/:userId/address/update", handlers.UpdateUserAddress)
			user.POST("/:userId/delete", handlers.DeleteUserAccount)
		}

		// Family
		family := protected.Group("/family")
		{
			family.GET("/members", handlers.GetFamilyMembers)
			family.POST("/members", handlers.AddFamilyMember)
			family.GET("/patient/:patientId/appointments", handlers.GetFamilyAppointments)
			family.GET("/patient/:patientId/records", handlers.GetFamilyMedicalRecords)
		}

		// Notifications
		notif := protected.Group("/notifications")
		{
			notif.GET("", handlers.GetNotifications)
			notif.PATCH("/:notifId/read", handlers.MarkNotificationRead)
			notif.PATCH("/read-all", handlers.MarkAllNotificationsRead)
		}

		// Services & Professionals
		services := protected.Group("/services")
		{
			services.GET("", handlers.GetAllServices)
			services.GET("/:serviceId", handlers.GetService)
		}
		categories := protected.Group("/categories")
		{
			categories.GET("", handlers.GetServiceCategories)
			categories.GET("/:categoryId/services", handlers.GetCategoryServices)
		}
		professionals := protected.Group("/professionals")
		{
			professionals.GET("", handlers.GetProfessionals)
			professionals.GET("/:professionalId", handlers.GetProfessional)
		}

		// Bookings
		bookings := protected.Group("/bookings")
		{
			bookings.POST("", handlers.CreateBooking)
			bookings.GET("", handlers.GetMyBookings)
			bookings.GET("/:bookingId", handlers.GetBooking)
			bookings.POST("/:bookingId/cancel", handlers.CancelBooking)
		}

		// Cart
		cart := protected.Group("/cart")
		{
			cart.GET("", handlers.GetCart)
			cart.POST("/add", handlers.AddToCart)
			cart.POST("/update-quantity", handlers.UpdateCartQuantity)
			cart.DELETE("/:serviceId", handlers.RemoveFromCart)
			cart.DELETE("", handlers.ClearCart)
			cart.POST("/checkout", handlers.CheckoutCart)
		}

		// Appointments
		appointments := protected.Group("/appointments")
		{
			appointments.POST("", handlers.CreateAppointment)
			appointments.GET("", handlers.GetMyAppointments)
			appointments.GET("/:appointmentId", handlers.GetAppointment)
			appointments.PATCH("/:appointmentId/status", handlers.UpdateAppointmentStatus)
			appointments.POST("/:appointmentId/cancel", handlers.CancelAppointment)
		}

		// Prescriptions
		protected.GET("/prescriptions", handlers.GetPrescriptions)

		// Payments
		payments := protected.Group("/payments")
		{
			payments.POST("/initiate", handlers.InitiatePayment)
			payments.POST("/:paymentId/confirm", handlers.ConfirmPayment)
			payments.GET("/history", handlers.GetPaymentHistory)
		}

		// Emergency / SOS
		emergency := protected.Group("/emergency")
		{
			emergency.POST("/sos", handlers.TriggerSOS)
			emergency.GET("", handlers.GetMyEmergencies)
			emergency.GET("/:emergencyId", handlers.GetEmergency)
		}

		// Medical Records
		records := protected.Group("/medical-records")
		{
			records.POST("", handlers.UploadMedicalRecord)
			records.GET("", handlers.GetMyMedicalRecords)
			records.DELETE("/:recordId", handlers.DeleteMedicalRecord)
		}

		// Reviews
		protected.POST("/reviews", handlers.SubmitReview)

		// Chatbot
		chatbot := protected.Group("/chatbot")
		{
			chatbot.POST("/message", handlers.SendChatMessage)
			chatbot.GET("/history", handlers.GetChatHistory)
		}

		// Support tickets (any authenticated user)
		support := protected.Group("/support")
		{
			support.POST("", handlers.CreateSupportTicket)
			support.GET("", handlers.GetMyTickets)
		}

		// Hospitals (public read, authenticated write)
		hospitals := protected.Group("/hospitals")
		{
			hospitals.POST("", handlers.RegisterHospital)
			hospitals.GET("", handlers.GetNearbyHospitals)
			hospitals.GET("/:hospitalId", handlers.GetHospital)
			hospitals.PUT("/:hospitalId", handlers.UpdateHospital)
			hospitals.GET("/:hospitalId/doctors", handlers.GetHospitalDoctors)
			hospitals.POST("/:hospitalId/ambulances", handlers.AddAmbulance)
			hospitals.GET("/:hospitalId/emergencies", handlers.GetHospitalEmergencies)
		}

		// Ambulance location update
		protected.PATCH("/ambulances/:ambulanceId/location", handlers.UpdateAmbulanceLocation)

		// Doctors
		doctors := protected.Group("/doctors")
		{
			doctors.POST("", handlers.RegisterDoctor)
			doctors.GET("", handlers.SearchDoctors)
			doctors.GET("/:doctorId", handlers.GetDoctorProfile)
			doctors.PUT("/:doctorId", handlers.UpdateDoctorProfile)
			doctors.PATCH("/:doctorId/availability", handlers.SetDoctorAvailability)
			doctors.GET("/:doctorId/schedule", handlers.GetDoctorSchedule)
			doctors.PUT("/:doctorId/schedule", handlers.SetDoctorSchedule)
			doctors.GET("/me/appointments", handlers.GetMyDoctorAppointments)
			doctors.GET("/me/earnings", handlers.GetDoctorEarnings)
			doctors.POST("/prescriptions", handlers.UploadPrescription)
		}

		// Nurses
		nurses := protected.Group("/nurses")
		{
			nurses.POST("", handlers.RegisterNurse)
			nurses.GET("", handlers.SearchNurses)
			nurses.GET("/:nurseId", handlers.GetNurseProfile)
			nurses.PATCH("/:nurseId/availability", handlers.UpdateNurseAvailability)
			nurses.GET("/me/earnings", handlers.GetNurseEarnings)
		}
	}

	// ── Admin ─────────────────────────────────────────────────────────────────
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(), middleware.RoleRequired(
		string(models.RoleAdmin), string(models.RoleSuperAdmin),
	))
	{
		admin.GET("/analytics", handlers.AdminAnalytics)

		admin.GET("/users", handlers.AdminListUsers)
		admin.GET("/users/:userId", handlers.AdminGetUser)
		admin.PATCH("/users/:userId/block", handlers.AdminBlockUser)

		admin.GET("/hospitals", handlers.AdminListHospitals)
		admin.PATCH("/hospitals/:hospitalId/approval", handlers.AdminApproveHospital)

		admin.GET("/doctors", handlers.AdminListDoctors)
		admin.PATCH("/doctors/:doctorId/approval", handlers.AdminApproveDoctor)

		admin.GET("/nurses", handlers.AdminListNurses)
		admin.PATCH("/nurses/:nurseId/approval", handlers.AdminApproveNurse)

		admin.POST("/services", handlers.AdminCreateService)
		admin.PUT("/services/:serviceId", handlers.AdminUpdateService)
		admin.POST("/categories", handlers.AdminCreateCategory)

		admin.GET("/bookings", handlers.AdminListBookings)
		admin.PATCH("/bookings/:bookingId/status", handlers.AdminUpdateBookingStatus)

		admin.GET("/emergencies/active", handlers.GetActiveEmergencies)
		admin.PATCH("/emergencies/:emergencyId/status", handlers.UpdateEmergencyStatus)

		admin.GET("/appointments", handlers.AdminListAppointments)

		// Support tickets
		admin.GET("/support", handlers.AdminListSupportTickets)
		admin.GET("/support/:ticketId", handlers.AdminGetSupportTicket)
		admin.PATCH("/support/:ticketId", handlers.AdminUpdateSupportTicket)

		// Subscription plans
		admin.GET("/plans", handlers.AdminListSubscriptionPlans)
		admin.POST("/plans", handlers.AdminCreateSubscriptionPlan)
		admin.PUT("/plans/:planId", handlers.AdminUpdateSubscriptionPlan)
		admin.DELETE("/plans/:planId", handlers.AdminDeleteSubscriptionPlan)
		admin.GET("/subscriptions", handlers.AdminListUserSubscriptions)

		// Service zones
		admin.GET("/zones", handlers.AdminListZones)
		admin.POST("/zones", handlers.AdminCreateZone)
		admin.PUT("/zones/:zoneId", handlers.AdminUpdateZone)
	}

	// ── Super Admin only ──────────────────────────────────────────────────────
	superAdmin := api.Group("/super-admin")
	superAdmin.Use(middleware.AuthRequired(), middleware.RoleRequired(
		string(models.RoleSuperAdmin),
	))
	{
		superAdmin.GET("/revenue", handlers.SuperAdminRevenueReport)
		superAdmin.GET("/admins", handlers.SuperAdminListAdmins)
		superAdmin.POST("/admins", handlers.SuperAdminCreateAdmin)
		superAdmin.DELETE("/admins/:adminId", handlers.SuperAdminDeleteAdmin)
		superAdmin.GET("/settings", handlers.GetPlatformSettings)
		superAdmin.PUT("/settings", handlers.UpdatePlatformSettings)
	}
}
