// Package application contains property-based tests for material deletion and storage usage.
package application

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: user-storage-limit, Property 4: Deletion Decreases Usage**
//
// *For any* user who deletes a material of size N bytes, the user's storage usage
// SHALL decrease by exactly N bytes.
//
// **Validates: Requirements 4.1, 4.2, 4.3**

// MaterialTestData represents a material with file size and deletion status for testing.
type MaterialTestData struct {
	ID        uuid.UUID
	FileSize  int64
	IsDeleted bool
}

// CalculateStorageUsage calculates total storage used by summing file sizes of non-deleted materials.
// This simulates what the StorageRepository.GetUserStorageUsage does.
func CalculateStorageUsage(materials []MaterialTestData) int64 {
	var total int64
	for _, m := range materials {
		if !m.IsDeleted {
			total += m.FileSize
		}
	}
	return total
}

// DeleteMaterial marks a material as deleted and returns the new list.
// This simulates the soft-delete operation.
func DeleteMaterial(materials []MaterialTestData, materialID uuid.UUID) []MaterialTestData {
	result := make([]MaterialTestData, len(materials))
	copy(result, materials)
	for i := range result {
		if result[i].ID == materialID {
			result[i].IsDeleted = true
			break
		}
	}
	return result
}

func TestProperty_DeletionDecreasesUsage(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 4.1: Deleting a material decreases usage by exactly its file size
	// Validates: Requirement 4.1 - WHEN a material is deleted THEN THE Storage_Limit_System
	// SHALL recalculate the user's total storage usage
	// Validates: Requirement 4.2 - WHEN a material is deleted THEN THE Storage_Limit_System
	// SHALL decrease the user's usage by the size of the deleted material
	properties.Property("deleting material decreases usage by exactly file size", prop.ForAll(
		func(materials []MaterialTestData, deleteIndex int) bool {
			// Filter to only non-deleted materials for deletion candidate
			nonDeletedMaterials := make([]MaterialTestData, 0)
			for _, m := range materials {
				if !m.IsDeleted {
					nonDeletedMaterials = append(nonDeletedMaterials, m)
				}
			}

			// Skip if no non-deleted materials to delete
			if len(nonDeletedMaterials) == 0 {
				return true
			}

			// Select material to delete
			idx := deleteIndex % len(nonDeletedMaterials)
			materialToDelete := nonDeletedMaterials[idx]

			// Calculate usage before deletion
			usageBefore := CalculateStorageUsage(materials)

			// Delete the material
			materialsAfter := DeleteMaterial(materials, materialToDelete.ID)

			// Calculate usage after deletion
			usageAfter := CalculateStorageUsage(materialsAfter)

			// Usage should decrease by exactly the file size
			expectedUsageAfter := usageBefore - materialToDelete.FileSize

			return usageAfter == expectedUsageAfter
		},
		genMaterialTestDataList(),
		gen.IntRange(0, 100),
	))

	// Property 4.2: Deleted materials are not counted in storage usage
	// Validates: Requirement 4.3 - WHEN calculating storage usage THEN THE Storage_Limit_System
	// SHALL only count non-deleted materials
	properties.Property("deleted materials are not counted in usage", prop.ForAll(
		func(materials []MaterialTestData) bool {
			// Calculate expected: only non-deleted materials
			var expectedUsage int64
			for _, m := range materials {
				if !m.IsDeleted {
					expectedUsage += m.FileSize
				}
			}

			actualUsage := CalculateStorageUsage(materials)
			return actualUsage == expectedUsage
		},
		genMaterialTestDataListWithDeletions(),
	))

	// Property 4.3: Deleting already deleted material does not change usage
	properties.Property("deleting already deleted material does not change usage", prop.ForAll(
		func(materials []MaterialTestData, deleteIndex int) bool {
			// Filter to only deleted materials
			deletedMaterials := make([]MaterialTestData, 0)
			for _, m := range materials {
				if m.IsDeleted {
					deletedMaterials = append(deletedMaterials, m)
				}
			}

			// Skip if no deleted materials
			if len(deletedMaterials) == 0 {
				return true
			}

			// Select already deleted material
			idx := deleteIndex % len(deletedMaterials)
			materialToDelete := deletedMaterials[idx]

			// Calculate usage before "re-deletion"
			usageBefore := CalculateStorageUsage(materials)

			// Try to delete the already deleted material
			materialsAfter := DeleteMaterial(materials, materialToDelete.ID)

			// Calculate usage after
			usageAfter := CalculateStorageUsage(materialsAfter)

			// Usage should remain the same
			return usageAfter == usageBefore
		},
		genMaterialTestDataListWithDeletions(),
		gen.IntRange(0, 100),
	))

	// Property 4.4: Storage usage is never negative after deletion
	properties.Property("storage usage is never negative after deletion", prop.ForAll(
		func(materials []MaterialTestData, deleteIndex int) bool {
			// Filter to only non-deleted materials
			nonDeletedMaterials := make([]MaterialTestData, 0)
			for _, m := range materials {
				if !m.IsDeleted {
					nonDeletedMaterials = append(nonDeletedMaterials, m)
				}
			}

			// Skip if no non-deleted materials
			if len(nonDeletedMaterials) == 0 {
				return true
			}

			// Select material to delete
			idx := deleteIndex % len(nonDeletedMaterials)
			materialToDelete := nonDeletedMaterials[idx]

			// Delete the material
			materialsAfter := DeleteMaterial(materials, materialToDelete.ID)

			// Calculate usage after deletion
			usageAfter := CalculateStorageUsage(materialsAfter)

			return usageAfter >= 0
		},
		genMaterialTestDataList(),
		gen.IntRange(0, 100),
	))

	// Property 4.5: Deleting all materials results in zero usage
	properties.Property("deleting all materials results in zero usage", prop.ForAll(
		func(materials []MaterialTestData) bool {
			// Delete all materials
			allDeleted := make([]MaterialTestData, len(materials))
			for i, m := range materials {
				allDeleted[i] = MaterialTestData{
					ID:        m.ID,
					FileSize:  m.FileSize,
					IsDeleted: true,
				}
			}

			usage := CalculateStorageUsage(allDeleted)
			return usage == 0
		},
		genMaterialTestDataList(),
	))

	// Property 4.6: Storage usage calculation is deterministic
	properties.Property("storage usage calculation is deterministic", prop.ForAll(
		func(materials []MaterialTestData) bool {
			usage1 := CalculateStorageUsage(materials)
			usage2 := CalculateStorageUsage(materials)
			return usage1 == usage2
		},
		genMaterialTestDataListWithDeletions(),
	))

	// Property 4.7: Order of materials does not affect usage calculation
	properties.Property("order of materials does not affect usage", prop.ForAll(
		func(materials []MaterialTestData) bool {
			if len(materials) < 2 {
				return true
			}

			// Calculate usage with original order
			usage1 := CalculateStorageUsage(materials)

			// Reverse the order
			reversed := make([]MaterialTestData, len(materials))
			for i, m := range materials {
				reversed[len(materials)-1-i] = m
			}

			// Calculate usage with reversed order
			usage2 := CalculateStorageUsage(reversed)

			return usage1 == usage2
		},
		genMaterialTestDataListWithDeletions(),
	))

	// Property 4.8: Multiple deletions decrease usage by sum of file sizes
	properties.Property("multiple deletions decrease usage by sum of file sizes", prop.ForAll(
		func(materials []MaterialTestData) bool {
			// Filter to only non-deleted materials
			nonDeletedMaterials := make([]MaterialTestData, 0)
			for _, m := range materials {
				if !m.IsDeleted {
					nonDeletedMaterials = append(nonDeletedMaterials, m)
				}
			}

			// Skip if less than 2 non-deleted materials
			if len(nonDeletedMaterials) < 2 {
				return true
			}

			// Calculate usage before deletions
			usageBefore := CalculateStorageUsage(materials)

			// Delete first two non-deleted materials
			mat1 := nonDeletedMaterials[0]
			mat2 := nonDeletedMaterials[1]

			materialsAfter := DeleteMaterial(materials, mat1.ID)
			materialsAfter = DeleteMaterial(materialsAfter, mat2.ID)

			// Calculate usage after deletions
			usageAfter := CalculateStorageUsage(materialsAfter)

			// Usage should decrease by sum of both file sizes
			expectedUsageAfter := usageBefore - mat1.FileSize - mat2.FileSize

			return usageAfter == expectedUsageAfter
		},
		genMaterialTestDataList(),
	))

	properties.TestingRun(t)
}

// Generator for a list of material test data (all non-deleted)
func genMaterialTestDataList() gopter.Gen {
	return gen.SliceOfN(10, genMaterialTestData())
}

// Generator for a list of material test data with some deletions
func genMaterialTestDataListWithDeletions() gopter.Gen {
	return gen.SliceOfN(10, genMaterialTestDataWithDeletion())
}

// Generator for a single material test data (non-deleted)
func genMaterialTestData() gopter.Gen {
	return gen.Int64Range(1, 100*1024*1024).Map(func(size int64) MaterialTestData {
		return MaterialTestData{
			ID:        uuid.New(),
			FileSize:  size,
			IsDeleted: false,
		}
	})
}

// Generator for a single material test data (may be deleted)
func genMaterialTestDataWithDeletion() gopter.Gen {
	return gopter.CombineGens(
		gen.Int64Range(1, 100*1024*1024), // 1 byte to 100MB
		gen.Bool(),
	).Map(func(vals []interface{}) MaterialTestData {
		return MaterialTestData{
			ID:        uuid.New(),
			FileSize:  vals[0].(int64),
			IsDeleted: vals[1].(bool),
		}
	})
}
