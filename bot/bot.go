package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/etas94/godcbot/config"
	"github.com/etas94/godcbot/database"
)

var BotId string
var goBot *discordgo.Session

const ImgDbFilePath = "./image_url.json"

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

	// 獲取機器人ID
	u, err := goBot.User("@me")
	if err != nil {
		fmt.Println("獲取ID失敗:", err)
		return
	}
	BotId = u.ID

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
			Description: "根據名稱獲取圖片",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "name",
					Description: "圖片的名稱",
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
			},
		},
		{
			Name:        "delimage",
			Description: "從圖庫中刪除圖片",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "name",
					Description: "圖片的名稱",
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
					Name:        "name",
					Description: "圖片的名稱",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
	}

	//註冊指令。
	for _, cmd := range commands {
		_, err := goBot.ApplicationCommandCreate(goBot.State.User.ID, "", cmd)
		if err != nil {
			fmt.Printf("無法創建 '%s' 指令: %v\n", cmd.Name, err)
		}
	}
}

// 處理Slash Command。
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
		name := i.ApplicationCommandData().Options[0].StringValue()
		db, err := database.LoadDatabase(ImgDbFilePath)
		if err != nil {
			fmt.Println("讀取圖片庫失敗:", err)
			return
		}

		imageURL, found := db.Images[name]
		if !found {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("找不到圖片 %q", name),
					Flags:   discordgo.MessageFlagsEphemeral, // 僅使用者可見。
				},
			}
			err := s.InteractionRespond(i.Interaction, response)
			if err != nil {
				fmt.Println("發送互動回應失敗:", err)
			}
			return
		}

		embed := &discordgo.MessageEmbed{
			Title: fmt.Sprintf("圖片: %s", name),
			Image: &discordgo.MessageEmbedImage{
				URL: imageURL,
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

		db, err := database.LoadDatabase(ImgDbFilePath)
		if err != nil {
			fmt.Println("讀取圖庫失敗:", err)
			return
		}

		db.Images[name] = url
		err = database.SaveDatabase(ImgDbFilePath, db)
		if err != nil {
			fmt.Println("上傳圖片失敗:", err)
			return
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("成功添加圖片 %q，網址為：%s", name, url),
				Flags:   discordgo.MessageFlagsEphemeral, // 僅使用者可見。
			},
		}
		err = s.InteractionRespond(i.Interaction, response)
		if err != nil {
			fmt.Println("發送回應失敗:", err)
		}

	case "delimage":
		name := i.ApplicationCommandData().Options[0].StringValue()

		db, err := database.LoadDatabase(ImgDbFilePath)
		if err != nil {
			fmt.Println("讀取圖庫失敗:", err)
			return
		}

		if _, found := db.Images[name]; !found {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("找不到 %q", name),
					Flags:   discordgo.MessageFlagsEphemeral, // 僅使用者可見。
				},
			}
			err := s.InteractionRespond(i.Interaction, response)
			if err != nil {
				fmt.Println("發送回應失敗:", err)
			}
			return
		}

		delete(db.Images, name)
		err = database.SaveDatabase(ImgDbFilePath, db)
		if err != nil {
			fmt.Println("刪除圖片失敗:", err)
			return
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("成功刪除 %q。", name),
				Flags:   discordgo.MessageFlagsEphemeral, // 僅使用者可見。
			},
		}
		err = s.InteractionRespond(i.Interaction, response)
		if err != nil {
			fmt.Println("發送回應失敗:", err)
		}

	case "send":
		name := i.ApplicationCommandData().Options[0].StringValue()

		db, err := database.LoadDatabase(ImgDbFilePath)
		if err != nil {
			fmt.Println("讀取圖庫失敗:", err)
			return
		}

		imageURL, found := db.Images[name]
		if !found {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("找不到圖片 %q", name),
					Flags:   discordgo.MessageFlagsEphemeral, // 僅使用者可見。
				},
			}
			err := s.InteractionRespond(i.Interaction, response)
			if err != nil {
				fmt.Println("發送回應失敗:", err)
			}
			return
		}

		embed := &discordgo.MessageEmbed{
			Image: &discordgo.MessageEmbedImage{
				URL: imageURL,
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
	}
}
