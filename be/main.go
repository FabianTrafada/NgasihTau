// Package main is a placeholder for swagger documentation generation.
// This file imports all handlers to help swag find all API endpoints.
// It is NOT used to run any service - each service has its own main.go in cmd/.
//
// @title NgasihTau API
// @version 1.0
// @description NgasihTau is a learning platform API that enables teachers to create Knowledge Pods for sharing learning materials with students.
// @description The platform uses a microservices architecture with Go as the primary backend language.
//
// @contact.name NgasihTau API Support
// @contact.email support@ngasihtau.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8001
// @BasePath /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token. Format: "Bearer {token}"
//
// @tag.name Auth
// @tag.description Authentication and authorization endpoints
//
// @tag.name Users
// @tag.description User profile and social features
//
// @tag.name Pods
// @tag.description Knowledge Pod management and collaboration
//
// @tag.name Materials
// @tag.description Learning material upload, versioning, and management
//
// @tag.name Comments
// @tag.description Material comments and discussions
//
// @tag.name Ratings
// @tag.description Material ratings and reviews
//
// @tag.name Bookmarks
// @tag.description Material bookmarking
//
// @tag.name Search
// @tag.description Full-text and semantic search
//
// @tag.name AI
// @tag.description AI-powered chat and Q&A
//
// @tag.name Notifications
// @tag.description In-app and email notifications
package main

import (
	_ "ngasihtau/docs"
	_ "ngasihtau/internal/ai/interfaces/http"
	_ "ngasihtau/internal/material/interfaces/http"
	_ "ngasihtau/internal/notification/interfaces/http"
	_ "ngasihtau/internal/pod/interfaces/http"
	_ "ngasihtau/internal/search/interfaces/http"
	_ "ngasihtau/internal/user/interfaces/http"
)

func main() {
	// This file is only used for swagger generation
}
