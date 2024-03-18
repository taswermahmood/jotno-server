package routes

import (
	"jotno-server/utils"
)

func SendNotification(
	url string,
	token string,
	title string,
	body string,
) {
	data := map[string]string{"url": url}

	err := utils.SendNotification(token, title, body, data)
	if err != nil {
		return
	}
}
