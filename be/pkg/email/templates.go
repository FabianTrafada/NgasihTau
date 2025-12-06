package email

const verificationTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .container { background: #f9f9f9; border-radius: 8px; padding: 30px; }
        .header { text-align: center; margin-bottom: 30px; }
        .header h1 { color: #2563eb; margin: 0; }
        .content { background: white; border-radius: 8px; padding: 30px; }
        .button { display: inline-block; background: #2563eb; color: white; padding: 12px 30px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .button:hover { background: #1d4ed8; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
        .expiry { color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.AppName}}</h1>
        </div>
        <div class="content">
            <h2>Verify Your Email Address</h2>
            <p>Hi {{.RecipientName}},</p>
            <p>Thank you for signing up! Please verify your email address by clicking the button below:</p>
            <p style="text-align: center;">
                <a href="{{.ActionURL}}" class="button">Verify Email</a>
            </p>
            <p class="expiry">This link will expire in {{.ExpiryHours}} hours.</p>
            <p>If you didn't create an account with {{.AppName}}, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>© {{.AppName}}. All rights reserved.</p>
            <p>Need help? Contact us at {{.SupportEmail}}</p>
        </div>
    </div>
</body>
</html>`

const passwordResetTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Your Password</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .container { background: #f9f9f9; border-radius: 8px; padding: 30px; }
        .header { text-align: center; margin-bottom: 30px; }
        .header h1 { color: #2563eb; margin: 0; }
        .content { background: white; border-radius: 8px; padding: 30px; }
        .button { display: inline-block; background: #2563eb; color: white; padding: 12px 30px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .button:hover { background: #1d4ed8; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
        .expiry { color: #666; font-size: 14px; }
        .warning { background: #fef3c7; border-left: 4px solid #f59e0b; padding: 12px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.AppName}}</h1>
        </div>
        <div class="content">
            <h2>Reset Your Password</h2>
            <p>Hi {{.RecipientName}},</p>
            <p>We received a request to reset your password. Click the button below to create a new password:</p>
            <p style="text-align: center;">
                <a href="{{.ActionURL}}" class="button">Reset Password</a>
            </p>
            <p class="expiry">This link will expire in {{.ExpiryHours}} hour(s).</p>
            <div class="warning">
                <strong>Didn't request this?</strong><br>
                If you didn't request a password reset, please ignore this email. Your password will remain unchanged.
            </div>
        </div>
        <div class="footer">
            <p>© {{.AppName}}. All rights reserved.</p>
            <p>Need help? Contact us at {{.SupportEmail}}</p>
        </div>
    </div>
</body>
</html>`

const collaboratorInviteTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>You're Invited to Collaborate</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .container { background: #f9f9f9; border-radius: 8px; padding: 30px; }
        .header { text-align: center; margin-bottom: 30px; }
        .header h1 { color: #2563eb; margin: 0; }
        .content { background: white; border-radius: 8px; padding: 30px; }
        .button { display: inline-block; background: #2563eb; color: white; padding: 12px 30px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .button:hover { background: #1d4ed8; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
        .pod-info { background: #eff6ff; border-radius: 8px; padding: 20px; margin: 20px 0; }
        .pod-name { font-size: 18px; font-weight: bold; color: #1e40af; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.AppName}}</h1>
        </div>
        <div class="content">
            <h2>You're Invited to Collaborate!</h2>
            <p>Hi {{.RecipientName}},</p>
            <p><strong>{{.InviterName}}</strong> has invited you to collaborate on a Knowledge Pod:</p>
            <div class="pod-info">
                <span class="pod-name">{{.PodName}}</span>
            </div>
            <p>Click the button below to accept the invitation and start collaborating:</p>
            <p style="text-align: center;">
                <a href="{{.ActionURL}}" class="button">Accept Invitation</a>
            </p>
            <p>If you don't want to join, simply ignore this email.</p>
        </div>
        <div class="footer">
            <p>© {{.AppName}}. All rights reserved.</p>
            <p>Need help? Contact us at {{.SupportEmail}}</p>
        </div>
    </div>
</body>
</html>`
