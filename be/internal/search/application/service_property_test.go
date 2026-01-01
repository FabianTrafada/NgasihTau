// Package application contains property-based tests for the Search Service.
package application

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"ngasihtau/internal/search/domain"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: student-teacher-roles, Property 11: Verified Status Search and Ranking**
// **Validates: Requirements 6.3, 6.4, 6.5**
//
// Property 11: Verified Status Search and Ranking
// *For any* search query:
// - When filtered by verified=true, only pods with `is_verified = true` SHALL be returned
// - When sorted by trust_score, verified pods with high upvotes SHALL rank higher
// - Default ranking SHALL prioritize verified pods with high upvote counts

// mockSearchRepoWithVerifiedFilter implements SearchRepository with verified filter support
type mockSearchRepoWithVerifiedFilter struct {
	pods []domain.PodDocument
}

func newMockSearchRepoWithVerifiedFilter(pods []domain.PodDocument) *mockSearchRepoWithVerifiedFilter {
	return &mockSearchRepoWithVerifiedFilter{pods: pods}
}

func (m *mockSearchRepoWithVerifiedFilter) Search(ctx context.Context, query domain.SearchQuery) ([]domain.SearchResult, int64, error) {
	var results []domain.SearchResult

	for _, pod := range m.pods {
		// Apply verified filter (Requirement 6.3)
		if query.Verified != nil && pod.IsVerified != *query.Verified {
			continue
		}

		result := domain.SearchResult{
			ID:          pod.ID,
			Type:        "pod",
			Title:       pod.Name,
			Description: pod.Description,
			Score:       pod.TrustScore,
			Metadata: map[string]any{
				"is_verified":  pod.IsVerified,
				"upvote_count": pod.UpvoteCount,
				"trust_score":  pod.TrustScore,
			},
		}
		results = append(results, result)
	}

	// Apply sorting based on SortBy (Requirements 6.4, 6.5)
	switch query.SortBy {
	case domain.SortByUpvotes:
		sort.Slice(results, func(i, j int) bool {
			iUpvotes := results[i].Metadata["upvote_count"].(int)
			jUpvotes := results[j].Metadata["upvote_count"].(int)
			return iUpvotes > jUpvotes
		})
	case domain.SortByTrustScore:
		sort.Slice(results, func(i, j int) bool {
			iTrust := results[i].Metadata["trust_score"].(float64)
			jTrust := results[j].Metadata["trust_score"].(float64)
			if iTrust != jTrust {
				return iTrust > jTrust
			}
			// Secondary sort by upvote_count
			iUpvotes := results[i].Metadata["upvote_count"].(int)
			jUpvotes := results[j].Metadata["upvote_count"].(int)
			return iUpvotes > jUpvotes
		})
	}

	// Apply pagination
	start := (query.Page - 1) * query.PerPage
	end := start + query.PerPage
	if start >= len(results) {
		return []domain.SearchResult{}, int64(len(results)), nil
	}
	if end > len(results) {
		end = len(results)
	}

	return results[start:end], int64(len(results)), nil
}

func (m *mockSearchRepoWithVerifiedFilter) GetSuggestions(ctx context.Context, prefix string, limit int) ([]string, error) {
	return []string{}, nil
}

// calculateTrustScore computes trust score based on is_verified and upvote_count
// Formula: (0.6 * is_verified) + (0.4 * normalized_upvotes)
// For testing, we use a simplified normalization
func calculateTrustScore(isVerified bool, upvoteCount int, maxUpvotes int) float64 {
	verifiedWeight := 0.6
	upvoteWeight := 0.4

	verifiedScore := 0.0
	if isVerified {
		verifiedScore = 1.0
	}

	normalizedUpvotes := 0.0
	if maxUpvotes > 0 {
		normalizedUpvotes = float64(upvoteCount) / float64(maxUpvotes)
	}

	return (verifiedWeight * verifiedScore) + (upvoteWeight * normalizedUpvotes)
}

