package emails

type EmailSendParams struct {
	To      []string
	Html    string
	Text    string
	Subject string
}

type EmailSender interface {
	Send(params EmailSendParams) error
}
