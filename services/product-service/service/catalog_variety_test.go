package service

import (
	"strings"
	"testing"
)

// TestCatalogVarietyAndUniqueness verifies that the catalog contains distinct products,
// unique IDs, distinct categories, and appropriate distribution across all categories.
func TestCatalogVarietyAndUniqueness(t *testing.T) {
	svc := setupTestProductService()

	products, err := svc.ListProducts()
	if err != nil {
		t.Fatalf("Failed to retrieve products: %v", err)
	}

	if len(products) != 26 {
		t.Errorf("Expected exactly 26 distinct products, got %d", len(products))
	}

	seenIDs := make(map[int]bool)
	seenNames := make(map[string]bool)
	categoryCounts := make(map[string]int)

	for _, p := range products {
		// 1. Verify unique IDs
		if seenIDs[p.ID] {
			t.Errorf("Duplicate product ID found: %d (%s)", p.ID, p.Name)
		}
		seenIDs[p.ID] = true

		// 2. Verify unique names
		normName := strings.ToLower(strings.TrimSpace(p.Name))
		if seenNames[normName] {
			t.Errorf("Duplicate product name found: %s", p.Name)
		}
		seenNames[normName] = true

		// 3. Verify non-empty and non-zero fields
		if p.Name == "" {
			t.Errorf("Product ID %d has empty name", p.ID)
		}
		if p.Description == "" {
			t.Errorf("Product ID %d has empty description", p.ID)
		}
		if p.Price <= 0 {
			t.Errorf("Product ID %d has invalid price: %f", p.ID, p.Price)
		}
		if p.Stock <= 0 {
			t.Errorf("Product ID %d has invalid stock: %d", p.ID, p.Stock)
		}
		if p.Category == "" {
			t.Errorf("Product ID %d has empty category", p.ID)
		}

		categoryCounts[p.Category]++
	}

	// 4. Verify all 6 supported categories are represented
	expectedCategories := []string{"Electronics", "Audio", "Accessories", "Wearables", "Photography", "Gaming"}
	for _, cat := range expectedCategories {
		count := categoryCounts[cat]
		if count == 0 {
			t.Errorf("Category '%s' has 0 products in catalog", cat)
		} else if count < 3 {
			t.Errorf("Category '%s' has only %d products, expected at least 3 for meaningful variety", cat, count)
		}
	}
}

// TestCategoryFiltering verifies that category filtering behaves accurately.
func TestCategoryFiltering(t *testing.T) {
	svc := setupTestProductService()

	products, err := svc.ListProducts()
	if err != nil {
		t.Fatalf("Failed to retrieve products: %v", err)
	}

	testCases := []struct {
		category string
		minCount int
	}{
		{"Electronics", 6},
		{"Audio", 4},
		{"Accessories", 5},
		{"Wearables", 3},
		{"Photography", 4},
		{"Gaming", 4},
	}

	for _, tc := range testCases {
		t.Run(tc.category, func(t *testing.T) {
			count := 0
			for _, p := range products {
				if strings.EqualFold(p.Category, tc.category) {
					count++
				}
			}
			if count < tc.minCount {
				t.Errorf("Category %s: expected at least %d products, got %d", tc.category, tc.minCount, count)
			}
		})
	}
}