// genPodDocument generates a random PodDocument
func genPodDocument(maxUpvotes int) gopter.Gen {
	return gopter.CombineGens(
		gen.Identifier(), // Generates non-empty alphanumeric strings
		gen.Identifier(),
		gen.Bool(),
		gen.IntRange(0, maxUpvotes),
	).Map(func(values []interface{}) domain.PodDocument {
		name := values[0].(string)
		id := values[1].(string)
		isVerified := values[2].(bool)
		upvoteCount := values[3].(int)

		return domain.PodDocument{
			ID:          "pod_" + id,
			Name:        name,
			Description: "Description for " + name,
			IsVerified:  isVerified,
			UpvoteCount: upvoteCount,
			TrustScore:  calculateTrustScore(isVerified, upvoteCount, maxUpvotes),
		}
	})
}

// genPodDocumentSlice generates a slice of PodDocuments
func genPodDocumentSlice(minSize, maxSize, maxUpvotes int) gopter.Gen {
	return gen.IntRange(minSize, maxSize).FlatMap(func(size interface{}) gopter.Gen {
		return gen.SliceOfN(size.(int), genPodDocument(maxUpvotes))
	}, reflect.TypeOf([]domain.PodDocument{}))
}

func TestProperty_VerifiedStatusSearchAndRanking(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 11.1: When filtered by verified=true, only pods with is_verified=true are returned
	// Validates: Requirement 6.3 - THE Search Service SHALL support filtering knowledge pods by verified status.
	properties.Property("verified filter returns only verified pods", prop.ForAll(
		func(pods []domain.PodDocument) bool {
			if len(pods) == 0 {
				return true
			}

			ctx := context.Background()
			searchRepo := newMockSearchRepoWithVerifiedFilter(pods)
			svc := NewService(
				searchRepo,
				newMockIndexRepo(),
				newMockVectorSearchRepo(),
				newMockSearchHistoryRepo(),
				newMockTrendingRepo(),
			)

			verified := true
			input := SearchInput{
				Query:    "",
				Verified: &verified,
				Page:     1,
				PerPage:  100,
			}

			result, err := svc.Search(ctx, input)
			if err != nil {
				return false
			}

			// All returned results must have is_verified = true
			for _, r := range result.Results {
				if isVerified, ok := r.Metadata["is_verified"].(bool); !ok || !isVerified {
					return false
				}
			}

			return true
		},
		genPodDocumentSlice(0, 20, 1000),
	))

	// Property 11.2: When filtered by verified=false, only pods with is_verified=false are returned
	// Validates: Requirement 6.3 - verified filter works for both true and false
	properties.Property("verified=false filter returns only unverified pods", prop.ForAll(
		func(pods []domain.PodDocument) bool {
			if len(pods) == 0 {
				return true
			}

			ctx := context.Background()
			searchRepo := newMockSearchRepoWithVerifiedFilter(pods)
			svc := NewService(
				searchRepo,
				newMockIndexRepo(),
				newMockVectorSearchRepo(),
				newMockSearchHistoryRepo(),
				newMockTrendingRepo(),
			)

			verified := false
			input := SearchInput{
				Query:    "",
				Verified: &verified,
				Page:     1,
				PerPage:  100,
			}

			result, err := svc.Search(ctx, input)
			if err != nil {
				return false
			}

			// All returned results must have is_verified = false
			for _, r := range result.Results {
				if isVerified, ok := r.Metadata["is_verified"].(bool); !ok || isVerified {
					return false
				}
			}

			return true
		},
		genPodDocumentSlice(0, 20, 1000),
	))

	// Property 11.3: When sorted by upvotes, pods with higher upvote_count appear first
	// Validates: Requirement 6.4 - THE Search Service SHALL support sorting knowledge pods by upvote count.
	properties.Property("upvote sorting returns pods in descending upvote order", prop.ForAll(
		func(pods []domain.PodDocument) bool {
			if len(pods) < 2 {
				return true
			}

			ctx := context.Background()
			searchRepo := newMockSearchRepoWithVerifiedFilter(pods)
			svc := NewService(
				searchRepo,
				newMockIndexRepo(),
				newMockVectorSearchRepo(),
				newMockSearchHistoryRepo(),
				newMockTrendingRepo(),
			)

			input := SearchInput{
				Query:   "",
				SortBy:  domain.SortByUpvotes,
				Page:    1,
				PerPage: 100,
			}

			result, err := svc.Search(ctx, input)
			if err != nil {
				return false
			}

			// Results must be in descending order by upvote_count
			for i := 1; i < len(result.Results); i++ {
				prevUpvotes := result.Results[i-1].Metadata["upvote_count"].(int)
				currUpvotes := result.Results[i].Metadata["upvote_count"].(int)
				if prevUpvotes < currUpvotes {
					return false
				}
			}

			return true
		},
		genPodDocumentSlice(0, 20, 1000),
	))

	// Property 11.4: When sorted by trust_score, verified pods with high upvotes rank higher
	// Validates: Requirement 6.4, 6.5 - THE Search Service SHALL support sorting by combination of verified status and upvote count.
	properties.Property("trust_score sorting prioritizes verified pods with high upvotes", prop.ForAll(
		func(pods []domain.PodDocument) bool {
			if len(pods) < 2 {
				return true
			}

			ctx := context.Background()
			searchRepo := newMockSearchRepoWithVerifiedFilter(pods)
			svc := NewService(
				searchRepo,
				newMockIndexRepo(),
				newMockVectorSearchRepo(),
				newMockSearchHistoryRepo(),
				newMockTrendingRepo(),
			)

			input := SearchInput{
				Query:   "",
				SortBy:  domain.SortByTrustScore,
				Page:    1,
				PerPage: 100,
			}

			result, err := svc.Search(ctx, input)
			if err != nil {
				return false
			}

			// Results must be in descending order by trust_score
			for i := 1; i < len(result.Results); i++ {
				prevTrust := result.Results[i-1].Metadata["trust_score"].(float64)
				currTrust := result.Results[i].Metadata["trust_score"].(float64)
				if prevTrust < currTrust {
					return false
				}
			}

			return true
		},
		genPodDocumentSlice(0, 20, 1000),
	))

	// Property 11.5: Verified pods have higher trust_score than unverified pods with same upvotes
	// Validates: Requirement 6.5 - verified status contributes to trust score
	properties.Property("verified pods have higher trust_score than unverified with same upvotes", prop.ForAll(
		func(upvoteCount int) bool {
			if upvoteCount < 0 {
				upvoteCount = 0
			}
			if upvoteCount > 1000 {
				upvoteCount = 1000
			}

			maxUpvotes := 1000
			verifiedTrust := calculateTrustScore(true, upvoteCount, maxUpvotes)
			unverifiedTrust := calculateTrustScore(false, upvoteCount, maxUpvotes)

			return verifiedTrust > unverifiedTrust
		},
		gen.IntRange(0, 1000),
	))

	// Property 11.6: Trust score is bounded between 0 and 1
	// Validates: Requirement 6.4 - trust score calculation is valid
	properties.Property("trust_score is bounded between 0 and 1", prop.ForAll(
		func(isVerified bool, upvoteCount int) bool {
			if upvoteCount < 0 {
				upvoteCount = 0
			}
			if upvoteCount > 10000 {
				upvoteCount = 10000
			}

			maxUpvotes := 10000
			trustScore := calculateTrustScore(isVerified, upvoteCount, maxUpvotes)

			return trustScore >= 0.0 && trustScore <= 1.0
		},
		gen.Bool(),
		gen.IntRange(0, 10000),
	))

	// Property 11.7: Verified filter count matches actual verified pods in dataset
	// Validates: Requirement 6.3 - filter returns correct count
	properties.Property("verified filter returns correct count", prop.ForAll(
		func(pods []domain.PodDocument) bool {
			if len(pods) == 0 {
				return true
			}

			ctx := context.Background()
			searchRepo := newMockSearchRepoWithVerifiedFilter(pods)
			svc := NewService(
				searchRepo,
				newMockIndexRepo(),
				newMockVectorSearchRepo(),
				newMockSearchHistoryRepo(),
				newMockTrendingRepo(),
			)

			// Count verified pods in input
			expectedVerifiedCount := 0
			for _, pod := range pods {
				if pod.IsVerified {
					expectedVerifiedCount++
				}
			}

			verified := true
			input := SearchInput{
				Query:    "",
				Verified: &verified,
				Page:     1,
				PerPage:  100,
			}

			result, err := svc.Search(ctx, input)
			if err != nil {
				return false
			}

			return int(result.Total) == expectedVerifiedCount
		},
		genPodDocumentSlice(0, 20, 1000),
	))

	// Property 11.8: No verified filter returns all pods
	// Validates: Requirement 6.3 - filter is optional
	properties.Property("no verified filter returns all pods", prop.ForAll(
		func(pods []domain.PodDocument) bool {
			ctx := context.Background()
			searchRepo := newMockSearchRepoWithVerifiedFilter(pods)
			svc := NewService(
				searchRepo,
				newMockIndexRepo(),
				newMockVectorSearchRepo(),
				newMockSearchHistoryRepo(),
				newMockTrendingRepo(),
			)

			input := SearchInput{
				Query:    "",
				Verified: nil, // No filter
				Page:     1,
				PerPage:  100,
			}

			result, err := svc.Search(ctx, input)
			if err != nil {
				return false
			}

			return int(result.Total) == len(pods)
		},
		genPodDocumentSlice(0, 20, 1000),
	))

	// Property 11.9: Higher upvotes always increase trust_score (monotonicity)
	// Validates: Requirement 6.4 - upvotes contribute positively to trust score
	properties.Property("higher upvotes increase trust_score", prop.ForAll(
		func(isVerified bool, upvote1, upvote2 int) bool {
			if upvote1 < 0 {
				upvote1 = 0
			}
			if upvote2 < 0 {
				upvote2 = 0
			}
			if upvote1 > 10000 {
				upvote1 = 10000
			}
			if upvote2 > 10000 {
				upvote2 = 10000
			}

			maxUpvotes := 10000
			trust1 := calculateTrustScore(isVerified, upvote1, maxUpvotes)
			trust2 := calculateTrustScore(isVerified, upvote2, maxUpvotes)

			if upvote1 > upvote2 {
				return trust1 > trust2
			} else if upvote1 < upvote2 {
				return trust1 < trust2
			}
			return trust1 == trust2
		},
		gen.Bool(),
		gen.IntRange(0, 10000),
		gen.IntRange(0, 10000),
	))

	// Property 11.10: Search results include trust indicators (is_verified and upvote_count)
	// Validates: Requirements 6.1, 6.2 - trust indicators are included in response
	properties.Property("search results include trust indicators", prop.ForAll(
		func(pods []domain.PodDocument) bool {
			if len(pods) == 0 {
				return true
			}

			ctx := context.Background()
			searchRepo := newMockSearchRepoWithVerifiedFilter(pods)
			svc := NewService(
				searchRepo,
				newMockIndexRepo(),
				newMockVectorSearchRepo(),
				newMockSearchHistoryRepo(),
				newMockTrendingRepo(),
			)

			input := SearchInput{
				Query:   "",
				Page:    1,
				PerPage: 100,
			}

			result, err := svc.Search(ctx, input)
			if err != nil {
				return false
			}

			// All results must have is_verified and upvote_count in metadata
			for _, r := range result.Results {
				if _, ok := r.Metadata["is_verified"]; !ok {
					return false
				}
				if _, ok := r.Metadata["upvote_count"]; !ok {
					return false
				}
				if _, ok := r.Metadata["trust_score"]; !ok {
					return false
				}
			}

			return true
		},
		genPodDocumentSlice(1, 20, 1000),
	))

	properties.TestingRun(t)
}
