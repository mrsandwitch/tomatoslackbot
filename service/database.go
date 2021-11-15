package service

import (
	"bushyang/tomatoslackbot/util"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type DbService struct {
	Db *gorm.DB
}

type ClockRecord struct {
	Id       int
	Start    time.Time
	Duration time.Duration
	Tag      string
	Desc     string
}

func InitDbService() (*DbService, error) {
	os.Mkdir(util.GetDataDir(), 0666)
	dbPath := filepath.Join(util.GetDataDir(), "storage.db")
	db, err := gorm.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Failed to open database. %v", err)
		return nil, err
	}
	db.SingularTable(true)

	db.AutoMigrate(&ClockRecord{})

	service := &DbService{}
	service.Db = db

	return service, nil
}

func (service *DbService) ClockRecordAdd(record ClockRecord) error {
	if result := service.Db.Create(&record); result.Error != nil {
		log.Println(result.Error)
		return result.Error
	}

	return nil
}

func (service *DbService) ClockRecordGet() ([]ClockRecord, error) {
	var records []ClockRecord

	if result := service.Db.Find(&records); result.Error != nil {
		log.Println(result.Error)
		return nil, result.Error
	}

	return records, nil
}

func (service *DbService) Close() {
	service.Db.Close()
}
