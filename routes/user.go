package routes

import (
	"encoding/json"
	"io"
	"jotno-server/models"
	"jotno-server/storage"
	"jotno-server/utils"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"
	jsonWT "github.com/kataras/iris/v12/middleware/jwt"
	"golang.org/x/crypto/bcrypt"
)

const baseImage = "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcT8whvraQ8GE5WRpAHd-7-v2m-rccRLF8BMPNG92HhmHB1T0yxxa4fPEPDvfXtYfew7FBE&usqp=CAU"

func Register(ctx iris.Context) {
	var userInput UserRegisterInput
	err := ctx.ReadJSON(&userInput)
	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}
	var newUser models.User
	userExists, userExistError := userExistsInDB(&newUser, userInput.Email)
	if userExistError != nil {
		utils.InternalServerError(ctx)
		return
	}
	if userExists {
		utils.EmailAlreadyRegistered(ctx)
		return
	}
	hashedPassword, hashErr := hashAndSaltPassword(userInput.Password)
	if hashErr != nil {
		utils.InternalServerError(ctx)
		return
	}
	newUser = models.User{
		FirstName:   userInput.FirstName,
		LastName:    userInput.LastName,
		Email:       strings.ToLower(userInput.Email),
		Password:    hashedPassword,
		SocialLogin: false,
		CountryCode: userInput.CountryCode,
		CallingCode: userInput.CallingCode,
		PhoneNumber: userInput.PhoneNumber,
		Avatar:      baseImage,
	}
	storage.DB.Create(&newUser)
	returnUser(newUser, ctx)
}

func Login(ctx iris.Context) {
	errorMsg := "Invalid email or password."
	var userInput UserLoginInput
	err := ctx.ReadJSON(&userInput)
	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}

	var user models.User
	userExists, userExistsError := userExistsInDB(&user, userInput.Email)

	if userExistsError != nil {
		utils.InternalServerError(ctx)
		return
	}
	if !userExists {
		utils.CreateError(iris.StatusUnauthorized,
			"Authentication Failure",
			errorMsg,
			ctx,
		)
		return
	}
	passwordError := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userInput.Password))
	if passwordError != nil {
		utils.CreateError(iris.StatusUnauthorized,
			"Authentication Failure",
			errorMsg,
			ctx,
		)
		return
	}
	returnUser(user, ctx)

}

func FacebookLoginOrSignUp(ctx iris.Context) {
	var userInput UserFacebookOrGoogleInput
	err := ctx.ReadJSON(&userInput)
	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}
	endpoint := "https://graph.facebook.com/me?fields=id,name,email&access_token=" + userInput.AccessToken
	client := &http.Client{}
	req, _ := http.NewRequest("GET", endpoint, nil)
	res, facebookErr := client.Do(req)
	if facebookErr != nil {
		utils.InternalServerError(ctx)
		return
	}

	defer res.Body.Close()
	body, bodyErr := io.ReadAll(res.Body)
	if bodyErr != nil {
		log.Panic(bodyErr)
		utils.InternalServerError(ctx)
		return
	}
	var facebookBody UserFacebookRes
	json.Unmarshal(body, &facebookBody)

	if facebookBody.Email != "" {
		var user models.User
		userExists, userExistsErr := userExistsInDB(&user, facebookBody.Email)

		if userExistsErr != nil {
			utils.InternalServerError(ctx)
			return
		}

		if !userExists {
			nameArr := strings.SplitN(facebookBody.Name, " ", 2)
			user = models.User{FirstName: nameArr[0],
				LastName:       nameArr[1],
				Email:          facebookBody.Email,
				SocialLogin:    true,
				SocialProvider: "Facebook",
				CallingCode:    facebookBody.CallingCode,
				CountryCode:    facebookBody.CountryCode,
				PhoneNumber:    facebookBody.PhoneNumber,
			}
			storage.DB.Create(&user)

			returnUser(user, ctx)
			return
		}

		if user.SocialLogin && user.SocialProvider == "Facebook" {
			returnUser(user, ctx)
			return
		}

		utils.EmailAlreadyRegistered(ctx)
		return
	}
}

