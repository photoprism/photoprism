package sample

import (
	"testing"

	"github.com/jinzhu/gorm"
)

// TestSites holds one site in a test file, which is loaded in more than one package variant.
func TestSites(t *testing.T) {
	var db *gorm.DB

	if db != nil && db.RecordNotFound() {
		t.Fail()
	}
}
