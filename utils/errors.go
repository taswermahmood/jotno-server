package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
)

func CreateError(statusCode int, title string, detail string, ctx iris.Context) {
	ctx.StopWithProblem(
		statusCode,
		iris.NewProblem().Title(title).Detail(detail))
}

func InternalServerError(ctx iris.Context) {
	CreateError(
		iris.StatusInternalServerError,
		"Internal Server Error",
		"Could not handle the request.",
		ctx,
	)
}

func EmailAlreadyRegistered(ctx iris.Context) {
	CreateError(
		iris.StatusConflict,
		"Conflict",
		"The email entered is already registed, please try signing in.",
		ctx,
	)
}

func CreateNotFound(ctx iris.Context) {
	CreateError(
		iris.StatusNoContent,
		"No Data Found",
		"Could not find any data for the given request.",
		ctx,
	)
}

func CreateConflict(ctx iris.Context) {
	CreateError(
		iris.StatusConflict,
		"Conflict",
		"Data already exists for this request.",
		ctx,
	)
}

func CreateForbidden(ctx iris.Context) {
	CreateError(
		iris.StatusForbidden,
		"Forbidden",
		"Authentication Failed.",
		ctx,
	)
}

func ValidationError(err error, ctx iris.Context) {
	if errs, ok := err.(validator.ValidationErrors); ok {
		validationErrors := wrapValidationErrors(errs)

		fmt.Println("validationErrors", validationErrors)
		ctx.StopWithProblem(
			iris.StatusBadRequest,
			iris.NewProblem().
				Title("Validation error").
				Detail("One or more fields failed to be validated.").
				Key("errors", validationErrors))

		return
	}
	fmt.Print(err)
	InternalServerError(ctx)
}

func wrapValidationErrors(errs validator.ValidationErrors) []validationError {
	validationErrors := make([]validationError, 0, len(errs))
	for _, validationErr := range errs {
		validationErrors = append(validationErrors, validationError{
			ActualTag: validationErr.ActualTag(),
			Namespace: validationErr.Namespace(),
			Kind:      validationErr.Kind().String(),
			Type:      validationErr.Type().String(),
			Value:     fmt.Sprintf("%v", validationErr.Value()),
			Param:     validationErr.Param(),
		})
	}

	return validationErrors
}

type validationError struct {
	ActualTag string `json:"tag"`
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	Param     string `json:"param"`
}
