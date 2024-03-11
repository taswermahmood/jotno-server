package routes

import (
	"jotno-server/models"
	"jotno-server/storage"
	"jotno-server/utils"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm/clause"
)

func CreateJobPosts(ctx iris.Context) {
	var jobPostInput CreateJobPostInput

	err := ctx.ReadJSON(&jobPostInput)
	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}
	jobPost := models.JobPost{
		UserID:        jobPostInput.UserID,
		JobType:       jobPostInput.JobType,
		Title:         jobPostInput.Title,
		Description:   jobPostInput.Description,
		Wage:          jobPostInput.Wage,
		WageCurrency:  "BDT",
		WageFrequency: jobPostInput.WageFrequency,
		DateTime:      jobPostInput.DateTime,
	}
	storage.DB.Create(&jobPost)
	ctx.JSON(jobPost)
}

func GetJobsPostsByUserID(ctx iris.Context) {
	params := ctx.Params()
	id := params.Get("id")

	var jobPosts []models.JobPost
	jobPostsExists := storage.DB.Preload(clause.Associations).Where("user_id = ?", id).Find(&jobPosts)
	if jobPostsExists.Error != nil {
		return
	}
	ctx.JSON(jobPosts)
}

func DeleteJobPost(ctx iris.Context) {
	params := ctx.Params()
	id := params.Get("jobId")

	jobPostsDeleted := storage.DB.Delete(&models.JobPost{}, id)
	if jobPostsDeleted.Error != nil {
		utils.CreateError(iris.StatusInternalServerError, "Error", jobPostsDeleted.Error.Error(), ctx)
		return
	}
	ctx.StatusCode(iris.StatusNoContent)
}

func GetCommentsByJobPostID(ctx iris.Context) {
	params := ctx.Params()
	id := params.Get("id")

	var comments []models.Comment
	commentsExist := storage.DB.Preload(clause.Associations).Where("job_post_id = ?", id).Find(&comments)

	if commentsExist.Error != nil{
		utils.CreateError(iris.StatusInternalServerError, "Error", commentsExist.Error.Error(), ctx)
		return
	}
	ctx.JSON(comments)
}

type CreateJobPostInput struct {
	JobType       string `json:"jobType" validate:"required,oneof=petCare elderlyCare babySitter houseKeeping teacher"`
	Title         string `json:"title" validate:"required,max=256"`
	Description   string `json:"description" validate:"required,max=512"`
	Wage          int    `json:"wage" validate:"required"`
	WageFrequency string `json:"wageFrequency" validate:"required,oneof=monthly oneTime"`
	DateTime      string `json:"dateTime" validate:"required,max=20"`
	UserID        uint   `json:"userID" validate:"required"`
}