func GoogleLoginOrSignUp(ctx iris.Context) {
	var userInput UserFacebookOrGoogleInput
	err := ctx.ReadJSON(&userInput)
	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}
	endpoint := "https://googleapis.com/userinfo/v2/me"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", endpoint, nil)
	header := "Bearer " + userInput.AccessToken
	req.Header.Set("Authorization", header)
	res, googleErr := client.Do(req)
	if googleErr != nil {
		utils.InternalServerError(ctx)
		return
	}

	defer res.Body.Close()
	body, bodyErr := io.ReadAll(res.Body)
	if bodyErr != nil {
		log.Panic(bodyErr)
		utils.InternalServerError(ctx)
		return
	}
	var googleBody UserGoogleRes
	json.Unmarshal(body, &googleBody)

	if googleBody.Email != "" {
		var user models.User
		userExists, userExistsErr := userExistsInDB(&user, googleBody.Email)

		if userExistsErr != nil {
			utils.InternalServerError(ctx)
			return
		}

		if !userExists {
			user = models.User{FirstName: googleBody.GivenName,
				LastName:       googleBody.FamilyName,
				Email:          googleBody.Email,
				SocialLogin:    true,
				SocialProvider: "Google",
				CallingCode:    googleBody.CallingCode,
				CountryCode:    googleBody.CountryCode,
				PhoneNumber:    googleBody.PhoneNumber,
			}
			storage.DB.Create(&user)

			returnUser(user, ctx)
			return
		}

		if user.SocialLogin && user.SocialProvider == "Google" {
			returnUser(user, ctx)
			return
		}

		utils.EmailAlreadyRegistered(ctx)
		return
	}
}

func ForgotPassword(ctx iris.Context) {
	var emailInput EmailRegisteredInput
	err := ctx.ReadJSON(&emailInput)
	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}

	var user models.User
	userExists, userExistsErr := userExistsInDB(&user, emailInput.Email)
	if userExistsErr != nil {
		utils.InternalServerError(ctx)
		return
	}
	if !userExists {
		utils.CreateError(iris.StatusUnauthorized, "Credentials Error", "Invalid email.", ctx)
		return
	}
	if userExists {
		if user.SocialLogin {
			utils.CreateError(iris.StatusUnauthorized, "Credentials Error", "Social Login Account", ctx)
			return
		}

		link := "exp://10.0.0.240:8081/--/screens/authentication/ResetPasswordScreen?token="
		token, tokenErr := utils.CreateForgotPasswordToken(user.ID, user.Email)

		if tokenErr != nil {
			utils.InternalServerError(ctx)
			return
		}

		link += token
		subject := "Forgot Your Password?"

		html := `
		<p>It looks like you forgot your password. 
		If you did, please click the link below to reset it. 		
		<br />Please update your password
		within 10 minutes, otherwise you will have to repeat this
		process. <a href=` + link + `>Click to Reset Password</a>
		</p><br />
		If you did not, disregard this email. <br />`

		emailSent, emailSentErr := utils.SendMail(user.Email, subject, html)
		if emailSentErr != nil {
			utils.InternalServerError(ctx)
			return
		}
		if emailSent {
			ctx.JSON(iris.Map{
				"emailSent": true,
			})
			return
		}
		ctx.JSON(iris.Map{"emailSent": false})
	}
}

func ResetPassword(ctx iris.Context) {
	var password ResetPasswordInput
	err := ctx.ReadJSON(&password)
	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}

	hashedPassword, hashErr := hashAndSaltPassword(password.Password)
	if hashErr != nil {
		utils.InternalServerError(ctx)
		return
	}
	claims := jsonWT.Get(ctx).(*utils.ForgotPasswordToken)

	var user models.User
	storage.DB.Model(&user).Where("id = ?", claims.ID).Update("password", hashedPassword)
	ctx.JSON(iris.Map{
		"passwordReset": true,
	})
}

func UpdateUserInformation(ctx iris.Context) {

	id := ctx.URLParam("id")

	const maxSize = 5 * iris.MB
	ctx.SetMaxRequestBodySize(maxSize)
	var userInput UserUpdateInput
	err := ctx.ReadJSON(&userInput)

	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}

	user := getUserByID(id, ctx)
	if user == nil {
		return
	}

	if userInput.FirstName != "" {
		user.FirstName = userInput.FirstName
	}

	if userInput.LastName != "" {
		user.LastName = userInput.LastName
	}

	if userInput.Email != "" {
		user.Email = userInput.Email
	}

	if userInput.Avatar != "" {
		defer func() {
			res := storage.UploadBase64Image(userInput.Avatar, strings.ReplaceAll(id+"/"+userInput.FirstName, " ", ""))
			user.Avatar = res["url"]
		}()
	}

	if userInput.Address != "" {
		user.Address = userInput.Address
	}

	if userInput.Lat != 0 {
		user.Lat = userInput.Lat
	}

	if userInput.Lon != 0 {
		user.Lon = userInput.Lon
	}

	if userInput.City != "" {
		user.City = userInput.City
	}

	rowsUpdated := storage.DB.Model(&user).Updates(user)

	if rowsUpdated.Error != nil {
		utils.InternalServerError(ctx)
		return
	}

	ctx.StatusCode(iris.StatusNoContent)
}

