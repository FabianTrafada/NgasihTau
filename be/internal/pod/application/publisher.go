// Package application contains the business logic for the pod service.
package application

import (
	"context"

	"ngasihtau/internal/pod/domain"
	natspkg "ngasihtau/pkg/nats"
)

// NATSEventPublisher implements EventPublisher using NATS.
type NATSEventPublisher struct {
	publisher *natspkg.Publisher
}

// NewNATSEventPublisher creates a new NATS event publisher.
func NewNATSEventPublisher(publisher *natspkg.Publisher) *NATSEventPublisher {
	return &NATSEventPublisher{
		publisher: publisher,
	}
}

// PublishPodCreated publishes a pod.created event.
func (p *NATSEventPublisher) PublishPodCreated(ctx context.Context, pod *domain.Pod) error {
	description := ""
	if pod.Description != nil {
		description = *pod.Description
	}

	event := natspkg.PodCreatedEvent{
		PodID:       pod.ID,
		OwnerID:     pod.OwnerID,
		Name:        pod.Name,
		Slug:        pod.Slug,
		Description: description,
		Visibility:  string(pod.Visibility),
		Categories:  pod.Categories,
		Tags:        pod.Tags,
	}

	return p.publisher.PublishPodCreated(ctx, event)
}

// PublishPodUpdated publishes a pod.updated event.
func (p *NATSEventPublisher) PublishPodUpdated(ctx context.Context, pod *domain.Pod) error {
	description := ""
	if pod.Description != nil {
		description = *pod.Description
	}

	event := natspkg.PodUpdatedEvent{
		PodID:       pod.ID,
		OwnerID:     pod.OwnerID,
		Name:        pod.Name,
		Slug:        pod.Slug,
		Description: description,
		Visibility:  string(pod.Visibility),
		Categories:  pod.Categories,
		Tags:        pod.Tags,
		StarCount:   pod.StarCount,
		ForkCount:   pod.ForkCount,
		ViewCount:   pod.ViewCount,
	}

	return p.publisher.PublishPodUpdated(ctx, event)
}

// PublishCollaboratorInvited publishes a collaborator.invited event.
func (p *NATSEventPublisher) PublishCollaboratorInvited(ctx context.Context, collaborator *domain.Collaborator, podName string) error {
	event := natspkg.CollaboratorInvitedEvent{
		PodID:     collaborator.PodID,
		InviterID: collaborator.InvitedBy,
		InviteeID: collaborator.UserID,
		Role:      string(collaborator.Role),
	}

	return p.publisher.PublishCollaboratorInvited(ctx, event)
}

// NoOpEventPublisher is a no-op implementation of EventPublisher.
type NoOpEventPublisher struct{}

// NewNoOpEventPublisher creates a new no-op event publisher.
func NewNoOpEventPublisher() *NoOpEventPublisher {
	return &NoOpEventPublisher{}
}

// PublishPodCreated is a no-op.
func (p *NoOpEventPublisher) PublishPodCreated(ctx context.Context, pod *domain.Pod) error {
	return nil
}

// PublishPodUpdated is a no-op.
func (p *NoOpEventPublisher) PublishPodUpdated(ctx context.Context, pod *domain.Pod) error {
	return nil
}

// PublishCollaboratorInvited is a no-op.
func (p *NoOpEventPublisher) PublishCollaboratorInvited(ctx context.Context, collaborator *domain.Collaborator, podName string) error {
	return nil
}
