package testextras

import (
	"os"
	"time"

	"gorm.io/gorm"
)

// Test Logging structure
type TestLog struct {
	ID        uint      `gorm:"primaryKey;"`
	LogTime   time.Time `sql:"index:idx_testlog_log_time"`
	ProcessId int
	Message   string `gorm:"size:200;default:''"`
}

func LogMessage(db *gorm.DB, message string) {
	pid := os.Getpid()
	record := TestLog{LogTime: time.Now().UTC(), ProcessId: pid, Message: message}
	db.Create(&record)
}
