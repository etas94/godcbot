package main

import (
	"github.com/etas94/godcbot/bot"
)

func main() {
	bot.Start()

	<-make(chan struct{})
}
