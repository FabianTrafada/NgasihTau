package email

const verificationTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Inter', sans-serif; 
            line-height: 1.6; 
            color: #1e293b; 
            max-width: 600px; 
            margin: 0 auto; 
            padding: 20px; 
            background-color: #FDFCF9;
        }
        .container { 
            background-color: #fffbf5;
            background-image: radial-gradient(circle, rgba(232, 220, 200, 0.3) 1.5px, transparent 1.5px);
            background-size: 20px 20px;
            border: 3px solid #000; 
            border-radius: 24px; 
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15) !important; 
            -webkit-box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15) !important;
            -moz-box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15) !important;
            padding: 60px 40px; 
            margin-top: 20px; 
            position: relative;
            overflow: hidden;
        }
        
        .content-wrapper {
            position: relative;
            z-index: 1;
        }
        
        .header { 
            text-align: center; 
            margin-bottom: 40px; 
        }
        .header h1 { 
            margin: 0 0 5px 0; 
            font-size: 48px; 
            font-weight: 700; 
        }
        .header h1 .highlight {
            color: #ff8c00;
        }
        .header h1 .dark {
            color: #000;
        }
        .brush-underline {
            width: 300px;
            height: 8px;
            margin: 5px auto 0;
            background: #ff8c00;
            position: relative;
            border-radius: 50%;
            transform: scaleY(0.6);
        }
        .brush-underline::before {
            content: '';
            position: absolute;
            left: -3px;
            top: -1px;
            width: 100%;
            height: 100%;
            background: inherit;
            border-radius: 50%;
            opacity: 0.4;
        }
        .brush-underline::after {
            content: '';
            position: absolute;
            right: -3px;
            top: 1px;
            width: 100%;
            height: 100%;
            background: inherit;
            border-radius: 50%;
            opacity: 0.3;
        }
        
        .badge {
            display: inline-block;
            background: #2c3e50;
            color: white;
            padding: 8px 20px;
            border-radius: 6px;
            font-size: 11px;
            font-weight: 600;
            letter-spacing: 1px;
            text-transform: uppercase;
            margin-bottom: 30px;
        }
        
        .content { 
            text-align: center; 
        }
        h2 { 
            font-size: 42px; 
            font-weight: 700; 
            margin-bottom: 10px; 
            color: #000;
        }
        .title-underline {
            width: 60px;
            height: 4px;
            background: #ff8c00;
            margin: 0 auto 30px;
            border-radius: 2px;
        }
        
        p { 
            margin-bottom: 15px; 
            font-size: 15px; 
            color: #999;
            line-height: 1.6;
        }
        p.greeting {
            font-size: 24px;
            color: #333;
            font-weight: 600;
            margin-bottom: 15px;
        }
        
        .button { 
            display: inline-block;
            background: #ff8c00;
            color: #ffffff; 
            padding: 18px 50px; 
            text-decoration: none; 
            border: 3px solid #000; 
            border-radius: 12px; 
            margin: 40px 0 20px; 
            font-weight: 700; 
            font-size: 18px;
            text-transform: uppercase;
            position: relative;
        }
        .button-wrapper {
            display: inline-block;
            position: relative;
        }
        .button-shadow {
            position: absolute;
            top: 6px;
            left: 6px;
            right: -6px;
            bottom: -6px;
            background: #ff8c00;
            border-radius: 12px;
            z-index: -1;
        }
        
        .warning-box {
            background: white;
            border: 2px solid #ffa500;
            border-radius: 12px;
            padding: 16px 20px;
            margin: 20px auto 30px;
            display: block;
            text-align: center;
            max-width: 450px;
        }
        .warning-content {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 12px;
        }
        .warning-icon {
            display: inline-block;
            width: 24px;
            height: 24px;
            background: #ffa500;
            border-radius: 50%;
            color: white;
            font-weight: bold;
            text-align: center;
            line-height: 24px;
            margin-right: 12px;
            vertical-align: middle;
        }
        .warning-text {
            display: inline-block;
            color: #666;
            font-size: 14px;
            text-align: left;
            line-height: 1.5;
        }
        
        .ignore-text {
            color: #666;
            font-size: 14px;
            margin-bottom: 40px;
        }
        
        .footer { 
            text-align: center; 
            margin-top: 40px; 
            font-size: 13px; 
            color: #666; 
            padding-top: 0;
            line-height: 1.8;
        }
        .footer p {
            margin: 8px 0;
        }
        .footer-bold {
            font-weight: 600;
            color: #333;
        }
        .support-link {
            color: #ff8c00;
            text-decoration: underline;
        }
        
        @media (max-width: 600px) {
            .container {
                padding: 40px 25px;
            }
            .header h1 {
                font-size: 36px;
            }
            h2 {
                font-size: 32px;
            }
            .button {
                padding: 15px 40px;
                font-size: 16px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="content-wrapper">
            <div class="header">
                <h1><span class="highlight">Ngasih</span><span class="dark">Tau</span></h1>
                <div class="brush-underline"></div>
            </div>
            
            <div class="content">
                <div class="badge">Email Verification</div>
                <h2>Verify Your Email</h2>
                <div class="title-underline"></div>
                
                <p class="greeting">Hello {{.RecipientName}},</p>
                <p>Thank you for joining our learning community! To get started, we need to verify your email address.</p>
                
                <p style="text-align: center;">
                    <span class="button-wrapper">
                        <span class="button-shadow"></span>
                        <a style="color: #ffffff;" href="{{.ActionURL}}" class="button">Verify Email</a>
                    </span>
                </p>
                
                <div class="warning-box">
                    <div class="warning-content">
                        <span class="warning-icon">!</span>
                        <span class="warning-text">This link expires in {{.ExpiryHours}} hours. Please complete verification soon.</span>
                    </div>
                </div>
                
                <p class="ignore-text">If you didn't create an account with {{.AppName}}, just ignore this email.</p>
            </div>
            
            <div class="footer">
                <p class="footer-bold">© {{.AppName}}. All rights reserved.</p>
                <p>Need help? Contact us at <a href="mailto:{{.SupportEmail}}" class="support-link">{{.SupportEmail}}</a></p>
            </div>
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
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Inter', sans-serif; 
            line-height: 1.6; 
            color: #1e293b; 
            max-width: 600px; 
            margin: 0 auto; 
            padding: 20px; 
            background-color: #FDFCF9;
        }
        .container { 
            background-color: #fffbf5;
            background-image: radial-gradient(circle, rgba(232, 220, 200, 0.3) 1.5px, transparent 1.5px);
            background-size: 20px 20px;
            border: 3px solid #000; 
            border-radius: 24px; 
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15); 
            padding: 60px 40px; 
            margin-top: 20px; 
            position: relative;
            overflow: hidden;
        }
        
        .content-wrapper {
            position: relative;
            z-index: 1;
        }
        
        .header { 
            text-align: center; 
            margin-bottom: 40px; 
        }
        .header h1 { 
            margin: 0 0 5px 0; 
            font-size: 48px; 
            font-weight: 700; 
        }
        .header h1 .highlight {
            color: #ff8c00;
        }
        .header h1 .dark {
            color: #000;
        }
        .brush-underline {
            width: 300px;
            height: 8px;
            margin: 5px auto 0;
            background: #ff8c00;
            position: relative;
            border-radius: 50%;
            transform: scaleY(0.6);
        }
        .brush-underline::before {
            content: '';
            position: absolute;
            left: -3px;
            top: -1px;
            width: 100%;
            height: 100%;
            background: inherit;
            border-radius: 50%;
            opacity: 0.4;
        }
        .brush-underline::after {
            content: '';
            position: absolute;
            right: -3px;
            top: 1px;
            width: 100%;
            height: 100%;
            background: inherit;
            border-radius: 50%;
            opacity: 0.3;
        }
        
        .badge {
            display: inline-block;
            background: #2c3e50;
            color: white;
            padding: 8px 20px;
            border-radius: 6px;
            font-size: 11px;
            font-weight: 600;
            letter-spacing: 1px;
            text-transform: uppercase;
            margin-bottom: 30px;
        }
        
        .content { 
            text-align: center; 
        }
        h2 { 
            font-size: 42px; 
            font-weight: 700; 
            margin-bottom: 10px; 
            color: #000;
        }
        .title-underline {
            width: 60px;
            height: 4px;
            background: #ff8c00;
            margin: 0 auto 30px;
            border-radius: 2px;
        }
        
        p { 
            margin-bottom: 15px; 
            font-size: 15px; 
            color: #999;
            line-height: 1.6;
        }
        p.greeting {
            font-size: 24px;
            color: #333;
            font-weight: 600;
            margin-bottom: 15px;
        }
        
        .button { 
            display: inline-block; 
            background: white; 
            color: #000; 
            padding: 18px 50px; 
            text-decoration: none; 
            border: 3px solid #000; 
            border-radius: 12px; 
            margin: 40px 0 20px; 
            font-weight: 700; 
            font-size: 18px;
            text-transform: uppercase;
            position: relative;
        }
        .button-wrapper {
            display: inline-block;
            position: relative;
        }
        .button-shadow {
            position: absolute;
            top: 6px;
            left: 6px;
            right: -6px;
            bottom: -6px;
            background: #ff8c00;
            border-radius: 12px;
            z-index: -1;
        }
        
        .warning-box {
            background: white;
            border: 2px solid #ffa500;
            border-radius: 12px;
            padding: 16px 20px;
            margin: 20px auto 30px;
            display: block;
            text-align: center;
            max-width: 450px;
        }
        .warning-content {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 12px;
        }
        .warning-icon {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            width: 24px;
            height: 24px;
            min-width: 24px;
            background: #ffa500;
            border-radius: 50%;
            color: white;
            font-weight: bold;
            font-size: 14px;
            flex-shrink: 0;
            line-height: 1;
        }
        .warning-text {
            display: inline-block;
            color: #666;
            font-size: 14px;
            text-align: left;
            line-height: 1.5;
        }
        
        .security-box {
            background: #fff7ed;
            border: 3px solid #ff8c00;
            border-radius: 12px;
            padding: 20px;
            margin: 28px 0;
            text-align: left;
        }
        .security-title {
            font-weight: 700;
            font-size: 16px;
            margin-bottom: 8px;
            color: #000;
        }
        .security-text {
            margin: 0;
            font-size: 14px;
            color: #666;
        }
        
        .footer { 
            text-align: center; 
            margin-top: 40px; 
            font-size: 13px; 
            color: #666; 
            padding-top: 0;
            line-height: 1.8;
        }
        .footer p {
            margin: 8px 0;
        }
        .footer-bold {
            font-weight: 600;
            color: #333;
        }
        .support-link {
            color: #ff8c00;
            text-decoration: underline;
        }
        
        @media (max-width: 600px) {
            .container {
                padding: 40px 25px;
            }
            .header h1 {
                font-size: 36px;
            }
            h2 {
                font-size: 32px;
            }
            .button {
                padding: 15px 40px;
                font-size: 16px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="content-wrapper">
            <div class="header">
                <h1><span class="highlight">Ngasih</span><span class="dark">Tau</span></h1>
                <div class="brush-underline"></div>
            </div>
            
            <div class="content">
                <div class="badge">Password Reset</div>
                <h2>Reset Your Password</h2>
                <div class="title-underline"></div>
                
                <p class="greeting">Hi {{.RecipientName}},</p>
                <p>We received a request to reset your password. Click the button below to create a new secure password:</p>
                
                <p style="text-align: center;">
                    <span class="button-wrapper">
                        <span class="button-shadow"></span>
                        <a href="{{.ActionURL}}" class="button">Reset Password</a>
                    </span>
                </p>
                
                <div class="warning-box">
                    <div class="warning-content">
                        <span class="warning-icon">⏱</span>
                        <span class="warning-text">This link will expire in {{.ExpiryHours}} hour(s). Act quickly to secure your account.</span>
                    </div>
                </div>
                
                <div class="security-box">
                    <div class="security-title">🔒 Security Notice</div>
                    <p class="security-text">Didn't request a password reset? You can safely ignore this email. Your password will remain unchanged and your account stays secure.</p>
                </div>
            </div>
            
            <div class="footer">
                <p class="footer-bold">© {{.AppName}}. All rights reserved.</p>
                <p>Need help? Contact us at <a href="mailto:{{.SupportEmail}}" class="support-link">{{.SupportEmail}}</a></p>
            </div>
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
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Inter', sans-serif; 
            line-height: 1.6; 
            color: #1e293b; 
            max-width: 600px; 
            margin: 0 auto; 
            padding: 20px; 
            background-color: #FDFCF9;
        }
        .container { 
            background-color: #fffbf5;
            background-image: radial-gradient(circle, rgba(232, 220, 200, 0.3) 1.5px, transparent 1.5px);
            background-size: 20px 20px;
            border: 3px solid #000; 
            border-radius: 24px; 
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15); 
            padding: 60px 40px; 
            margin-top: 20px; 
            position: relative;
            overflow: hidden;
        }
        
        .content-wrapper {
            position: relative;
            z-index: 1;
        }
        
        .header { 
            text-align: center; 
            margin-bottom: 40px; 
        }
        .header h1 { 
            margin: 0 0 5px 0; 
            font-size: 48px; 
            font-weight: 700; 
        }
        .header h1 .highlight {
            color: #ff8c00;
        }
        .header h1 .dark {
            color: #000;
        }
        .brush-underline {
            width: 300px;
            height: 8px;
            margin: 5px auto 0;
            background: #ff8c00;
            position: relative;
            border-radius: 50%;
            transform: scaleY(0.6);
        }
        .brush-underline::before {
            content: '';
            position: absolute;
            left: -3px;
            top: -1px;
            width: 100%;
            height: 100%;
            background: inherit;
            border-radius: 50%;
            opacity: 0.4;
        }
        .brush-underline::after {
            content: '';
            position: absolute;
            right: -3px;
            top: 1px;
            width: 100%;
            height: 100%;
            background: inherit;
            border-radius: 50%;
            opacity: 0.3;
        }
        
        .badge {
            display: inline-block;
            background: #2c3e50;
            color: white;
            padding: 8px 20px;
            border-radius: 6px;
            font-size: 11px;
            font-weight: 600;
            letter-spacing: 1px;
            text-transform: uppercase;
            margin-bottom: 30px;
        }
        
        .content { 
            text-align: center; 
        }
        h2 { 
            font-size: 42px; 
            font-weight: 700; 
            margin-bottom: 10px; 
            color: #000;
        }
        .title-underline {
            width: 60px;
            height: 4px;
            background: #ff8c00;
            margin: 0 auto 30px;
            border-radius: 2px;
        }
        
        p { 
            margin-bottom: 15px; 
            font-size: 15px; 
            color: #999;
            line-height: 1.6;
        }
        p.greeting {
            font-size: 24px;
            color: #333;
            font-weight: 600;
            margin-bottom: 15px;
        }
        
        .button { 
            display: inline-block; 
            background: white; 
            color: #000; 
            padding: 18px 50px; 
            text-decoration: none; 
            border: 3px solid #000; 
            border-radius: 12px; 
            margin: 40px 0 30px; 
            font-weight: 700; 
            font-size: 18px;
            text-transform: uppercase;
            position: relative;
        }
        .button-wrapper {
            display: inline-block;
            position: relative;
        }
        .button-shadow {
            position: absolute;
            top: 6px;
            left: 6px;
            right: -6px;
            bottom: -6px;
            background: #ff8c00;
            border-radius: 12px;
            z-index: -1;
        }
        
        .pod-info { 
            background: white; 
            border: 3px solid #000; 
            border-radius: 16px; 
            padding: 28px; 
            margin: 28px auto; 
            box-shadow: 6px 6px 0 rgba(255, 140, 0, 0.3);
            text-align: left;
            max-width: 400px;
        }
        .pod-label {
            font-size: 11px;
            font-weight: 700;
            color: #ff8c00;
            text-transform: uppercase;
            letter-spacing: 1.5px;
            margin-bottom: 12px;
        }
        .pod-name { 
            font-size: 24px; 
            font-weight: 800; 
            color: #000; 
            letter-spacing: -0.5px;
            line-height: 1.3;
            margin: 0 0 16px 0;
        }
        .inviter-info {
            padding-top: 16px;
            border-top: 2px solid #f1f5f9;
            font-size: 14px;
            color: #999;
        }
        .inviter-name {
            font-weight: 700;
            color: #333;
        }
        
        .ignore-text {
            color: #999;
            font-size: 14px;
            margin-top: 32px;
        }
        
        .footer { 
            text-align: center; 
            margin-top: 40px; 
            font-size: 13px; 
            color: #666; 
            padding-top: 0;
            line-height: 1.8;
        }
        .footer p {
            margin: 8px 0;
        }
        .footer-bold {
            font-weight: 600;
            color: #333;
        }
        .support-link {
            color: #ff8c00;
            text-decoration: underline;
        }
        
        @media (max-width: 600px) {
            .container {
                padding: 40px 25px;
            }
            .header h1 {
                font-size: 36px;
            }
            h2 {
                font-size: 32px;
            }
            .button {
                padding: 15px 40px;
                font-size: 16px;
            }
            .pod-info {
                padding: 20px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="content-wrapper">
            <div class="header">
                <h1><span class="highlight">Ngasih</span><span class="dark">Tau</span></h1>
                <div class="brush-underline"></div>
            </div>
            
            <div class="content">
                <div class="badge">Collaboration Invite</div>
                <h2>You're Invited! 🎉</h2>
                <div class="title-underline"></div>
                
                <p class="greeting">Hi {{.RecipientName}},</p>
                <p>You've been invited to collaborate on a Knowledge Pod. Join the team and start sharing knowledge!</p>
                
                <div class="pod-info">
                    <div class="pod-label">📚 Knowledge Pod</div>
                    <h3 class="pod-name">{{.PodName}}</h3>
                    <div class="inviter-info">
                        Invited by <span class="inviter-name">{{.InviterName}}</span>
                    </div>
                </div>
                
                <p>Click below to accept the invitation and start collaborating:</p>
                
                <p style="text-align: center;">
                    <span class="button-wrapper">
                        <span class="button-shadow"></span>
                        <a href="{{.ActionURL}}" class="button">Accept Invite</a>
                    </span>
                </p>
                
                <p class="ignore-text">Not interested? You can safely ignore this email.</p>
            </div>
            
            <div class="footer">
                <p class="footer-bold">© {{.AppName}}. All rights reserved.</p>
                <p>Need help? Contact us at <a href="mailto:{{.SupportEmail}}" class="support-link">{{.SupportEmail}}</a></p>
            </div>
        </div>
    </div>
</body>
</html>`