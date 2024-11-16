package main

import (
    "github.com/bwmarrin/discordgo"
    "log"
)

func main() {
    discord, err := discordgo.New("Bot " + DiscordBotToken)
    if err != nil {
        log.Fatalf("Failed to create Discord session: %v", err)
    }
    defer discord.Close()

    err = discord.Open()
    if err != nil {
        log.Fatalf("Failed to open Discord session: %v", err)
    }

    log.Println("Bot is running!")
    select {}
}