func GetUserFavoritedSpecialists(ctx iris.Context) {
	id := ctx.URLParam("id")
	user := getUserByID(id, ctx)
	if user == nil {
		return
	}

	var specialists []models.Specialist
	var favoritedSpecialists []uint
	unmarshalErr := json.Unmarshal(user.Favorited, &favoritedSpecialists)
	if unmarshalErr != nil {
		utils.InternalServerError(ctx)
		return
	}
	specialistsExists := storage.DB.Preload("Jobs").Where("id = ?", favoritedSpecialists).Find(&specialists)

	if specialistsExists.Error != nil {
		utils.InternalServerError(ctx)
		return
	}

	var specialistList []any
	for _, specialist := range specialists {
		specialistList = append(specialistList, specialistMap(specialist))
	}
	ctx.JSON(specialistList)
}

func AlterUserFavorites(ctx iris.Context) {
	id := ctx.URLParam("id")

	user := getUserByID(id, ctx)
	if user == nil {
		return
	}

	var alterFav AlterFavorites
	err := ctx.ReadJSON(&alterFav)
	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}

	specialistID := strconv.FormatUint(uint64(alterFav.SpecialistID), 10)
	validSpecialistID := getSpecialistAndAssociationsByID(specialistID, ctx)

	if validSpecialistID == nil {
		return
	}
	var favoritedSpecialists []uint
	var unMarshalledFavorites []uint

	if user.Favorited != nil {
		unmarshalErr := json.Unmarshal(user.Favorited, &unMarshalledFavorites)

		if unmarshalErr != nil {
			utils.InternalServerError(ctx)
			return
		}
	}

	if alterFav.Op == "add" {
		if !slices.Contains(unMarshalledFavorites, alterFav.SpecialistID) {
			favoritedSpecialists = append(unMarshalledFavorites, alterFav.SpecialistID)
		} else {
			favoritedSpecialists = unMarshalledFavorites
		}
	} else if alterFav.Op == "remove" && len(unMarshalledFavorites) > 0 {
		for _, specialistID := range unMarshalledFavorites {
			if alterFav.SpecialistID != specialistID {
				favoritedSpecialists = append(favoritedSpecialists, specialistID)
			}
		}
	}

	marshalledSpecialists, marshalErr := json.Marshal(favoritedSpecialists)

	if marshalErr != nil {
		utils.InternalServerError(ctx)
		return
	}

	user.Favorited = marshalledSpecialists

	rowsUpdated := storage.DB.Model(&user).Updates(user)

	if rowsUpdated.Error != nil {
		utils.InternalServerError(ctx)
		return
	}

	ctx.StatusCode(iris.StatusNoContent)
}

func AlterPushToken(ctx iris.Context) {
	id := ctx.URLParam("id")

	user := getUserByID(id, ctx)
	if user == nil {
		return
	}

	var req AlterPushTokenInput
	err := ctx.ReadJSON(&req)
	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}

	var unMarshalledTokens []string
	var pushTokens []string

	if user.PushTokens != nil {
		unmarshalErr := json.Unmarshal(user.PushTokens, &unMarshalledTokens)

		if unmarshalErr != nil {
			utils.InternalServerError(ctx)
			return
		}
	}

	if req.Op == "add" {
		if !slices.Contains(unMarshalledTokens, req.Token) {
			pushTokens = append(unMarshalledTokens, req.Token)
		} else {
			pushTokens = unMarshalledTokens
		}
	} else if req.Op == "remove" && len(unMarshalledTokens) > 0 {
		for _, token := range unMarshalledTokens {
			if req.Token != token {
				pushTokens = append(pushTokens, token)
			}
		}
	}

	marshalledTokens, marshalErr := json.Marshal(pushTokens)
	if marshalErr != nil {
		utils.InternalServerError(ctx)
		return
	}

	user.PushTokens = marshalledTokens
	rowsUpdated := storage.DB.Model(&user).Updates(user)
	if rowsUpdated.Error != nil {
		utils.InternalServerError(ctx)
		return
	}
	ctx.StatusCode(iris.StatusNoContent)
}

