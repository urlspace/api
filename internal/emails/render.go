package emails

import (
	"bytes"
	_ "embed"
	templateHtml "html/template"
	templateTxt "text/template"
)

//go:embed templates/auth-signup.html
var AuthSignupTemplateHtml string

//go:embed templates/auth-signup.txt
var AuthSignupTemplateTxt string

type AuthSignupParams struct {
	Username string
	Email    string
	Url      string
}

//go:embed templates/auth-resend-verification.html
var AuthResendVerificationTemplateHtml string

//go:embed templates/auth-resend-verification.txt
var AuthResendVerificationTemplateTxt string

type AuthResendVerificationParams struct {
	Url string
}

//go:embed templates/auth-reset-password-request.html
var AuthResetPasswordRequestTemplateHtml string

//go:embed templates/auth-reset-password-request.txt
var AuthResetPasswordRequestTemplateTxt string

type AuthResetPasswordRequestParams struct {
	Url string
}

//go:embed templates/admin-new-user.html
var AdminNewUserTemplateHtml string

//go:embed templates/admin-new-user.txt
var AdminNewUserTemplateTxt string

type AdminNewUserParams struct {
	Username string
	Email    string
}

func RenderTemplateHtml(template string, data any) (string, error) {
	tmpl, err := templateHtml.New("email").Parse(template)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func RenderTemplateTxt(template string, data any) (string, error) {
	tmpl, err := templateTxt.New("email").Parse(template)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
