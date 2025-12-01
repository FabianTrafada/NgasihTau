// Package postgres provides PostgreSQL implementations of the Pod Service repositories.
package postgres

import (
	"ngasihtau/internal/pod/domain"
)

// Ensure repository implementations satisfy their interfaces.
var (
	_ domain.PodRepository          = (*PodRepository)(nil)
	_ domain.CollaboratorRepository = (*CollaboratorRepository)(nil)
	_ domain.PodStarRepository      = (*PodStarRepository)(nil)
	_ domain.PodFollowRepository    = (*PodFollowRepository)(nil)
	_ domain.ActivityRepository     = (*ActivityRepository)(nil)
)
