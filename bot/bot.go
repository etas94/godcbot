package bot

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"
	"github.com/etas94/godcbot/config"
	"github.com/etas94/godcbot/database"
)

var goBot *discordgo.Session

const ImgDbFilePath = "./image.json"

// 初始化機器人並啟動
func Start() {
	// 讀取配置
	cfg, err := config.ReadConfig()
	if err != nil {
		fmt.Println("讀取配置失敗:", err)
		return
	}

	goBot, err = discordgo.New("Bot " + cfg.Token)
	if err != nil {
		fmt.Println("初始化Discord對話失敗:", err)
		return
	}

	// 註冊互動事件處理器來管理 Slash Command。
	goBot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handleCommand(s, i)
	})

	// 與Discord連接。
	err = goBot.Open()
	if err != nil {
		fmt.Println("與Discord連接失敗:", err)
		return
	}

	fmt.Println("機器人已成功連接！")

	// 註冊Slash Commands
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "回應 Pong!(測試用)",
		},
		{
			Name:        "image",
			Description: "根據名稱或ID獲取圖片",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "identifier",
					Description: "圖片的名稱或ID",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		{
			Name:        "addimage",
			Description: "添加圖片到圖庫",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "name",
					Description: "圖片的名稱",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "url",
					Description: "圖片的網址",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "category",
					Description: "圖片的分類(可選)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
				},
			},
		},
		{
			Name:        "delimage",
			Description: "從圖庫中刪除圖片",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "identifier",
					Description: "圖片的名稱或ID",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		{
			Name:        "send",
			Description: "機器人代為傳圖",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "identifier",
					Description: "圖片的名稱或ID",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		{
			Name:        "list",
			Description: "列出指定分類中的圖片",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "category",
					Description: "篩選的分類",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "page",
					Description: "要查看的頁數(可選，預設為1)",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
				},
			},
		},
		{
			Name:        "listall",
			Description: "列出所有分類中的圖片",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "page",
					Description: "要查看的頁數(可選，預設為1)",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
				},
			},
		},
		{
			Name:        "classify",
			Description: "更新圖片的分類",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "identifier",
					Description: "圖片的名稱或ID",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "category",
					Description: "新的分類名稱",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
	}

	// 註冊指令至全域和特定伺服器。
	for _, cmd := range commands {
		// 註冊到全域
		_, err := goBot.ApplicationCommandCreate(goBot.State.User.ID, "", cmd)
		if err != nil {
			fmt.Printf("無法創建全域指令 '%s': %v\n", cmd.Name, err)
		}

		// 註冊到特定伺服器，方便測試
		// _, err = goBot.ApplicationCommandCreate(goBot.State.User.ID, "1300011739290669109", cmd)
		// if err != nil {
		// 	fmt.Printf("無法創建伺服器指令 '%s' 在伺服器ID 1300011739290669109: %v\n", cmd.Name, err)
		// }
	}
}

