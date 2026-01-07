# Traefik Routes Mapping

Dokumentasi mapping antara Traefik routes dan Backend endpoints.

## User Service (Port 8001)

### Authentication Routes
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login with email/password
- `POST /api/v1/auth/google` - Google OAuth login
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout user
- `POST /api/v1/auth/2fa/login` - Verify 2FA login
- `POST /api/v1/auth/verify-email` - Verify email with token
- `POST /api/v1/auth/password/forgot` - Request password reset
- `POST /api/v1/auth/password/reset` - Reset password with token
- `POST /api/v1/auth/2fa/enable` - Enable 2FA (protected)
- `POST /api/v1/auth/2fa/verify` - Verify 2FA (protected)
- `POST /api/v1/auth/2fa/disable` - Disable 2FA (protected)
- `POST /api/v1/auth/send-verification` - Send verification email (protected)

### User Profile Routes
- `GET /api/v1/users/me` - Get current user profile (protected)
- `PUT /api/v1/users/me` - Update current user profile (protected)
- `GET /api/v1/users/:id` - Get user public profile
- `POST /api/v1/users/:id/follow` - Follow user (protected)
- `DELETE /api/v1/users/:id/follow` - Unfollow user (protected)
- `GET /api/v1/users/:id/followers` - Get user followers
- `GET /api/v1/users/:id/following` - Get user following

### Teacher Verification Routes
- `POST /api/v1/users/verification/teacher` - Submit verification (protected)
- `GET /api/v1/users/verification/status` - Get verification status (protected)

### Admin Routes
- `GET /api/v1/admin/verifications` - List pending verifications (protected)
- `POST /api/v1/admin/verifications/:id/approve` - Approve verification (protected)
- `POST /api/v1/admin/verifications/:id/reject` - Reject verification (protected)

### Learning Interests Routes
- `GET /api/v1/interests/predefined` - Get predefined interests
- `GET /api/v1/interests/predefined/categories` - Get interests by category
- `GET /api/v1/interests/me` - Get user's interests (protected)
- `PUT /api/v1/interests/me` - Set/replace all interests (protected)
- `POST /api/v1/interests/me` - Add single interest (protected)
- `DELETE /api/v1/interests/me/:id` - Remove single interest (protected)
- `GET /api/v1/interests/onboarding/status` - Check onboarding status (protected)
- `POST /api/v1/interests/onboarding/complete` - Complete onboarding (protected)

### API Documentation
- `GET /api/docs/*` - Swagger UI
- `GET /api/openapi.json` - OpenAPI JSON spec
- `GET /api/openapi.yaml` - OpenAPI YAML spec

**Traefik Routes:**
- `user-service-auth` - Priority 100 - `/api/v1/auth`
- `user-service-users` - Priority 90 - `/api/v1/users`
- `user-service-admin` - Priority 95 - `/api/v1/admin`
- `user-service-interests` - Priority 90 - `/api/v1/interests`
- `user-followers` - Priority 100 - `/api/v1/users/:id/followers`
- `user-following` - Priority 100 - `/api/v1/users/:id/following`
- `api-docs` - Priority 50 - `/api/docs` and `/api/openapi`

---

## Pod Service (Port 8002)

### Pod CRUD Routes
- `POST /api/v1/pods` - Create pod (protected)
- `GET /api/v1/pods` - List pods with filters
- `GET /api/v1/pods/:id` - Get pod details
- `PUT /api/v1/pods/:id` - Update pod (protected, requires edit access)
- `DELETE /api/v1/pods/:id` - Delete pod (protected, requires owner access)

### Pod Interaction Routes
- `POST /api/v1/pods/:id/fork` - Fork pod (protected)
- `POST /api/v1/pods/:id/star` - Star pod (protected)
- `DELETE /api/v1/pods/:id/star` - Unstar pod (protected)
- `POST /api/v1/pods/:id/upvote` - Upvote pod (protected)
- `DELETE /api/v1/pods/:id/upvote` - Remove upvote (protected)
- `POST /api/v1/pods/:id/follow` - Follow pod (protected)
- `DELETE /api/v1/pods/:id/follow` - Unfollow pod (protected)

