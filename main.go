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

    discord.AddHandler(handleInteraction)

    err = discord.Open()
    if err != nil {
        log.Fatalf("Failed to open Discord session: %v", err)
    }

    log.Println("Bot is running!")
    select {}
}

func handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
    if i.Type == discordgo.InteractionApplicationCommand {
        guildID := i.GuildID
        log.Printf("Command executed in Guild ID: %s", guildID)

        err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "This command is being executed in your server!",
            },
        })
        if err != nil {
            log.Printf("Error responding to interaction: %v", err)
        }
    }
}
