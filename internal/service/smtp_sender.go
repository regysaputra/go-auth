package service

import (
	"auth/internal"
	"bytes"
	"fmt"
	"html/template"
	"mime/multipart"
	"net/smtp"
)

// SmtpEmailSender Real implementation of EmailSender using an SMTP server
type SmtpEmailSender struct {
	Host      string
	Port      string
	Username  string
	Password  string
	From      string
	BaseURL   string
	templates *template.Template
}

// SmtpConfig holds all the necessary configuration for the SMTP sender.
type SmtpConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	BaseURL  string
}

// NewSmtpEmailSender creates a new SMTP email sender
func NewSmtpEmailSender(config SmtpConfig) *SmtpEmailSender {
	templates, err := template.ParseFS(internal.TemplateFS, "templates/*.html", "templates/*.txt")
	if err != nil {
		panic(fmt.Sprintf("failed to parse email templates: %v", err))
	}
	return &SmtpEmailSender{
		Host:      config.Host,
		Port:      config.Port,
		Username:  config.Username,
		Password:  config.Password,
		From:      config.From,
		BaseURL:   config.BaseURL,
		templates: templates,
	}
}

// SendEmailVerificationLink connects to the SMTP server and sends the email
func (sender *SmtpEmailSender) SendEmailVerificationLink(email, token string) error {
	data := map[string]string{
		"VerificationLink": fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", sender.BaseURL, token),
	}

	return sender.sendEmail(email, "Password Reset Link", "password_reset_link_template", data)
}

// SendEmailPasswordResetLink connects to the SMTP server and sends the email
func (sender *SmtpEmailSender) SendEmailPasswordResetLink(email string, token string) error {
	data := map[string]string{
		"ResetLink": fmt.Sprintf("%s/api/v1/auth/password/reset?token=%s", sender.BaseURL, token),
	}

	return sender.sendEmail(email, "Password Reset Link", "password_reset_link_template", data)
}

// SendEmailVerificationCode connects to the SMTP server and sends the email
func (sender *SmtpEmailSender) SendEmailVerificationCode(email string, code string) error {
	data := map[string]interface{}{
		"Code": code,
	}

	return sender.sendEmail(email, "Your Verification Code", "verification_email_template", data)
}

// SendEmailLoginOTP connects to the SMTP server and sends the email
func (sender *SmtpEmailSender) SendEmailLoginOTP(email string, code string) error {
	data := map[string]string{
		"LoginCode": code,
	}

	return sender.sendEmail(email, "Your login code", "login_otp_template", data)
}

// sendEmail is a helper function to construct and send email
func (sender *SmtpEmailSender) sendEmail(email string, subject string, templateName string, data any) error {
	auth := smtp.PlainAuth("", sender.Username, sender.Password, sender.Host)

	var body bytes.Buffer

	// Create a new multipart writer with a random boundary
	writer := multipart.NewWriter(&body)

	// Write the main headers. The content-type is now multipart/alternative
	fromHeader := fmt.Sprintf("From: %s\r\n", sender.From)
	toHeader := fmt.Sprintf("To: %s\r\n", email)
	subjectHeader := fmt.Sprintf("Subject: %s\r\n", subject)
	mimeHeader := "MIME-Version: 1.0\r\n"
	contentTypeHeader := fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n\r\n", writer.Boundary())

	body.WriteString(fromHeader)
	body.WriteString(toHeader)
	body.WriteString(subjectHeader)
	body.WriteString(mimeHeader)
	body.WriteString(contentTypeHeader)

	// Create a plain text part
	plainWriter, err := writer.CreatePart(map[string][]string{
		"Content-Type": {"text/plain: charset=\"UTF-8\""},
	})

	if err != nil {
		return err
	}

	err = sender.templates.ExecuteTemplate(plainWriter, templateName+".txt", data)

	if err != nil {
		return err
	}

	// Create an HTML part
	htmlWriter, err := writer.CreatePart(map[string][]string{
		"Content-Type": {"text/html: charset=\"UTF-8\""},
	})

	if err != nil {
		return err
	}

	err = sender.templates.ExecuteTemplate(htmlWriter, templateName+".html", data)
	if err != nil {
		return err
	}

	// Close the multipart writer to write the final boundary
	err = writer.Close()
	if err != nil {
		return err
	}

	address := fmt.Sprintf("%s:%s", sender.Host, sender.Port)

	return smtp.SendMail(address, auth, sender.From, []string{email}, body.Bytes())
}