### Upload Request Routes (Teacher Collaboration)
- `POST /api/v1/pods/:id/upload-request` - Create upload request (protected)
- `GET /api/v1/users/me/upload-requests` - Get user's upload requests (protected)
- `POST /api/v1/upload-requests/:id/approve` - Approve upload request (protected)
- `POST /api/v1/upload-requests/:id/reject` - Reject upload request (protected)
- `DELETE /api/v1/upload-requests/:id` - Revoke upload permission (protected)

### Shared Pods Routes (Teacher-Student)
- `POST /api/v1/pods/:id/share` - Share pod with student (protected)
- `GET /api/v1/users/me/shared-pods` - Get shared pods (protected)

### Collaborators Routes
- `GET /api/v1/pods/:id/collaborators` - Get collaborators
- `POST /api/v1/pods/:id/collaborators` - Invite collaborator (protected)
- `PUT /api/v1/pods/:id/collaborators/:userId` - Update collaborator (protected)
- `DELETE /api/v1/pods/:id/collaborators/:userId` - Remove collaborator (protected)

### Activity Routes
- `GET /api/v1/pods/:id/activity` - Get pod activity
- `GET /api/v1/feed` - Get user's activity feed (protected)

### Recommendation Routes
- `POST /api/v1/pods/:id/track` - Track interaction (protected)
- `POST /api/v1/pods/:id/track/time` - Track time spent (protected)
- `GET /api/v1/pods/:id/similar` - Get similar pods
- `GET /api/v1/feed/recommended` - Get personalized feed (protected)
- `GET /api/v1/feed/trending` - Get trending feed
- `GET /api/v1/users/me/preferences` - Get user preferences (protected)

### User's Pods Routes
- `GET /api/v1/users/:id/pods` - Get user's pods
- `GET /api/v1/users/:id/starred` - Get user's starred pods
- `GET /api/v1/users/me/upvoted-pods` - Get user's upvoted pods (protected)

**Traefik Routes:**
- `pod-service-user-preferences` - Priority 100 - `/api/v1/users/me/preferences`
- `pod-service-feed` - Priority 95 - `/api/v1/feed`
- `user-pods` - Priority 100 - `/api/v1/users/:id/pods`
- `user-starred` - Priority 100 - `/api/v1/users/:id/starred`
- `user-me-upvoted-pods` - Priority 110 - `/api/v1/users/me/upvoted-pods`
- `user-me-upload-requests` - Priority 110 - `/api/v1/users/me/upload-requests`
- `user-me-shared-pods` - Priority 110 - `/api/v1/users/me/shared-pods`
- `pod-service-upload-requests` - Priority 95 - `/api/v1/upload-requests`
- `pod-service` - Priority 80 - `/api/v1/pods`

---

## Material Service (Port 8003)

### Material Upload Routes
- `POST /api/v1/materials/upload-url` - Get upload URL (protected)
- `POST /api/v1/materials/confirm` - Confirm upload (protected)
- `GET /api/v1/materials/:id` - Get material details (protected)
- `PUT /api/v1/materials/:id` - Update material (protected)
- `DELETE /api/v1/materials/:id` - Delete material (protected)
- `GET /api/v1/materials/:id/preview` - Get preview URL (protected)
- `GET /api/v1/materials/:id/download` - Get download URL (protected)

### Material Version Routes
- `POST /api/v1/materials/:id/versions` - Create version (protected)
- `GET /api/v1/materials/:id/versions` - Get version history (protected)
- `POST /api/v1/materials/:id/versions/:version/restore` - Restore version (protected)

### Comment Routes
- `POST /api/v1/materials/:id/comments` - Add comment (protected)
- `GET /api/v1/materials/:id/comments` - Get comments (protected)
- `PUT /api/v1/comments/:id` - Update comment (protected)
- `DELETE /api/v1/comments/:id` - Delete comment (protected)

