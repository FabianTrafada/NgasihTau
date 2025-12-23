package email

const verificationTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1e293b; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f8fafc; background-image: radial-gradient(#cbd5e1 1px, transparent 1px); background-size: 20px 20px; }
        .container { background: #ffffff; border: 2px solid #1e293b; border-radius: 16px; box-shadow: 8px 8px 0px #FF8811; padding: 40px; margin-top: 20px; }
        .header { text-align: center; margin-bottom: 30px; }
        .header h1 { color: #1e293b; margin: 0; font-size: 28px; font-weight: 800; letter-spacing: -0.5px; }
        .content { padding: 10px 0; text-align: center; }
        h2 { font-size: 24px; font-weight: 700; margin-bottom: 20px; color: #1e293b; }
        p { margin-bottom: 15px; font-size: 16px; color: #475569; }
        .button { display: inline-block; background: #FF8811; color: #ffffff; padding: 14px 32px; text-decoration: none; border: 2px solid #1e293b; border-radius: 8px; margin: 20px 0; font-weight: 700; box-shadow: 4px 4px 0px #1e293b; transition: all 0.2s; }
        .button:hover { transform: translate(-2px, -2px); box-shadow: 6px 6px 0px #1e293b; }
        .footer { text-align: center; margin-top: 40px; font-size: 12px; color: #94a3b8; padding-top: 20px; }
        .expiry { font-size: 14px; color: #64748b; background: #f1f5f9; border-radius: 6px; padding: 10px; display: inline-block; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.AppName}}</h1>
        </div>
        <div class="content">
            <h2>Verify Your Email</h2>
            <p>Hi {{.RecipientName}},</p>
            <p>Thank you for signing up! We need to verify your email address to get you started.</p>
            <p style="text-align: center;">
                <a href="{{.ActionURL}}" class="button">Verify Email</a>
            </p>
            <div style="text-align: center;">
                <p class="expiry">This link expires in {{.ExpiryHours}} hours.</p>
            </div>
            <p>If you didn't create an account with {{.AppName}}, just ignore this email.</p>
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
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1e293b; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f8fafc; background-image: radial-gradient(#cbd5e1 1px, transparent 1px); background-size: 20px 20px; }
        .container { background: #ffffff; border: 2px solid #1e293b; border-radius: 16px; box-shadow: 8px 8px 0px #FF8811; padding: 40px; margin-top: 20px; }
        .header { text-align: center; margin-bottom: 30px; }
        .header h1 { color: #1e293b; margin: 0; font-size: 28px; font-weight: 800; letter-spacing: -0.5px; }
        .content { padding: 10px 0; text-align: center; }
        h2 { font-size: 24px; font-weight: 700; margin-bottom: 20px; color: #1e293b; }
        p { margin-bottom: 15px; font-size: 16px; color: #475569; }
        .button { display: inline-block; background: #FF8811; color: #ffffff; padding: 14px 32px; text-decoration: none; border: 2px solid #1e293b; border-radius: 8px; margin: 20px 0; font-weight: 700; box-shadow: 4px 4px 0px #1e293b; transition: all 0.2s; }
        .button:hover { transform: translate(-2px, -2px); box-shadow: 6px 6px 0px #1e293b; }
        .footer { text-align: center; margin-top: 40px; font-size: 12px; color: #94a3b8; padding-top: 20px; }
        .expiry { font-size: 14px; color: #64748b; background: #f1f5f9; border-radius: 6px; padding: 10px; display: inline-block; margin-top: 20px; }
        .warning { background: #fff7ed; border: 2px solid #FF8811; border-radius: 8px; padding: 15px; margin: 20px 0; color: #9a3412; font-weight: 600; }
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
            <p>We received a request to reset your password. Click the button below to create a new one:</p>
            <p style="text-align: center;">
                <a href="{{.ActionURL}}" class="button">Reset Password</a>
            </p>
            <div style="text-align: center;">
                <p class="expiry">This link expires in {{.ExpiryHours}} hour(s).</p>
            </div>
            <div class="warning">
                <strong>Didn't request this?</strong><br>
                Ignore this email. Your password will stay safe.
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
    <title>You're Invited</title>
    <style>1e293b; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f8fafc; background-image: radial-gradient(#cbd5e1 1px, transparent 1px); background-size: 20px 20px; }
        .container { background: #ffffff; border: 2px solid #1e293b; border-radius: 16px; box-shadow: 8px 8px 0px #FF8811; padding: 40px; margin-top: 20px; }
        .header { text-align: center; margin-bottom: 30px; }
        .header h1 { color: #1e293b; margin: 0; font-size: 28px; font-weight: 800; letter-spacing: -0.5px; }
        .content { padding: 10px 0; text-align: center; }
        h2 { font-size: 24px; font-weight: 700; margin-bottom: 20px; color: #1e293b; }
        p { margin-bottom: 15px; font-size: 16px; color: #475569; }
        .button { display: inline-block; background: #FF8811; color: #ffffff; padding: 14px 32px; text-decoration: none; border: 2px solid #1e293b; border-radius: 8px; margin: 20px 0; font-weight: 700; box-shadow: 4px 4px 0px #1e293b; transition: all 0.2s; }
        .button:hover { transform: translate(-2px, -2px); box-shadow: 6px 6px 0px #1e293b; }
        .footer { text-align: center; margin-top: 40px; font-size: 12px; color: #94a3b8; padding-top: 20px; }
        .pod-info { background: #fff7ed; border: 2px solid #1e293b; border-radius: 8px; padding: 20px; margin: 20px 0; box-shadow: 4px 4px 0px #1e293b; text-align: left; }
        .pod-name { font-size: 20px; font-weight: 800; color: #1e293b; text-transform: uppercase; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.AppName}}</h1>
        </div>
        <div class="content">
            <h2>Collaboration Invite</h2>
            <p>Hi {{.RecipientName}},</p>
            <p><strong>{{.InviterName}}</strong> wants you to collaborate on a Knowledge Pod:</p>
            <div class="pod-info">
                <div style="font-size: 12px; font-weight: bold; margin-bottom: 5px; color: #FF8811
                <div style="font-size: 12px; font-weight: bold; margin-bottom: 5px;">KNOWLEDGE POD</div>
                <span class="pod-name">{{.PodName}}</span>
            </div>
            <p>Join the team by clicking below:</p>
            <p style="text-align: center;">
                <a href="{{.ActionURL}}" class="button">Accept Invite</a>
            </p>
            <p>Not interested? You can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>© {{.AppName}}. All rights reserved.</p>
            <p>Need help? Contact us at {{.SupportEmail}}</p>
        </div>
    </div>
</body>
</html>`
