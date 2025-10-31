package mailer

import "embed"

const (
	FromName            = "Baldur"
	maxRetries          = 3
	UserWelcomeTemplate = "user_invitation.templ"
)

//go:embed templates/*
var FS embed.FS

type Client interface {
	Send(templateFile, username, email string, data any, isProduction bool) (int, error)
}
