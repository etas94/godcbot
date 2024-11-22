package database

import (
	"encoding/json"
	"os"
	"sync"
)

var dbLock sync.Mutex

// 用來保存圖片連結
type ImageDB struct {
	Images map[string]string `json:"images"`
}

func LoadDatabase(filePath string) (*ImageDB, error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 預防萬一，如果文件不存在，創建一個空的數據庫
			return &ImageDB{Images: make(map[string]string)}, nil
		}
		return nil, err
	}
	defer file.Close()

	var db ImageDB
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&db); err != nil {
		return nil, err
	}

	return &db, nil
}

func SaveDatabase(filePath string, db *ImageDB) error {
	dbLock.Lock()
	defer dbLock.Unlock()

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(db); err != nil {
		return err
	}

	return nil
}
