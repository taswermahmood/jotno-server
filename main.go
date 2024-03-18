package main

import (
	"jotno-server/models"
	"jotno-server/routes"
	"jotno-server/storage"
	"jotno-server/utils"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
	"github.com/madflojo/tasks"
)

func main() {
	// Initiation
	godotenv.Load()
	storage.InitializeDB()
	storage.InitializeS3()
	storage.InitializeRedis()

	app := iris.Default()
	app.Validator = validator.New()

	// Reset token verifiers
	resetTokenVerifier := jwt.NewVerifier(jwt.HS256, []byte(os.Getenv("EMAIL_TOKEN_SECRET")))
	resetTokenVerifier.WithDefaultBlocklist()
	resetTokenVerifierMiddleware := resetTokenVerifier.Verify(func() interface{} {
		return new(utils.ForgotPasswordToken)
	})
	// JWT token verifiers
	accessTokenVerifier := jwt.NewVerifier(jwt.HS256, []byte(os.Getenv("ACCESS_TOKEN_SECRET")))
	accessTokenVerifier.WithDefaultBlocklist()
	accessTokenVerifierMiddleware := accessTokenVerifier.Verify(func() interface{} {
		return new(utils.AccessToken)
	})
	// JWT reset token verifiers
	refreshTokenVerifier := jwt.NewVerifier(jwt.HS256, []byte(os.Getenv("REFRESH_TOKEN_SECRET")))
	refreshTokenVerifier.WithDefaultBlocklist()
	refreshTokenVerifierMiddleware := refreshTokenVerifier.Verify(func() interface{} {
		return new(jwt.Claims)
	})

	refreshTokenVerifier.Extractors = append(refreshTokenVerifier.Extractors, func(ctx iris.Context) string {
		var tokenInput utils.RefreshTokenInput
		err := ctx.ReadJSON(&tokenInput)
		if err != nil {
			return ""
		}
		return tokenInput.RefreshToken
	})

	app.Post("/jotno/api/refresh", refreshTokenVerifierMiddleware, utils.RefreshToken)

	location := app.Party("/jotno/api/location")
	{
		location.Get("/autocomplete", routes.Autocomplete)
		location.Get("/search", routes.Search)
	}

	user := app.Party("/jotno/api/user")
	{
		user.Post("/register", routes.Register)
		user.Post("/login", routes.Login)
		user.Post("/facebook", routes.FacebookLoginOrSignUp)
		user.Post("/google", routes.GoogleLoginOrSignUp)
		user.Post("/forgotPassword", routes.ForgotPassword)
		user.Post("/resetPassword", resetTokenVerifierMiddleware, routes.ResetPassword)

		user.Get("/specialist/favorited", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetUserFavoritedSpecialists)
		user.Patch("/updateUserInformation", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.UpdateUserInformation)
		user.Patch("/specialist/favorited", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.AlterUserFavorites)
		user.Patch("/pushToken", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.AlterPushToken)
		user.Patch("/settings/notifications", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.AllowsNotifications)
	}

	specialist := app.Party("/jotno/api/specialist")
	{
		specialist.Post("/register", routes.RegisterSpecialist)
		// specialist.Get("/{specialistId}/user", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetSpecialistByID)
		specialist.Get("/getSpecialist", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetSpecialistByIDAndJobName)
		specialist.Post("/search", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetSpecialistByBoundingBox)
	}

	jobPost := app.Party("/jotno/api/jobPost")
	{
		jobPost.Get("/getJobPosts", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetJobsPostsByUserID)
		jobPost.Post("/createJobPosts", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.CreateJobPosts)
		jobPost.Delete("/deleteJobPost", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.DeleteJobPost)
	}

	// notification := app.Party("/jotno/api/notification")
	// {
	// 	notification.Post("/sendNotification", routes.SendNotification)
	// }

	chat := app.Party("/jotno/api/chat")
	{
		chat.Post("/create", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.CreateChat)
		chat.Post("/open", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetChatByUserAndSpecialistID)
		chat.Get("/getChat", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetChatByID)
		chat.Get("/getChats", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetChatsByUserID)
	}

	messages := app.Party("/jotno/api/messages")
	{
		messages.Post("/create", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.CreateMessage)
	}

	booking := app.Party("/jotno/api/booking")
	{
		booking.Get("/getBookingByUser", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetBookingByUserID)
		booking.Post("/create", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.CreateBooking)
		booking.Patch("/cancelBooking", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.CancelBooking)
		booking.Get("/getPendingPaymentsByBookingID", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetPendingPaymentsByBookingID)
		// booking.Patch("/updateBooking", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.DeleteJobPost)
		// booking.Patch("/updatePayment", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.DeleteJobPost)
	}

	scheduler := tasks.New()
	defer scheduler.Stop()

	scheduler.Add(&tasks.Task{
		Interval: (24 * time.Hour),
		TaskFunc: func() error {
			var activeBookings []models.Booking
			activeBookingsExists:= storage.DB.Where("active = true AND frequency = 'monthly' AND overdue = false").Find(&activeBookings)
			if activeBookingsExists.Error != nil {
				return nil
			}
			for i := 0; i < len(activeBookings); i++ {
				routes.CreateBill(
					activeBookings[i].ID,
					activeBookings[i].Amount,
					activeBookings[i].Currency,
					activeBookings[i].UserID,
				) 
			}
			return nil
		},
	})

	app.Listen(":4000")
}
