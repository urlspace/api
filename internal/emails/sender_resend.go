package emails

import (
	"github.com/resend/resend-go/v3"
)

type ResendEmailSender struct {
	client *resend.Client
}

func (s *ResendEmailSender) Send(params EmailSendParams) error {
	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    "url.space <auth@mail.url.space>",
		To:      params.To,
		Html:    params.Html,
		Text:    params.Text,
		Subject: params.Subject,
		ReplyTo: "mail@url.space",
	})

	return err
}

func NewResendEmailSender(client *resend.Client) *ResendEmailSender {
	return &ResendEmailSender{
		client: client,
	}
}
