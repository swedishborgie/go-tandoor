package fdc

import (
	"encoding/json"
	"testing"
)

// TestFoodJSONRoundTrip verifies that the Food struct marshals and
// unmarshals correctly with the expected JSON field names.
func TestFoodJSONRoundTrip(t *testing.T) {
	jsonData := `{
		"fdcId": 534358,
		"dataType": "Branded",
		"description": "NUT 'N BERRY MIX",
		"brandOwner": "Kar Nut Products Company",
		"gtinUpc": "077034085228",
		"servingSize": 28,
		"servingSizeUnit": "g",
		"foodNutrients": [
			{
				"id": 167514,
				"amount": 90.3,
				"nutrient": {
					"id": 1005,
					"number": "305",
					"name": "Carbohydrate, by difference",
					"unitName": "g"
				}
			}
		],
		"foodPortions": [
			{
				"id": 135806,
				"amount": 1,
				"gramWeight": 91,
				"portionDescription": "1 cup",
				"measureUnit": {
					"id": 999,
					"name": "undetermined"
				}
			}
		]
	}`

	var food Food
	if err := json.Unmarshal([]byte(jsonData), &food); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if food.FDCID != 534358 {
		t.Errorf("FDCID: got %d, want 534358", food.FDCID)
	}
	if food.DataType != "Branded" {
		t.Errorf("DataType: got %s, want Branded", food.DataType)
	}
	if food.BrandOwner != "Kar Nut Products Company" {
		t.Errorf("BrandOwner: got %s, want Kar Nut Products Company", food.BrandOwner)
	}
	if food.GTINUPC != "077034085228" {
		t.Errorf("GTINUPC: got %s, want 077034085228", food.GTINUPC)
	}
	if food.ServingSize == nil || *food.ServingSize != 28 {
		t.Errorf("ServingSize: got %v, want 28", food.ServingSize)
	}
	if len(food.FoodNutrients) != 1 {
		t.Fatalf("expected 1 food nutrient, got %d", len(food.FoodNutrients))
	}
	nutrient := food.FoodNutrients[0]
	if nutrient.ID != 167514 {
		t.Errorf("nutrient.ID: got %d, want 167514", nutrient.ID)
	}
	if nutrient.Nutrient == nil || nutrient.Nutrient.Name != "Carbohydrate, by difference" {
		t.Errorf("unexpected nutrient name: %v", nutrient.Nutrient)
	}
	if len(food.FoodPortions) != 1 {
		t.Fatalf("expected 1 food portion, got %d", len(food.FoodPortions))
	}
	portion := food.FoodPortions[0]
	if portion.PortionDescription != "1 cup" {
		t.Errorf("portionDescription: got %s, want 1 cup", portion.PortionDescription)
	}

	// Marshal back and verify it doesn't error
	out, err := json.Marshal(food)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var reFood Food
	if err := json.Unmarshal(out, &reFood); err != nil {
		t.Fatalf("re-Unmarshal: %v", err)
	}
	if reFood.FDCID != food.FDCID {
		t.Errorf("round-trip FDCID: got %d, want %d", reFood.FDCID, food.FDCID)
	}
}

// TestAbridgedFoodJSON verifies AbridgedFood unmarshals correctly.
func TestAbridgedFoodJSON(t *testing.T) {
	jsonData := `{
		"fdcId": 747448,
		"dataType": "Foundation",
		"description": "Strawberries, raw",
		"ndbNumber": "9316",
		"publicationDate": "12/16/2019"
	}`

	var food AbridgedFood
	if err := json.Unmarshal([]byte(jsonData), &food); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if food.FDCID != 747448 {
		t.Errorf("FDCID: got %d, want 747448", food.FDCID)
	}
	if food.DataType != "Foundation" {
		t.Errorf("DataType: got %s, want Foundation", food.DataType)
	}
	if !food.NDBNumber.HasValue || food.NDBNumber.Value != 9316 {
		t.Errorf("NDBNumber: got %v, want 9316", food.NDBNumber)
	}
}

// TestSearchResponseJSON verifies SearchResponse unmarshals correctly.
func TestSearchResponseJSON(t *testing.T) {
	jsonData := `{
		"totalHits": 1034,
		"currentPage": 1,
		"totalPages": 21,
		"foods": [
			{
				"fdcId": 45001529,
				"description": "BROCCOLI",
				"dataType": "Branded",
				"score": 12.5
			}
		]
	}`

	var resp SearchResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if resp.TotalHits != 1034 {
		t.Errorf("TotalHits: got %d, want 1034", resp.TotalHits)
	}
	if resp.TotalPages != 21 {
		t.Errorf("TotalPages: got %d, want 21", resp.TotalPages)
	}
	if len(resp.Foods) != 1 {
		t.Fatalf("expected 1 food, got %d", len(resp.Foods))
	}
	food := resp.Foods[0]
	if food.FDCID != 45001529 {
		t.Errorf("food.FDCID: got %d, want 45001529", food.FDCID)
	}
	if food.Description != "BROCCOLI" {
		t.Errorf("food.Description: got %s, want BROCCOLI", food.Description)
	}
	if food.Score == nil || *food.Score != 12.5 {
		t.Errorf("food.Score: got %v, want 12.5", food.Score)
	}
}

