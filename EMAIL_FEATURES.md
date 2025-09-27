# Email Verification & Password Reset Features

## Overview
EraLove backend now supports email verification and password reset functionality.

## Features Added

### 1. Email Verification
- **Registration**: Users receive verification email upon registration
- **Verification**: Users can verify their email using the token sent via email
- **Resend**: Users can request a new verification email if needed

### 2. Password Reset
- **Forgot Password**: Users can request password reset via email
- **Reset Password**: Users can reset password using the token sent via email

## API Endpoints

### Authentication Endpoints
- `POST /api/v1/auth/register` - Register new user (sends verification email)
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/verify-email` - Verify email address
- `POST /api/v1/auth/resend-verification` - Resend verification email
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password with token

## Request/Response Examples

### Register User
```json
POST /api/v1/auth/register
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123"
}
```

### Verify Email
```json
POST /api/v1/auth/verify-email
{
  "token": "verification-token-from-email"
}
```

### Resend Verification Email
```json
POST /api/v1/auth/resend-verification
{
  "email": "john@example.com"
}
```

### Forgot Password
```json
POST /api/v1/auth/forgot-password
{
  "email": "john@example.com"
}
```

### Reset Password
```json
POST /api/v1/auth/reset-password
{
  "token": "reset-token-from-email",
  "new_password": "newpassword123"
}
```

## Configuration

### Environment Variables
Add these to your `.env` file:

```env
# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=noreply@eralove.com
FROM_NAME=EraLove

# Frontend URL for email links
FRONTEND_URL=http://localhost:3000
```

### Gmail Setup
1. Enable 2-factor authentication on your Gmail account
2. Generate an App Password:
   - Go to Google Account settings
   - Security → 2-Step Verification → App passwords
   - Generate password for "Mail"
   - Use this password in `SMTP_PASSWORD`

## Database Changes

### User Model Updates
- `is_email_verified`: Boolean flag for email verification status
- `email_verification_token`: Token for email verification (hidden from API)
- `email_verification_expiry`: Expiry time for verification token
- `password_reset_token`: Token for password reset (hidden from API)
- `password_reset_expiry`: Expiry time for reset token

## Security Features

### Token Expiry
- **Email Verification**: 24 hours
- **Password Reset**: 1 hour

### Security Measures
- Tokens are cryptographically secure (32 bytes, hex encoded)
- Sensitive fields are hidden from API responses
- Email existence is not revealed for security
- Expired tokens are automatically rejected

## Email Templates

### Verification Email
- Beautiful HTML template with EraLove branding
- Clear call-to-action button
- Fallback link for manual copy-paste
- Security warnings and expiry information

### Password Reset Email
- Professional HTML template
- Security warnings about unauthorized requests
- Clear instructions for password reset
- Expiry time information

## Testing

### Using Swagger UI
1. Start the backend server
2. Go to `http://localhost:8080/swagger/`
3. Test the new endpoints:
   - Register a user
   - Check email for verification link
   - Use verification token in verify-email endpoint

### Development Mode
- CORS is configured to allow all origins in development
- Email sending can be disabled by leaving SMTP credentials empty
- All operations will be logged for debugging

## Error Handling

### Common Errors
- **Invalid/Expired Token**: Returns 400 with appropriate message
- **Email Already Verified**: Returns 400 with message
- **User Not Found**: Returns generic message for security
- **SMTP Not Configured**: Logs warning but doesn't fail registration

## Production Considerations

### Email Service
- Use a reliable SMTP service (SendGrid, AWS SES, etc.)
- Configure proper SPF/DKIM records
- Monitor email delivery rates

### Security
- Use strong JWT secrets
- Configure proper CORS origins
- Enable HTTPS in production
- Monitor for suspicious activities

### Performance
- Consider email queue for high volume
- Implement rate limiting for email endpoints
- Cache frequently accessed data

## Frontend Integration

### Email Verification Flow
1. User registers → receives verification email
2. User clicks link → redirected to frontend with token
3. Frontend calls `/api/v1/auth/verify-email` with token
4. Show success/error message to user

### Password Reset Flow
1. User requests reset → receives reset email
2. User clicks link → redirected to frontend with token
3. Frontend shows password reset form
4. Frontend calls `/api/v1/auth/reset-password` with token and new password
5. Show success/error message to user

## Monitoring

### Logs to Monitor
- Email sending failures
- Token verification attempts
- Expired token usage
- Suspicious password reset requests

### Metrics to Track
- Email delivery rates
- Verification completion rates
- Password reset success rates
- Failed authentication attempts
