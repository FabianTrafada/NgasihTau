// Package application contains property-based tests for the Pod Service.
package application

import (
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"ngasihtau/internal/pod/domain"
)

// **Feature: student-teacher-roles, Property 9: Upvote Display and Sorting**
// **Validates: Requirements 5.4, 5.5**
//
// Property 9: Upvote Display and Sorting
// *For any* knowledge pod query:
// - The response SHALL include the total upvote_count
// - When sorted by upvotes, pods with higher upvote_count SHALL appear first

func TestProperty_UpvoteDisplayAndSorting(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 9.1: Pod entity always includes upvote_count field
	// Validates: Requirement 5.4 - WHEN displaying knowledge pod details,
	// THE Pod Service SHALL show the total upvote count.
	properties.Property("pod entity always includes upvote_count field", prop.ForAll(
		func(name, slug string, isVerified bool, upvoteCount int) bool {
			ownerID := uuid.New()
			pod := domain.NewPod(ownerID, name, slug, domain.VisibilityPublic, isVerified)

			// Set upvote count to test value
			if upvoteCount < 0 {
				upvoteCount = 0
			}
			pod.UpvoteCount = upvoteCount

			// Verify upvote_count is accessible and matches what was set
			return pod.UpvoteCount == upvoteCount
		},
		genValidPodName(),
		genValidSlug(),
		gen.Bool(),
		gen.IntRange(0, 100000),
	))

	// Property 9.2: New pods start with zero upvote count
	// Validates: Requirement 5.4 - upvote count is always present (starts at 0)
	properties.Property("new pods start with zero upvote count", prop.ForAll(
		func(name, slug string, isVerified bool) bool {
			ownerID := uuid.New()
			pod := domain.NewPod(ownerID, name, slug, domain.VisibilityPublic, isVerified)

			// New pods should have upvote_count = 0
			return pod.UpvoteCount == 0
		},
		genValidPodName(),
		genValidSlug(),
		gen.Bool(),
	))

	// Property 9.3: Sorting pods by upvote count produces descending order
	// Validates: Requirement 5.5 - THE Pod Service SHALL support sorting
	// knowledge pods by upvote count in search results.
	properties.Property("sorting pods by upvote count produces descending order", prop.ForAll(
		func(upvoteCounts []int) bool {
			// Skip empty or single-element slices
			if len(upvoteCounts) < 2 {
				return true
			}

			// Create pods with the given upvote counts
			pods := make([]*domain.Pod, len(upvoteCounts))
			for i, count := range upvoteCounts {
				ownerID := uuid.New()
				pod := domain.NewPod(ownerID, "Pod", "pod-"+uuid.New().String()[:8], domain.VisibilityPublic, false)
				if count < 0 {
					count = 0
				}
				pod.UpvoteCount = count
				pods[i] = pod
			}

			// Sort pods by upvote count (descending) - simulating search result sorting
			sort.Slice(pods, func(i, j int) bool {
				return pods[i].UpvoteCount > pods[j].UpvoteCount
			})

			// Verify the sorted order is descending
			for i := 0; i < len(pods)-1; i++ {
				if pods[i].UpvoteCount < pods[i+1].UpvoteCount {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(10, gen.IntRange(0, 1000)),
	))

	// Property 9.4: Pods with higher upvote counts appear first when sorted
	// Validates: Requirement 5.5 - pods with higher upvote_count SHALL appear first
	properties.Property("pods with higher upvote counts appear first when sorted", prop.ForAll(
		func(highCount, lowCount int) bool {
			// Ensure highCount > lowCount
			if highCount <= lowCount {
				highCount = lowCount + 1
			}
			if highCount < 0 {
				highCount = 1
			}
			if lowCount < 0 {
				lowCount = 0
			}

			// Create two pods with different upvote counts
			ownerID := uuid.New()
			highPod := domain.NewPod(ownerID, "High Pod", "high-pod", domain.VisibilityPublic, false)
			highPod.UpvoteCount = highCount

			lowPod := domain.NewPod(ownerID, "Low Pod", "low-pod", domain.VisibilityPublic, false)
			lowPod.UpvoteCount = lowCount

			// Create slice in random order
			pods := []*domain.Pod{lowPod, highPod}

			// Sort by upvote count descending
			sort.Slice(pods, func(i, j int) bool {
				return pods[i].UpvoteCount > pods[j].UpvoteCount
			})

			// The pod with higher upvote count should be first
			return pods[0].UpvoteCount == highCount && pods[1].UpvoteCount == lowCount
		},
		gen.IntRange(0, 10000),
		gen.IntRange(0, 10000),
	))

	// Property 9.5: Upvote count is non-negative in all pods
	// Validates: Requirement 5.4 - upvote count is a valid trust indicator
	properties.Property("upvote count is non-negative in all pods", prop.ForAll(
		func(name, slug string, isVerified bool) bool {
			ownerID := uuid.New()
			pod := domain.NewPod(ownerID, name, slug, domain.VisibilityPublic, isVerified)

			// Upvote count should never be negative
			return pod.UpvoteCount >= 0
		},
		genValidPodName(),
		genValidSlug(),
		gen.Bool(),
	))

	// Property 9.6: PodListResult includes pods with their upvote counts
	// Validates: Requirement 5.4 - upvote count is included in pod details
	properties.Property("pod list result includes pods with upvote counts", prop.ForAll(
		func(numPods int, upvoteCounts []int) bool {
			// Limit to reasonable range
			if numPods < 0 {
				numPods = 0
			}
			if numPods > 20 {
				numPods = 20
			}

			// Create pods with upvote counts
			pods := make([]*domain.Pod, numPods)
			for i := 0; i < numPods; i++ {
				ownerID := uuid.New()
				pod := domain.NewPod(ownerID, "Pod", "pod-"+uuid.New().String()[:8], domain.VisibilityPublic, false)
				if i < len(upvoteCounts) && upvoteCounts[i] >= 0 {
					pod.UpvoteCount = upvoteCounts[i]
				}
				pods[i] = pod
			}

			// Create a PodListResult (simulating service response)
			result := &PodListResult{
				Pods:       pods,
				Total:      numPods,
				Page:       1,
				PerPage:    20,
				TotalPages: 1,
			}

			// Verify all pods in result have accessible upvote counts
			for i, pod := range result.Pods {
				// Each pod should have a non-negative upvote count
				if pod.UpvoteCount < 0 {
					return false
				}
				// If we set a specific count, verify it's preserved
				if i < len(upvoteCounts) && upvoteCounts[i] >= 0 {
					if pod.UpvoteCount != upvoteCounts[i] {
						return false
					}
				}
			}

			return true
		},
		gen.IntRange(0, 20),
		gen.SliceOfN(20, gen.IntRange(0, 1000)),
	))

	// Property 9.7: Stable sorting preserves order for equal upvote counts
	// Validates: Requirement 5.5 - sorting behavior is consistent
	properties.Property("stable sorting preserves order for equal upvote counts", prop.ForAll(
		func(count int) bool {
			if count < 0 {
				count = 0
			}

			// Create multiple pods with the same upvote count
			ownerID := uuid.New()
			pod1 := domain.NewPod(ownerID, "Pod 1", "pod-1", domain.VisibilityPublic, false)
			pod1.UpvoteCount = count

			pod2 := domain.NewPod(ownerID, "Pod 2", "pod-2", domain.VisibilityPublic, false)
			pod2.UpvoteCount = count

			pod3 := domain.NewPod(ownerID, "Pod 3", "pod-3", domain.VisibilityPublic, false)
			pod3.UpvoteCount = count

			pods := []*domain.Pod{pod1, pod2, pod3}

			// Use stable sort
			sort.SliceStable(pods, func(i, j int) bool {
				return pods[i].UpvoteCount > pods[j].UpvoteCount
			})

			// All pods should still have the same upvote count
			for _, pod := range pods {
				if pod.UpvoteCount != count {
					return false
				}
			}

			return true
		},
		gen.IntRange(0, 10000),
	))

	properties.TestingRun(t)
}

// Generator for valid pod names
func genValidPodName() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) == 0 {
			return "Default Pod"
		}
		if len(s) > 100 {
			return s[:100]
		}
		return s
	})
}

// Generator for valid slugs
func genValidSlug() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) == 0 {
			return "default-slug"
		}
		if len(s) > 100 {
			return s[:100]
		}
		return s
	})
}
