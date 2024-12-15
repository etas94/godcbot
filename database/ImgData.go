package database

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

// 使用互斥鎖預防對數據庫的同時訪問
var dbLock sync.Mutex

// ImageDB 結構保存所有圖片和分類數據
// Images 是圖片資料映射
// Categories 是分類名稱與對應編號的映射
type ImageDB struct {
	Images     map[string]ImageData `json:"images"`
	Categories map[string]string    `json:"categories"` // 儲存分類名稱與對應的編號
}

// ImageData 結構保存單張圖片的詳細資訊
type ImageData struct {
	URL      string `json:"url"`      // 圖片網址
	Name     string `json:"name"`     // 圖片名稱
	ID       string `json:"id"`       // 圖片ID
	Category string `json:"category"` // 圖片分類的編號
}

// LoadDatabase 從文件中加載數據庫
// 返回 *ImageDB 和錯誤（如果發生）
func LoadDatabase(filePath string) (*ImageDB, error) {
	dbLock.Lock()
	defer dbLock.Unlock() //在當前函數執行完後自動解鎖

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 如果文件不存在，創建一個空的數據庫
			return &ImageDB{Images: make(map[string]ImageData), Categories: make(map[string]string)}, nil
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

// SaveDatabase 將數據庫保存到文件中
// 返回錯誤（如果發生）
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

// SearchImageByName 根據部分名稱搜尋圖片
// 返回第一個匹配的圖片ID（string），如果沒有匹配則返回空字串和nil
func SearchImageByName(db *ImageDB, searchString string) (string, error) {
	var matchedID string

	// 遍歷所有圖片資料
	for _, image := range db.Images {
		// 檢查圖片名稱是否包含搜尋字串（大小寫不敏感）
		if strings.Contains(strings.ToLower(image.Name), strings.ToLower(searchString)) {
			matchedID = image.ID
			break // 找到一個就停止
		}
	}

	// 如果沒有找到匹配的圖片，返回空字串
	if matchedID == "" {
		return "", nil
	}
	return matchedID, nil
}
