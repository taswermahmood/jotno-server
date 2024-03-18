package utils

import (
	"strconv"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

func UserIDMiddleware(ctx iris.Context) {
	id := ctx.URLParam("id")
	claims := jwt.Get(ctx).(*AccessToken)

	userID := strconv.FormatUint(uint64(claims.ID), 10)

	if userID != id {
		CreateForbidden(ctx)
		return
	}
	ctx.Next()
}
