package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/eralove/eralove-backend/internal/config"
	"go.uber.org/zap"
)

// EmailService handles email sending
type EmailService struct {
	config *config.Config
	logger *zap.Logger
}

// NewEmailService creates a new email service
func NewEmailService(config *config.Config, logger *zap.Logger) *EmailService {
	return &EmailService{
		config: config,
		logger: logger,
	}
}

// EmailData represents data for email templates
type EmailData struct {
	Name         string
	Email        string
	Token        string
	VerifyURL    string
	ResetURL     string
	FrontendURL  string
	SupportEmail string
}

// SendVerificationEmail sends email verification email
func (s *EmailService) SendVerificationEmail(name, email, token string) error {
	subject := "Verify Your Email - EraLove"
	
	data := EmailData{
		Name:         name,
		Email:        email,
		Token:        token,
		VerifyURL:    fmt.Sprintf("%s/verify-email?token=%s", s.config.FrontendURL, token),
		FrontendURL:  s.config.FrontendURL,
		SupportEmail: s.config.FromEmail,
	}

	body, err := s.renderTemplate(verificationEmailTemplate, data)
	if err != nil {
		s.logger.Error("Failed to render verification email template", zap.Error(err))
		return fmt.Errorf("failed to render email template: %w", err)
	}

	return s.sendEmail(email, subject, body)
}

// SendPasswordResetEmail sends password reset email
func (s *EmailService) SendPasswordResetEmail(name, email, token string) error {
	subject := "Reset Your Password - EraLove"
	
	data := EmailData{
		Name:         name,
		Email:        email,
		Token:        token,
		ResetURL:     fmt.Sprintf("%s/reset-password?token=%s", s.config.FrontendURL, token),
		FrontendURL:  s.config.FrontendURL,
		SupportEmail: s.config.FromEmail,
	}

	body, err := s.renderTemplate(passwordResetEmailTemplate, data)
	if err != nil {
		s.logger.Error("Failed to render password reset email template", zap.Error(err))
		return fmt.Errorf("failed to render email template: %w", err)
	}

	return s.sendEmail(email, subject, body)
}

// sendEmail sends an email using SMTP
func (s *EmailService) sendEmail(to, subject, body string) error {
	// Skip sending email if SMTP is not configured
	if s.config.SMTPUsername == "" || s.config.SMTPPassword == "" {
		s.logger.Warn("SMTP not configured, skipping email send", 
			zap.String("to", to), 
			zap.String("subject", subject))
		return nil
	}

	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", to, subject, body))

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	
	err := smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, msg)
	if err != nil {
		s.logger.Error("Failed to send email", 
			zap.Error(err),
			zap.String("to", to),
			zap.String("subject", subject))
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("Email sent successfully", 
		zap.String("to", to),
		zap.String("subject", subject))
	
	return nil
}

// renderTemplate renders an email template with data
func (s *EmailService) renderTemplate(templateStr string, data EmailData) (string, error) {
	tmpl, err := template.New("email").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// Email templates
const verificationEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify Your Email</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #ff6b9d; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 24px; background-color: #ff6b9d; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { padding: 20px; text-align: center; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to EraLove! 💕</h1>
        </div>
        <div class="content">
            <h2>Hi {{.Name}},</h2>
            <p>Thank you for signing up for EraLove! We're excited to help you track your love journey.</p>
            <p>To complete your registration, please verify your email address by clicking the button below:</p>
            <p style="text-align: center;">
                <a href="{{.VerifyURL}}" class="button">Verify Email Address</a>
            </p>
            <p>If the button doesn't work, you can copy and paste this link into your browser:</p>
            <p><a href="{{.VerifyURL}}">{{.VerifyURL}}</a></p>
            <p>This verification link will expire in 24 hours for security reasons.</p>
            <p>If you didn't create an account with EraLove, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>Need help? Contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
            <p>&copy; 2024 EraLove. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const passwordResetEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #ff6b9d; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 24px; background-color: #ff6b9d; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { padding: 20px; text-align: center; color: #666; font-size: 12px; }
        .warning { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 10px; border-radius: 5px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request 🔐</h1>
        </div>
        <div class="content">
            <h2>Hi {{.Name}},</h2>
            <p>We received a request to reset your password for your EraLove account.</p>
            <p>If you requested this password reset, click the button below to set a new password:</p>
            <p style="text-align: center;">
                <a href="{{.ResetURL}}" class="button">Reset Password</a>
            </p>
            <p>If the button doesn't work, you can copy and paste this link into your browser:</p>
            <p><a href="{{.ResetURL}}">{{.ResetURL}}</a></p>
            <div class="warning">
                <strong>Important:</strong>
                <ul>
                    <li>This password reset link will expire in 1 hour for security reasons.</li>
                    <li>If you didn't request a password reset, please ignore this email.</li>
                    <li>Your password will remain unchanged until you create a new one.</li>
                </ul>
            </div>
        </div>
        <div class="footer">
            <p>Need help? Contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
            <p>&copy; 2024 EraLove. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`