func AllowsNotifications(ctx iris.Context) {
	id := ctx.URLParam("id")

	var req AllowsNotificationsInput
	err := ctx.ReadJSON(&req)
	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}

	user := getUserByID(id, ctx)
	if user == nil {
		return
	}

	user.AllowsNotifications = req.AllowsNotifications

	rowsUpdated := storage.DB.Model(&user).Updates(user)

	if rowsUpdated.Error != nil {
		utils.InternalServerError(ctx)
		return
	}
	ctx.StatusCode(iris.StatusNoContent)
}

func getUserByID(id string, ctx iris.Context) *models.User {
	var user models.User
	userExists := storage.DB.Where("id = ?", id).Find(&user)

	if userExists.Error != nil {
		utils.InternalServerError(ctx)
		return nil
	}
	if userExists.RowsAffected == 0 {
		utils.CreateNotFound(ctx)
		return nil
	}
	return &user
}

func hashAndSaltPassword(password string) (hashedPassword string, err error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func userExistsInDB(user *models.User, email string) (exist bool, err error) {
	userExistQuery := storage.DB.Where("email = ?", strings.ToLower(email)).Limit(1).Find(&user)
	if userExistQuery.Error != nil {
		return false, userExistQuery.Error
	}
	userExists := userExistQuery.RowsAffected > 0
	if userExists {
		return true, nil
	}

	return false, nil
}

func returnUser(user models.User, ctx iris.Context) {
	tokenPair, tokenErr := utils.CreateTokenPair(user.ID)
	if tokenErr != nil {
		utils.InternalServerError(ctx)
		return
	}

	ctx.JSON(iris.Map{
		"ID":                  user.ID,
		"firstName":           user.FirstName,
		"lastName":            user.LastName,
		"email":               user.Email,
		"countryCode":         user.CountryCode,
		"callingCode":         user.CallingCode,
		"phoneNumber":         user.PhoneNumber,
		"address":             user.Address,
		"city":                user.City,
		"lat":                 user.Lat,
		"lon":                 user.Lon,
		"avatar":              user.Avatar,
		"favorited":           user.Favorited,
		"allowsNotifications": user.AllowsNotifications,
		"accessToken":         string(tokenPair.AccessToken),
		"refreshToken":        string(tokenPair.RefreshToken),
	})
}

type UserRegisterInput struct {
	FirstName   string `json:"firstName" validate:"required,max=256"`
	LastName    string `json:"lastName" validate:"required,max=256"`
	Email       string `json:"email" validate:"required,max=256,email"`
	Password    string `json:"password" validate:"required,min=8,max=256"`
	CallingCode string `json:"callingCode" validate:"required"`
	CountryCode string `json:"countryCode" validate:"required"`
	PhoneNumber string `json:"phoneNumber" validate:"required"`
}

type UserLoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserFacebookRes struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	CallingCode string `json:"callingCode" validate:"required"`
	CountryCode string `json:"countryCode" validate:"required"`
	PhoneNumber string `json:"phoneNumber" validate:"required"`
}

type UserFacebookOrGoogleInput struct {
	AccessToken string `json:"accessToken" validate:"required"`
}

type UserGoogleRes struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	GivenName   string `json:"given_name"`
	FamilyName  string `json:"family_name"`
	CallingCode string `json:"callingCode" validate:"required"`
	CountryCode string `json:"countryCode" validate:"required"`
	PhoneNumber string `json:"phoneNumber" validate:"required"`
}

type EmailRegisteredInput struct {
	Email string `json:"email" validate:"required"`
}

type ResetPasswordInput struct {
	Password string `json:"password" validate:"required,min=8,max=256"`
}

type UserUpdateInput struct {
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     string  `json:"email"`
	Avatar    string  `json:"avatar"`
	Address   string  `json:"address"`
	Lat       float32 `json:"lat"`
	Lon       float32 `json:"lon"`
	City      string  `json:"city"`
}

type AlterFavorites struct {
	SpecialistID uint   `json:"specialistID" validate:"required"`
	Op           string `json:"op" validate:"required"`
}

type AlterPushTokenInput struct {
	Token string `json:"token" validate:"required"`
	Op    string `json:"op" validate:"required"`
}

type AllowsNotificationsInput struct {
	AllowsNotifications *bool `json:"allowsNotifications" validate:"required"`
}
