package service

import (
	"auth/internal"
	"bytes"
	"context"
	"fmt"
	"html/template"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
)

// SMTPEmailSender Real implementation of EmailSender using an SMTP server
type SMTPEmailSender struct {
	Host      string
	Port      string
	Username  string
	Password  string
	From      string
	BaseURL   string
	templates *template.Template
}

// SMTPConfig holds all the necessary configuration for the SMTP sender.
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	BaseURL  string
}

// NewSMTPEmailSender creates a new SMTP email sender
func NewSMTPEmailSender(config SMTPConfig) *SMTPEmailSender {
	templates, err := template.ParseFS(internal.TemplateFS, "templates/*.html", "templates/*.txt")
	if err != nil {
		panic(fmt.Sprintf("failed to parse email templates: %v", err))
	}
	return &SMTPEmailSender{
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
func (sender *SMTPEmailSender) SendEmailVerificationLink(ctx context.Context, email, token string) error {
	data := map[string]string{
		"VerificationLink": fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", sender.BaseURL, token),
	}

	return sender.sendEmail(ctx, email, "password_reset_link_template", data)
}

// SendEmailPasswordResetLink connects to the SMTP server and sends the email
func (sender *SMTPEmailSender) SendEmailPasswordResetLink(ctx context.Context, email string, token string) error {
	data := map[string]string{
		"ResetLink": fmt.Sprintf("%s/api/v1/auth/password/reset?token=%s", sender.BaseURL, token),
	}

	return sender.sendEmail(ctx, email, "password_reset_link_template", data)
}

// SendEmailVerificationCode connects to the SMTP server and sends the email
func (sender *SMTPEmailSender) SendEmailVerificationCode(ctx context.Context, email string, code string) error {
	data := map[string]interface{}{
		"Code": code,
	}

	return sender.sendEmail(ctx, email, "verification_email_template", data)
}

// SendEmailLoginOTP connects to the SMTP server and sends the email
func (sender *SMTPEmailSender) SendEmailLoginOTP(ctx context.Context, email string, code string) error {
	data := map[string]string{
		"LoginCode": code,
	}

	return sender.sendEmail(ctx, email, "login_otp_template", data)
}

// sendEmail is a helper function to construct and send email
func (sender *SMTPEmailSender) sendEmail(ctx context.Context, email string, templateName string, data any) error {
	//body.WriteString(fromHeader)
	//body.WriteString(toHeader)
	//body.WriteString(subjectHeader)
	//body.WriteString(mimeHeader)
	//body.WriteString(contentTypeHeader)

	var body bytes.Buffer

	// --- 1. Set standard email headers ---
	fromHeader := fmt.Sprintf("From: %s\r\n", sender.From)
	toHeader := fmt.Sprintf("To: %s\r\n", email)

	// --- 2. Execute the Subject template ---
	subjectTplName := fmt.Sprintf("%s.subject.txt", templateName)
	var subjectBuf bytes.Buffer
	err := sender.templates.ExecuteTemplate(&subjectBuf, subjectTplName, data)
	if err != nil {
		return fmt.Errorf("failed to execute subject template %s: %w", subjectTplName, err)
	}
	subjectHeader := fmt.Sprintf("Subject: %s\r\n", subjectBuf.String())

	body.Write([]byte(fromHeader + toHeader + subjectHeader))

	// --- 3. Set up the multipart writer ---
	// This creates the boundary string that separates the parts of the email.
	mimeWriter := multipart.NewWriter(&body)
	mimeHeader := fmt.Sprintf("MIME-version: 1.0;\r\nContent-Type: multipart/alternative; boundary=%s\n\n", mimeWriter.Boundary())
	body.Write([]byte(mimeHeader))

	// --- 4. Write the Plain Text part ---
	textWriter, err := mimeWriter.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/plain; charset=UTF-8"}})
	if err != nil {
		return err
	}
	txtTplName := fmt.Sprintf("%s.txt", templateName)
	err = sender.templates.ExecuteTemplate(textWriter, txtTplName, data)
	if err != nil {
		return fmt.Errorf("failed to execute text template %s: %w", txtTplName, err)
	}

	// --- 5. Write the HTML part ---
	htmlWriter, err := mimeWriter.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/html; charset=UTF-8"}})
	if err != nil {
		return err
	}
	htmlTplName := fmt.Sprintf("%s.html", templateName)
	err = sender.templates.ExecuteTemplate(htmlWriter, htmlTplName, data)
	if err != nil {
		return fmt.Errorf("failed to execute html template %s: %w", htmlTplName, err)
	}

	// --- 6. Close the multipart writer to add the final boundary ---
	err = mimeWriter.Close()
	if err != nil {
		return err
	}

	// --- 7. Send the email ---
	auth := smtp.PlainAuth("", sender.Username, sender.Password, sender.Host)
	addr := fmt.Sprintf("%s:%s", sender.Host, sender.Port)

	// --- THIS IS THE FIX ---
	// We run the blocking call in a goroutine and listen for its
	// completion on a channel, OR for the context to be canceled.

	errChan := make(chan error, 1)
	go func() {
		err := smtp.SendMail(addr, auth, sender.From, []string{email}, body.Bytes())
		errChan <- err
	}()

	select {
	case <-ctx.Done():
		// The context was canceled (e.g., worker shutting down).
		return ctx.Err() // This will be context.Canceled or context.DeadlineExceeded

	case err := <-errChan:
		// The smtp.SendMail function finished.
		if err != nil {
			return fmt.Errorf("failed to send smtp mail: %w", err)
		}
		// The email was sent successfully.
		return nil
	}
}
