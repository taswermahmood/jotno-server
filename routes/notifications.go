package routes

import (
	"jotno-server/utils"

	"github.com/kataras/iris/v12"
)

func TestMessageNotification(ctx iris.Context) {
	data := map[string]string{
		"url": "exp://10.0.0.240:8081/--/screens/messages/2/TestNotification",
	}

	err := utils.SendNotification(
		"ExponentPushToken[Xxxxxxxxxxxxxxxxxxxxxx]",
		"Push Title",
		"Push body is this message",
		data,
	)
	if err != nil {
		utils.InternalServerError(ctx)
		return
	}

	ctx.JSON(iris.Map{
		"sent": true,
	})
}
