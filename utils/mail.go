package utils

import (
	"github.com/mailjet/mailjet-apiv3-go/v4"
	"os"
)

func SendMail(userEmail string, subject string, html string) (bool, error) {
	mailjetClient := mailjet.NewMailjetClient(os.Getenv("EMAIL_API_KEY"), os.Getenv("EMAIL_SECRET_KEY"))
	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: "jotnoexpo@gmail.com",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: userEmail,
				},
			},
			Subject:  subject,
			HTMLPart: html,
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := mailjetClient.SendMailV31(&messages)
	if err != nil {
		return false, err
	}

	return true, nil
}
