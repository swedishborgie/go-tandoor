package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "conversions.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadUnitConversionsFromFileBareArray(t *testing.T) {
	path := writeTempFile(t, `[
		{"food": 402, "base_unit": 27, "base_amount": 1, "converted_unit": 17, "converted_amount": 105},
		{"food": 603, "base_unit": 148, "base_amount": 1, "converted_unit": 17, "converted_amount": 110}
	]`)
	list, err := loadUnitConversionsFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 conversions, got %d", len(list))
	}
	c := list[0]
	if c.Food == nil || c.Food.ID != 402 || c.BaseUnit == nil || c.BaseUnit.ID != 27 ||
		c.BaseAmount != 1 || c.ConvertedUnit == nil || c.ConvertedUnit.ID != 17 || c.ConvertedAmount != 105 {
		t.Fatalf("unexpected conversion: %+v", c)
	}
}

func TestLoadUnitConversionsFromFileWrapped(t *testing.T) {
	path := writeTempFile(t, `{"conversions": [
		{"food": 1504, "base_unit": 20, "base_amount": 1, "converted_unit": 17, "converted_amount": 453.6}
	]}`)
	list, err := loadUnitConversionsFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 conversion, got %d", len(list))
	}
	c := list[0]
	if c.Food == nil || c.Food.ID != 1504 || c.BaseUnit == nil || c.BaseUnit.ID != 20 ||
		c.ConvertedUnit == nil || c.ConvertedUnit.ID != 17 || c.ConvertedAmount != 453.6 {
		t.Fatalf("unexpected conversion: %+v", c)
	}
}

func TestLoadUnitConversionsFromFileObjectRefs(t *testing.T) {
	path := writeTempFile(t, `{"conversions": [
		{"base_unit": {"id": 27}, "base_amount": 1, "converted_unit": {"id": 17}, "converted_amount": 140, "food": {"id": 603}}
	]}`)
	list, err := loadUnitConversionsFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 conversion, got %d", len(list))
	}
	c := list[0]
	if c.Food == nil || c.Food.ID != 603 || c.BaseUnit == nil || c.BaseUnit.ID != 27 ||
		c.ConvertedUnit == nil || c.ConvertedUnit.ID != 17 || c.ConvertedAmount != 140 {
		t.Fatalf("unexpected conversion: %+v", c)
	}
}

func TestLoadUnitConversionsFromFileInvalid(t *testing.T) {
	if _, err := loadUnitConversionsFromFile(writeTempFile(t, `{"foo": 1}`)); err == nil {
		t.Fatal("expected error for unrecognized shape")
	}
	if _, err := loadUnitConversionsFromFile(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected error for missing file")
	}
}
