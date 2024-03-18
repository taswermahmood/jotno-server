package routes

import (
	"jotno-server/models"
	"jotno-server/storage"
	"jotno-server/utils"
	"strings"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm/clause"
)

func RegisterSpecialist(ctx iris.Context) {
	var specialistInput SpecialistSignUpInput
	err := ctx.ReadJSON(&specialistInput)
	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}
	var newSpecialist models.Specialist
	userExists, userExistError := specialistExistsInDB(&newSpecialist, specialistInput.Email)
	if userExistError != nil {
		utils.InternalServerError(ctx)
		return
	}
	if userExists {
		utils.EmailAlreadyRegistered(ctx)
		return
	}
	hashedPassword, hashErr := hashAndSaltPassword(specialistInput.Password)
	if hashErr != nil {
		utils.InternalServerError(ctx)
		return
	}

	newSpecialist = models.Specialist{
		FirstName:   specialistInput.FirstName,
		LastName:    specialistInput.LastName,
		Email:       strings.ToLower(specialistInput.Email),
		Password:    hashedPassword,
		PhoneNumber: specialistInput.PhoneNumber,
		Avatar:      baseImage,
		Address:     specialistInput.Address,
		Lat:         specialistInput.Lat,
		Lon:         specialistInput.Lon,
	}
	storage.DB.Create(&newSpecialist)
	returnSpecialist(newSpecialist, ctx)
}

func GetSpecialistByID(ctx iris.Context) {
	id := ctx.URLParam("specialistId")

	specialist := getSpecialistAndAssociationsByID(id, ctx)
	if specialist == nil {
		return
	}
	returnSpecialist(*specialist, ctx)
}

func GetSpecialistByIDAndJobName(ctx iris.Context) {
	id := ctx.URLParam("specialistId")
	jobName := ctx.URLParam("jobName")

	var specialist models.Specialist
	specialistExists := storage.DB.Preload("Jobs", "job_name = ? ", jobName).Preload("Reviews").Preload("Posts").Where("id = ?", id).Find(&specialist)

	if specialistExists.Error != nil {
		utils.InternalServerError(ctx)
		return
	}
	if specialistExists.RowsAffected == 0 {
		utils.CreateNotFound(ctx)
		return
	}
	returnSpecialist(specialist, ctx)
}

func GetSpecialistByBoundingBox(ctx iris.Context) {
	var boundingBox BoundingBoxInput
	err := ctx.ReadJSON(&boundingBox)
	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}
	var specialists []models.Specialist

	subQuery := storage.DB.Select("specialist_id").Where("job_name = ?", boundingBox.JobName).Table("jobs")
	storage.DB.Preload("Jobs", "job_name = ? ", boundingBox.JobName).
		Where("id IN (?) AND lat >= ? AND lat <= ? AND lon >= ? AND lon <= ?", subQuery, boundingBox.LatLow, boundingBox.LatHigh, boundingBox.LonLow, boundingBox.LonHigh).
		Find(&specialists)

	var specialistList []any
	for _, specialist := range specialists {
		specialistList = append(specialistList, specialistMap(specialist))
	}
	ctx.JSON(specialistList)

}

func specialistExistsInDB(user *models.Specialist, email string) (exist bool, err error) {
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

func getSpecialistAndAssociationsByID(id string, ctx iris.Context) *models.Specialist {

	var specialist models.Specialist
	specialistExists := storage.DB.Preload(clause.Associations).Find(&specialist, id)

	if specialistExists.Error != nil {
		utils.InternalServerError(ctx)
		return nil
	}

	if specialistExists.RowsAffected == 0 {
		utils.CreateNotFound(ctx)
		return nil
	}

	return &specialist
}

type SpecialistSignUpInput struct {
	FirstName   string  `json:"firstName" validate:"required,max=256"`
	LastName    string  `json:"lastName" validate:"required,max=256"`
	Email       string  `json:"email" validate:"required,max=256,email"`
	Password    string  `json:"password" validate:"required,min=8,max=256"`
	PhoneNumber string  `json:"phoneNumber" validate:"required"`
	Address     string  `json:"address" validate:"required"`
	Lat         float32 `json:"lat" validate:"required"`
	Lon         float32 `json:"lon" validate:"required"`
}

type BoundingBoxInput struct {
	JobName string  `json:"jobName" validate:"required"`
	LatLow  float32 `json:"latLow" validate:"required"`
	LatHigh float32 `json:"latHigh" validate:"required"`
	LonLow  float32 `json:"lonLow" validate:"required"`
	LonHigh float32 `json:"lonHigh" validate:"required"`
}

func returnSpecialist(user models.Specialist, ctx iris.Context) {
	ctx.JSON(specialistMap(user))
}

func specialistMap(user models.Specialist) iris.Map {
	return iris.Map{
		"ID":          user.ID,
		"firstName":   user.FirstName,
		"lastName":    user.LastName,
		"email":       user.Email,
		"countryCode": user.CountryCode,
		"callingCode": user.CallingCode,
		"phoneNumber": user.PhoneNumber,
		"avatar":      user.Avatar,
		"images":      user.Images,
		"idCard":      user.IdCard,
		"address":     user.Address,
		"city":        user.City,
		"lat":         user.Lat,
		"lon":         user.Lon,
		"experience":  user.Experience,
		"stars":       user.Stars,
		"about":       user.About,
		"verified":    user.Verified,
		"jobs":        user.Jobs,
		"reviews":     user.Reviews,
		"posts":       user.Posts,
	}
}
