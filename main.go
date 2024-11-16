package main

import (
    "github.com/bwmarrin/discordgo"
    "log"
)

var GuildID string

func main() {
    discord, err := discordgo.New("Bot " + DiscordBotToken)
    if err != nil {
        log.Fatalf("Failed to create Discord session: %v", err)
    }
    defer discord.Close()

    discord.AddHandler(handleReady)
    discord.AddHandler(handleInteraction)

    err = discord.Open()
    if err != nil {
        log.Fatalf("Failed to open Discord session: %v", err)
    }

    registerCommands(discord)
    log.Println("Bot is running!")
    select {}
}
