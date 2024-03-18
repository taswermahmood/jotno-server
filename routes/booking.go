package routes

import (
	"encoding/json"
	"fmt"
	"jotno-server/models"
	"jotno-server/storage"
	"jotno-server/utils"
	"strconv"

	"github.com/kataras/iris/v12"
)

func CreateBooking(ctx iris.Context) {
	var bookingInput CreateBookingInput

	err := ctx.ReadJSON(&bookingInput)
	if err != nil {
		utils.ValidationError(err, ctx)
		return
	}
	booking := models.Booking{
		UserID:       bookingInput.UserID,
		SpecialistID: bookingInput.SpecialistID,
		JobType:      bookingInput.JobType,
		Active:       false,
		Status:       "pending",
		Frequency:    bookingInput.Frequency,
		Amount:       bookingInput.Amount,
		Currency:     bookingInput.Currency,
		Overdue:      false,
		StartDate:    bookingInput.StartDate,
		EndDate:      bookingInput.EndDate,
	}
	storage.DB.Create(&booking)
	ctx.JSON(booking)
}

func CreateBill(ID uint, amount int32, currency string, userID uint) {
	var bill models.Bill
	billExists := storage.DB.Where("booking_id = ?", strconv.FormatUint(uint64(ID), 10)).Order("created_at DESC").First(&bill)
	if billExists.Error != nil {
		payment := models.Bill{
			BookingID: ID,
			Paid:      false,
			Received:  false,
			Complete:  false,
			Amount:    amount,
			Currency:  currency,
		}
		fmt.Println("first")
		storage.DB.Create(&payment)
		return
	}
	if bill.Complete {
		payment := models.Bill{
			BookingID: ID,
			Paid:      false,
			Received:  false,
			Complete:  false,
			Amount:    amount,
			Currency:  currency,
		}
		fmt.Println("second")
		storage.DB.Create(&payment)
		return
	}
	var user models.User
	userExists := storage.DB.Where("id = ?", userID).First(&user)
	if userExists.Error == nil && userExists.RowsAffected == 1 {
		var tokens []string
		if user.PushTokens != nil {
			unmarshalErr := json.Unmarshal(user.PushTokens, &tokens)
			if unmarshalErr == nil {
				for i := 0; i < len(tokens); i++ {
					SendNotification("exp://10.0.0.240:8081/--/tabs/bookingScreen/", tokens[i], "Your specialist is waiting.", "Finish paying you specialist for uninterupted service.")
				}
			}
		}
	}
}

func GetBookingByUserID(ctx iris.Context) {
	id := ctx.URLParam("id")
	response := map[string][]models.Booking{}

	var active []models.Booking
	activeExists := storage.DB.Where("user_id = ? AND active = true", id).Order("created_at DESC").Find(&active)
	if activeExists.Error != nil {
		response["active"] = nil
	} else {
		for i := 0; i < len(active); i++ {
			var bill models.Bill
			billExists := storage.DB.Where("booking_id = ? AND complete = false", active[i].ID).Order("created_at DESC").First(&bill)
			if billExists.Error == nil {
				active[i].Bills = append(active[i].Bills, bill)
			}
		}
		response["active"] = active
	}

	var pending []models.Booking
	pendingExists := storage.DB.Where("user_id = ? AND status = pending", id).Order("created_at DESC").Find(&pending)
	if pendingExists.Error != nil {
		response["pending"] = nil
	} else {
		response["pending"] = pending
	}

	var completed []models.Booking
	completedExists := storage.DB.Where("user_id = ? AND status = completed", id).Order("created_at DESC").Find(&completed)
	if completedExists.Error != nil {
		response["completed"] = nil
	} else {
		response["completed"] = completed
	}

	ctx.JSON(response)
}

func CancelBooking(ctx iris.Context) {
	id := ctx.URLParam("bookingID")

	bookingDeleted := storage.DB.Delete(&models.Booking{}, id)
	if bookingDeleted.Error != nil {
		utils.CreateError(iris.StatusInternalServerError, "Error", bookingDeleted.Error.Error(), ctx)
		return
	}
	ctx.StatusCode(iris.StatusNoContent)
}

func GetPendingPaymentsByBookingID(ctx iris.Context) {
	id := ctx.URLParam("bookingId")

	var payments models.Comment
	paymentsExists := storage.DB.Where("booking_id = ? AND complete = false", id).Order("created_at DESC").First(&payments)

	if paymentsExists.Error != nil {
		utils.CreateError(iris.StatusInternalServerError, "Error", paymentsExists.Error.Error(), ctx)
		return
	}
	ctx.JSON(payments)
}

type CreateBookingInput struct {
	UserID       uint   `json:"userID" validate:"required"`
	SpecialistID uint   `json:"specialistID" validate:"required"`
	JobType      string `json:"jobType" validate:"required,oneof=petCare elderlyCare babySitting houseKeeping teaching"`
	Frequency    string `json:"frequency" validate:"required,oneof=monthly daily"`
	Amount       int32  `json:"amount" validate:"required"`
	Currency     string `json:"currency" validate:"required"`
	StartDate    string `json:"startDate" validate:"required"`
	EndDate      string `json:"endDate"`
}
