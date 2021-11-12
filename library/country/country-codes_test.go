package country

import (
	"testing"
)

func TestFindByName(t *testing.T) {
	matches := FindByName("United States Minor")

	if len(matches) != 1 {
		t.Fatalf("Extra matches found")
	}

	um, _ := GetByAlpha2("UM")

	if matches[0] != um {
		t.Fatalf("Match for United States Minor Outlying Islands failed")
	}
}

func TestGetByNumeric(t *testing.T) {
	code, _ := GetByNumeric(840)

	if code.Name != "United States" {
		t.Fatalf("GetByNumeric failed")
	}
}