### Rating Routes
- `POST /api/v1/materials/:id/ratings` - Rate material (protected)
- `GET /api/v1/materials/:id/ratings` - Get ratings (protected)

### Bookmark Routes
- `POST /api/v1/materials/:id/bookmark` - Bookmark material (protected)
- `DELETE /api/v1/materials/:id/bookmark` - Remove bookmark (protected)
- `GET /api/v1/bookmarks` - Get bookmarks (protected)
- `GET /api/v1/bookmarks/folders` - Get bookmark folders (protected)

### Pod Materials Routes
- `GET /api/v1/pods/:podId/materials` - Get materials in pod (protected)

**Traefik Routes:**
- `material-service-upload` - Priority 110 - `/api/v1/materials/upload-url` and `/api/v1/materials/confirm`
- `material-service-bookmarks` - Priority 95 - `/api/v1/bookmarks`
- `material-service-comments` - Priority 95 - `/api/v1/comments`
- `pod-materials` - Priority 100 - `/api/v1/pods/:id/materials`
- `material-service` - Priority 70 - `/api/v1/materials`

---

## Search Service (Port 8004)

### Search Routes
- `GET /api/v1/search` - Full-text search (optional auth)
- `GET /api/v1/search/semantic` - Semantic search
- `GET /api/v1/search/hybrid` - Hybrid search (optional auth)
- `GET /api/v1/search/suggestions` - Get search suggestions
- `GET /api/v1/search/trending` - Get trending searches
- `GET /api/v1/search/popular` - Get popular searches
- `GET /api/v1/search/history` - Get search history (protected)
- `DELETE /api/v1/search/history` - Clear search history (protected)

**Traefik Routes:**
- `search-service` - Priority 90 - `/api/v1/search`

---

## AI Service (Port 8005)

### Material Chat Routes
- `POST /api/v1/materials/:id/chat` - Chat with material (protected)
- `GET /api/v1/materials/:id/chat/history` - Get chat history (protected)
- `GET /api/v1/materials/:id/chat/suggestions` - Get suggestions (protected)
- `POST /api/v1/materials/:id/chat/export` - Export chat (protected)

### Pod Chat Routes
- `POST /api/v1/pods/:id/chat` - Chat with pod context (protected)

### Feedback Routes
- `POST /api/v1/chat/:messageId/feedback` - Submit chat feedback (protected)

**Traefik Routes:**
- `ai-service-material-chat` - Priority 120 - `/api/v1/materials/:id/chat`
- `ai-service-pod-chat` - Priority 120 - `/api/v1/pods/:id/chat`
- `ai-service-feedback` - Priority 95 - `/api/v1/chat`

---

## Notification Service (Port 8006)

### Notification Routes
- `GET /api/v1/notifications` - Get notifications (protected)
- `PUT /api/v1/notifications/:id/read` - Mark as read (protected)
- `PUT /api/v1/notifications/read-all` - Mark all as read (protected)
- `GET /api/v1/notifications/preferences` - Get preferences (protected)
- `PUT /api/v1/notifications/preferences` - Update preferences (protected)

**Traefik Routes:**
- `notification-service` - Priority 90 - `/api/v1/notifications`

---

## Route Priority Guide

Higher priority routes are matched first:
- **120** - AI chat endpoints (specific paths with regex)
- **110** - Material upload endpoints, user "me" specific routes
- **100** - Specific user paths (followers, pods, starred), pod-service preferences
- **95** - Feed, bookmarks, comments, upload-requests, admin routes, chat feedback
- **90** - General service prefixes (users, interests, notifications, search)
- **80** - General pod service routes
- **70** - General material service routes
- **50** - API documentation

## Notes

1. All routes dengan prefix `/api/v1/users/:id/` (kecuali `/me/`) harus dicek service tujuannya
2. Routes dengan priority lebih tinggi akan diproses duluan oleh Traefik
3. Semua protected routes menggunakan JWT authentication middleware di backend
4. CORS dan security headers diterapkan di semua routes via Traefik middleware
