package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/etas94/godcbot/config"
)

var BotId string
var goBot *discordgo.Session

func Start() {
	cfg, err := config.ReadConfig()
	if err != nil {
		fmt.Println("Failed reading configuration:", err)
		return
	}

	goBot, err = discordgo.New("Bot " + cfg.Token)
	if err != nil {
		fmt.Println("Failed initializing Discord Session:", err)
		return
	}

	u, err := goBot.User("@me")
	if err != nil {
		fmt.Println("Failed getting current User:", err)
		return
	}

	BotId = u.ID

	goBot.AddHandler(func(s *discordgo.Session, e *discordgo.MessageCreate) {
		messageHandler(s, e, cfg)
	})

	err = goBot.Open()
	if err != nil {
		fmt.Println("Failed opening connection to Discord:", err)
		return
	}

	fmt.Println("Bot is now connected!")
}

func messageHandler(s *discordgo.Session, e *discordgo.MessageCreate, cfg *config.Config) {
	// 忽略機器人自己發的消息
	if e.Author.ID == BotId {
		return
	}

	// 從伺服器或私信中檢查消息
	prefix := cfg.BotPrefix
	if strings.HasPrefix(e.Content, prefix) {
		args := strings.Fields(e.Content)
		cmd := args[0][len(prefix):]
		//arguments := args[1:]

		switch cmd {
		case "ping":
			_, err := s.ChannelMessageSend(e.ChannelID, "Pong!")
			if err != nil {
				fmt.Println("Failed sending Pong response:", err)
			}
		default:
			_, err := s.ChannelMessageSend(e.ChannelID, fmt.Sprintf("Unknown command %q.", cmd))
			if err != nil {
				fmt.Println("Failed sending Unknown Command response:", err)
			}
		}
	}
}
