// Package postgres provides PostgreSQL implementations of the User Service repositories.
package postgres

import "ngasihtau/internal/user/domain"

// Compile-time interface implementation checks.
// These ensure that our concrete types properly implement the domain interfaces.
var (
	_ domain.UserRepository              = (*UserRepository)(nil)
	_ domain.OAuthAccountRepository      = (*OAuthAccountRepository)(nil)
	_ domain.RefreshTokenRepository      = (*RefreshTokenRepository)(nil)
	_ domain.BackupCodeRepository        = (*BackupCodeRepository)(nil)
	_ domain.FollowRepository            = (*FollowRepository)(nil)
	_ domain.VerificationTokenRepository = (*VerificationTokenRepository)(nil)
)