// 處理Slash Command
func handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "ping":
		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Pong!",
			},
		}
		err := s.InteractionRespond(i.Interaction, response)
		if err != nil {
			fmt.Println("發送回應失敗:", err)
		}

	case "image":
		identifier := i.ApplicationCommandData().Options[0].StringValue()
		db, err := database.LoadDatabase(ImgDbFilePath)
		if err != nil {
			fmt.Println("讀取圖片庫失敗:", err)
			return
		}

		var imageData database.ImageData
		var found bool

		// 檢查輸入是否為 ID
		for _, img := range db.Images {
			if img.ID == identifier {
				imageData = img
				found = true
				break
			}
		}

		if !found {
			// 使用 SearchImageByName 函數進行搜尋
			matchedID, err := database.SearchImageByName(db, identifier)
			if err != nil || matchedID == "" {
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("找不到圖片 %q", identifier),
						Flags:   discordgo.MessageFlagsEphemeral, // 僅使用者可見。
					},
				}
				err = s.InteractionRespond(i.Interaction, response)
				if err != nil {
					fmt.Println("發送回應失敗:", err)
				}
				return
			}

			// 從 db.Images 中尋找對應的圖片資料
			for _, img := range db.Images {
				if img.ID == matchedID {
					imageData = img
					found = true
					break
				}
			}
		}

		if !found { //找不到此ID
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("找不到圖片 %q", identifier),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}
			err = s.InteractionRespond(i.Interaction, response)
			if err != nil {
				fmt.Println("發送回應失敗:", err)
			}
			return
		}

		// 建立圖片嵌入訊息
		embed := &discordgo.MessageEmbed{
			Title: fmt.Sprintf("圖片: %s", imageData.Name),
			Image: &discordgo.MessageEmbedImage{
				URL: imageData.URL,
			},
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral, // 僅使用者可見。
			},
		}
		err = s.InteractionRespond(i.Interaction, response)
		if err != nil {
			fmt.Println("發送回應失敗:", err)
		}

	case "addimage":
		name := i.ApplicationCommandData().Options[0].StringValue()
		url := i.ApplicationCommandData().Options[1].StringValue()

		var category string
		if len(i.ApplicationCommandData().Options) > 2 { //有category參數
			category = i.ApplicationCommandData().Options[2].StringValue()
		} else {
			category = "NULL"
		}

		db, err := database.LoadDatabase(ImgDbFilePath)
		if err != nil {
			fmt.Println("讀取圖庫失敗:", err)
			return
		}

		if db.Categories == nil { //初始化分類
			db.Categories = make(map[string]string)
		}

		categoryID := db.Categories[category]
		if categoryID == "" { //沒提供分類或無此分類
			if category == "NULL" {
				categoryID = "00"
			} else {
				categoryID = fmt.Sprintf("%02d", len(db.Categories)) //分類ID為2位數
			}
			db.Categories[category] = categoryID
		}

		// 計算ID，先填補空缺的ID
		existingIDs := make(map[string]bool)
		for _, img := range db.Images {
			if img.Category == categoryID {
				existingIDs[img.ID] = true
			}
		}

		id := ""
		for i := 1; i <= len(existingIDs)+1; i++ { // 限制迴圈範圍
			candidateID := fmt.Sprintf("%s%03d", categoryID, i)
			if !existingIDs[candidateID] {
				id = candidateID
				break
			}
		}

		if id == "" {
			fmt.Println("無法生成唯一的ID")
			return
		}

		db.Images[name] = database.ImageData{
			URL:      url,
			Name:     name,
			ID:       id,
			Category: categoryID,
		}

		err = database.SaveDatabase(ImgDbFilePath, db)
		if err != nil {
			fmt.Println("上傳圖片失敗:", err)
			return
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("成功添加圖片 %q，分類為：%s，ID為：%s，網址為：%s", name, category, id, url),
				Flags:   discordgo.MessageFlagsEphemeral, // 僅使用者可見。
			},
		}
		err = s.InteractionRespond(i.Interaction, response)
		if err != nil {
			fmt.Println("發送回應失敗:", err)
		}

	case "delimage":
		identifier := i.ApplicationCommandData().Options[0].StringValue()

		db, err := database.LoadDatabase(ImgDbFilePath)
		if err != nil {
			fmt.Println("讀取圖庫失敗:", err)
			return
		}

		var nameToDelete string
		for name, img := range db.Images {
			if img.Name == identifier || img.ID == identifier {
				nameToDelete = name
				break
			}
		}

		if nameToDelete == "" {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("找不到圖片 %q", identifier),
					Flags:   discordgo.MessageFlagsEphemeral, // 僅使用者可見。
				},
			}
			err := s.InteractionRespond(i.Interaction, response)
			if err != nil {
				fmt.Println("發送回應失敗:", err)
			}
			return
		}

		delete(db.Images, nameToDelete)
		err = database.SaveDatabase(ImgDbFilePath, db)
		if err != nil {
			fmt.Println("刪除圖片失敗:", err)
			return
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("成功刪除 %q。", identifier),
				Flags:   discordgo.MessageFlagsEphemeral, // 僅使用者可見。
			},
		}
		err = s.InteractionRespond(i.Interaction, response)
		if err != nil {
			fmt.Println("發送回應失敗:", err)
		}

	case "send":
		identifier := i.ApplicationCommandData().Options[0].StringValue()
		db, err := database.LoadDatabase(ImgDbFilePath)
		if err != nil {
			fmt.Println("讀取圖片庫失敗:", err)
			return
		}

		var imageData database.ImageData
		var found bool

		// 檢查輸入是否為 ID
		for _, img := range db.Images {
			if img.ID == identifier {
				imageData = img
				found = true
				break
			}
		}

		if !found {
			// 使用 SearchImageByName 函數進行搜尋
			matchedID, err := database.SearchImageByName(db, identifier)
			if err != nil || matchedID == "" { //找不到此名稱
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("找不到圖片 %q", identifier),
						Flags:   discordgo.MessageFlagsEphemeral, // 僅使用者可見。
					},
				}
				err = s.InteractionRespond(i.Interaction, response)
				if err != nil {
					fmt.Println("發送回應失敗:", err)
				}
				return
			}

			// 從 db.Images 中尋找對應的圖片資料
			for _, img := range db.Images {
				if img.ID == matchedID {
					imageData = img
					found = true
					break
				}
			}
		}

		if !found { //找不到此ID
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("找不到圖片 %q", identifier),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}
			err = s.InteractionRespond(i.Interaction, response)
			if err != nil {
				fmt.Println("發送回應失敗:", err)
			}
			return
		}

		embed := &discordgo.MessageEmbed{
			Image: &discordgo.MessageEmbedImage{
				URL: imageData.URL,
			},
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("From %s", i.Member.User.Mention()),
				Embeds:  []*discordgo.MessageEmbed{embed},
			},
		}
		err = s.InteractionRespond(i.Interaction, response)
		if err != nil {
			fmt.Println("發送回應失敗:", err)
		}

	case "list":
		var categoryFilter string
		if len(i.ApplicationCommandData().Options) > 0 {
			categoryFilter = i.ApplicationCommandData().Options[0].StringValue()
		}

		db, err := database.LoadDatabase(ImgDbFilePath)
		if err != nil {
			fmt.Println("讀取圖庫失敗:", err)
			return
		}

		var filteredImages []database.ImageData
		if categoryFilter != "" {
			categoryCode := ""
			for name, code := range db.Categories {
				if name == categoryFilter {
					categoryCode = code
					break
				}
			}
			if categoryCode == "" { //無此分類
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("找不到分類 %q。", categoryFilter),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				}
				err := s.InteractionRespond(i.Interaction, response)
				if err != nil {
					fmt.Println("發送回應失敗:", err)
				}
				return
			}
			for _, img := range db.Images {
				if img.Category == categoryCode {
					filteredImages = append(filteredImages, img)
				}
			}
		} else {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "請提供分類來列出圖片。",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}
			err := s.InteractionRespond(i.Interaction, response)
			if err != nil {
				fmt.Println("發送回應失敗:", err)
			}
			return
		}

		// 按ID排序
		sort.Slice(filteredImages, func(i, j int) bool {
			return filteredImages[i].ID < filteredImages[j].ID
		})

		totalImages := len(filteredImages)
		pages := (totalImages + 19) / 20

		currentPage := 0
		if len(i.ApplicationCommandData().Options) > 1 {
			currentPage = int(i.ApplicationCommandData().Options[1].IntValue()) - 1 // 調整為零基索引
		}

		if currentPage < 0 || currentPage >= pages {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("頁數超出範圍。總共 %d 頁。", pages),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}
			err := s.InteractionRespond(i.Interaction, response)
			if err != nil {
				fmt.Println("發送回應失敗:", err)
			}
			return
		}

		start := currentPage * 20
		end := start + 20
		if end > totalImages {
			end = totalImages
		}

		content := ""
		if categoryFilter != "" {
			content += fmt.Sprintf("%s:\n", categoryFilter) //列出分類名稱
		}

		for _, img := range filteredImages[start:end] {
			content += fmt.Sprintf("ID: %s   名稱: %s\n", img.ID, img.Name)
		}

		if content == "" {
			content = "無圖片可顯示。"
		}

		content += fmt.Sprintf("\n第 %d/%d 頁", currentPage+1, pages)

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}
		err = s.InteractionRespond(i.Interaction, response)
		if err != nil {
			fmt.Println("發送回應失敗:", err)
		}

	case "listall":
		// 初始化當前頁數，如果未提供則預設為第 1 頁
		var currentPage int = 1
		if len(i.ApplicationCommandData().Options) > 0 {
			currentPage = int(i.ApplicationCommandData().Options[0].IntValue())
			if currentPage < 1 {
				currentPage = 1 // 確保頁數最小為 1
			}
		}

		// 從資料庫加載圖片數據
		db, err := database.LoadDatabase(ImgDbFilePath)
		if err != nil {
			fmt.Println("讀取圖庫失敗:", err)
			return
		}

		// 將所有圖片添加到切片中
		var allImages []database.ImageData
		for _, img := range db.Images {
			allImages = append(allImages, img)
		}

		// 按ID對圖片進行排序
		sort.Slice(allImages, func(i, j int) bool {
			return allImages[i].ID < allImages[j].ID
		})

		// 計算總圖片數量和頁數
		totalImages := len(allImages)
		pages := (totalImages + 19) / 20

		// 如果當前頁數超出範圍，返回錯誤消息
		if currentPage > pages {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("頁數超出範圍。總共 %d 頁。", pages),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}
			err := s.InteractionRespond(i.Interaction, response)
			if err != nil {
				fmt.Println("發送回應失敗:", err)
			}
			return
		}

		// 計算當前頁面的圖片範圍
		start := (currentPage - 1) * 20
		end := start + 20
		if end > totalImages {
			end = totalImages
		}

		// 構建輸出內容
		content := ""
		lastCategory := ""
		for _, img := range allImages[start:end] {
			// 如果分類變化，添加分類標題
			if img.Category != lastCategory {
				lastCategory = img.Category
				categoryName := "未分類" // 默認為未分類
				for name, code := range db.Categories {
					if code == lastCategory {
						categoryName = name
						break
					}
				}
				if lastCategory == "00" {
					categoryName = "未分類"
				}
				content += fmt.Sprintf("\n%s:\n", categoryName)
			}
			// 添加圖片的ID和名稱
			content += fmt.Sprintf("ID: %s   名稱: %s\n", img.ID, img.Name)
		}

		// 如果沒有內容，顯示無圖片
		if content == "" {
			content = "無圖片可顯示"
		}

		// 添加頁碼
		content += fmt.Sprintf("\n第 %d/%d 頁", currentPage, pages)

		// 發送響應
		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}
		err = s.InteractionRespond(i.Interaction, response)
		if err != nil {
			fmt.Println("發送回應失敗:", err)
		}

	case "classify":
		identifier := i.ApplicationCommandData().Options[0].StringValue()
		newCategory := i.ApplicationCommandData().Options[1].StringValue()

		db, err := database.LoadDatabase(ImgDbFilePath)
		if err != nil {
			fmt.Println("讀取圖庫失敗:", err)
			return
		}

		var imgKey string
		var imageToClassify *database.ImageData
		for key, img := range db.Images {
			if img.ID == identifier || img.Name == identifier {
				imgKey = key
				imageToClassify = &img
				break
			}
		}

		if imageToClassify == nil { //找不到圖片
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("找不到圖片 %q", identifier),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}
			err = s.InteractionRespond(i.Interaction, response)
			if err != nil {
				fmt.Println("發送回應失敗:", err)
			}
			return
		}

		categoryCode := ""
		for name, code := range db.Categories {
			if name == newCategory {
				categoryCode = code
				break
			}
		}

		if categoryCode == "" {
			newCode := fmt.Sprintf("%02d", len(db.Categories)+1)
			db.Categories[newCategory] = newCode
			categoryCode = newCode
		}

		// 更新分類並重新分配ID
		existingIDs := make(map[string]bool)
		for _, img := range db.Images {
			if img.Category == categoryCode {
				existingIDs[img.ID] = true
			}
		}

		// 分配新的ID
		for i := 1; ; i++ {
			newID := fmt.Sprintf("%s%03d", categoryCode, i)
			if !existingIDs[newID] {
				imageToClassify.ID = newID
				break
			}
		}

		imageToClassify.Category = categoryCode
		db.Images[imgKey] = *imageToClassify
		err = database.SaveDatabase(ImgDbFilePath, db)
		if err != nil {
			fmt.Println("儲存圖庫失敗:", err)
			return
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("成功將圖片 %q 分類到 %q，新的ID為 %q", imageToClassify.Name, newCategory, imageToClassify.ID),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}
		err = s.InteractionRespond(i.Interaction, response)
		if err != nil {
			fmt.Println("發送回應失敗:", err)
		}
	}

}
