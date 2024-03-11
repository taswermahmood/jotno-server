package main

import (
	"jotno-server/routes"
	"jotno-server/storage"
	"jotno-server/utils"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
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

		user.Get("/{id}/specialist/favorited", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetUserFavoritedSpecialists)
		user.Patch("/{id}/updateUserInformation", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.UpdateUserInformation)
		user.Patch("/{id}/specialist/favorited", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.AlterUserFavorites)
		user.Patch("/{id}/pushToken", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.AlterPushToken)
		user.Patch("/{id}/settings/notifications", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.AllowsNotifications)
	}

	specialist := app.Party("/jotno/api/specialist")
	{
		specialist.Post("/register", routes.RegisterSpecialist)

		specialist.Get("/{specialistId}/user/{id}", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetSpecialistByID)
		specialist.Get("/{specialistId}/{jobName}/user/{id}", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetSpecialistByIDAndJobName)
		specialist.Post("/search/user/{id}", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetSpecialistByBoundingBox)
	}

	jobPost := app.Party("/jotno/api/jobPost")
	{
		jobPost.Get("/getJobPosts/user/{id}", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetJobsPostsByUserID)
		jobPost.Post("/createJobPosts/user/{id}", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.CreateJobPosts)
		jobPost.Delete("/jobPost/{jobId}/user/{id}", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.DeleteJobPost)
	}

	notification := app.Party("/jotno/api/notification")
	{
		notification.Post("/sendNotification", routes.TestMessageNotification)
	}

	chat := app.Party("/jotno/api/chat")
	{
		chat.Post("/create/user/{id}", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.CreateChat)
		chat.Post("/open/user/{id}", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetChatByUserAndSpecialistID)
		chat.Get("/{chatId}/user/{id}", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetChatByID)
		chat.Get("/user/{id}", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.GetChatsByUserID)
	}

	messages := app.Party("/jotno/api/messages")
	{
		messages.Post("/create/user/{id}", accessTokenVerifierMiddleware, utils.UserIDMiddleware, routes.CreateMessage)
	}

	app.Listen(":4000")
}