// TestLabelNutrientsJSON verifies label nutrients unmarshal correctly.
func TestLabelNutrientsJSON(t *testing.T) {
	jsonData := `{
		"calories": {"value": 140},
		"fat": {"value": 8.9992},
		"protein": {"value": 4.0012},
		"sodium": {"value": 0},
		"carbohydrates": {"value": 12.0008}
	}`

	var ln LabelNutrients
	if err := json.Unmarshal([]byte(jsonData), &ln); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if ln.Calories == nil || *ln.Calories.Value != 140 {
		t.Errorf("Calories: got %v, want 140", ln.Calories)
	}
	if ln.Fat == nil || *ln.Fat.Value != 8.9992 {
		t.Errorf("Fat: got %v, want 8.9992", ln.Fat)
	}
}

// TestFoodsCriteriaJSON verifies request body marshaling.
func TestFoodsCriteriaJSON(t *testing.T) {
	format := FormatFull
	criteria := FoodsCriteria{
		FDCIDs:    []int{534358, 373052},
		Format:    &format,
		Nutrients: []int{203, 204},
	}

	data, err := json.Marshal(criteria)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if result["format"] != "full" {
		t.Errorf("format: got %v, want full", result["format"])
	}
}

// TestFoodSearchCriteriaJSON verifies search request marshaling.
func TestFoodSearchCriteriaJSON(t *testing.T) {
	sortBy := SortFieldDescription
	sortOrder := SortOrderAsc
	pageSize := 25
	brandOwner := "Test Brand"
	criteria := FoodSearchCriteria{
		Query:      "cheddar cheese",
		DataType:   []DataType{DataTypeFoundation},
		PageSize:   &pageSize,
		SortBy:     &sortBy,
		SortOrder:  &sortOrder,
		BrandOwner: &brandOwner,
	}

	data, err := json.Marshal(criteria)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if result["query"] != "cheddar cheese" {
		t.Errorf("query: got %v, want cheddar cheese", result["query"])
	}
	if result["pageSize"].(float64) != 25 {
		t.Errorf("pageSize: got %v, want 25", result["pageSize"])
	}
}

// TestFoodListCriteriaJSON verifies list criteria marshaling.
func TestFoodListCriteriaJSON(t *testing.T) {
	sortBy := SortFieldFdcID
	sortOrder := SortOrderDesc
	pageSize := 100
	pageNumber := 2
	criteria := FoodListCriteria{
		DataType:   []DataType{DataTypeBranded, DataTypeFoundation},
		PageSize:   &pageSize,
		PageNumber: &pageNumber,
		SortBy:     &sortBy,
		SortOrder:  &sortOrder,
	}

	data, err := json.Marshal(criteria)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if result["pageNumber"].(float64) != 2 {
		t.Errorf("pageNumber: got %v, want 2", result["pageNumber"])
	}
}

// TestFlexibleStringJSON verifies FlexibleString handles string, number, and null.
func TestFlexibleStringJSON(t *testing.T) {
	// Test unmarshal from string
	var s1 FlexibleString
	if err := json.Unmarshal([]byte(`"whole"`), &s1); err != nil {
		t.Fatalf("Unmarshal string: %v", err)
	}
	if !s1.HasValue {
		t.Error("HasValue should be true for string input")
	}
	if s1.Value != "whole" {
		t.Errorf("Value: got %q, want \"whole\"", s1.Value)
	}
	if s1.String() != "whole" {
		t.Errorf("String(): got %q, want \"whole\"", s1.String())
	}

	// Test unmarshal from number
	var s2 FlexibleString
	if err := json.Unmarshal([]byte(`1`), &s2); err != nil {
		t.Fatalf("Unmarshal number: %v", err)
	}
	if !s2.HasValue {
		t.Error("HasValue should be true for number input")
	}
	if s2.Value != "1" {
		t.Errorf("Value: got %q, want \"1\"", s2.Value)
	}

	// Test unmarshal from null
	var s3 FlexibleString
	if err := json.Unmarshal([]byte(`null`), &s3); err != nil {
		t.Fatalf("Unmarshal null: %v", err)
	}
	if s3.HasValue {
		t.Error("HasValue should be false for null input")
	}
	if s3.Value != "" {
		t.Errorf("Value: got %q, want \"\"", s3.Value)
	}
	if s3.String() != "" {
		t.Errorf("String(): got %q, want \"\"", s3.String())
	}

	// Test marshal string
	data, err := json.Marshal(s1)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(data) != `"whole"` {
		t.Errorf("Marshal: got %s, want \"whole\"", data)
	}

	// Test marshal null
	data, err = json.Marshal(s3)
	if err != nil {
		t.Fatalf("Marshal null: %v", err)
	}
	if string(data) != "null" {
		t.Errorf("Marshal null: got %s, want null", data)
	}

	// Test round-trip
	var s4 FlexibleString
	if err := json.Unmarshal(data, &s4); err != nil {
		t.Fatalf("Re-unmarshal: %v", err)
	}
	if s4.HasValue || s4.Value != "" {
		t.Errorf("round-trip null: got %+v, want zero value", s4)
	}
}

func TestFlexibleIntString(t *testing.T) {
	if (FlexibleInt{Value: 42, HasValue: true}).String() != "42" {
		t.Errorf("String(): got %q, want \"42\"", (FlexibleInt{Value: 42, HasValue: true}).String())
	}
	if (FlexibleInt{}).String() != "" {
		t.Errorf("String(): got %q, want \"\"", (FlexibleInt{}).String())
	}
}
