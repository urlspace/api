package handlers

import (
	"bytes"
	"html/template"
	texttemplate "text/template"
)

const emailVerifyTemplateHtml = `
  <!DOCTYPE html>
  <html>
		<body>
			<h1>Welcome to href.tools!</h1>
			<ul>
				<li>Your username: {{.Username}}</li>
				<li>Your email: {{.Email}}</li>
				<li>Please verify your email using the following token:</li>
				<li><a href="https://href.tools/verify?token={{.Token}}">https://href.tools/verify?token={{.Token}}</a></li>
				<li>This token will expire in 24 hours.</li>
			</ul> 
		</body>
  </html>
  `

const emailVerifyTemplateTxt = `
Welcome to href.tools!

Your username: {{.Username}}
Your email: {{.Email}}
Please verify your email using the following token: https://href.tools/verify?token={{.Token}}
This token will expire in 24 hours.`

type EmailVerifyData struct {
	Username string
	Email    string
	Token    string
}

func emailVerifyRenderHtml(data EmailVerifyData) (string, error) {
	tmpl, err := template.New("email").Parse(emailVerifyTemplateHtml)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func emailVerifyRenderTxt(data EmailVerifyData) (string, error) {
	tmpl, err := texttemplate.New("email").Parse(emailVerifyTemplateTxt)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
