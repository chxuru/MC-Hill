package main

import (
    "github.com/bwmarrin/discordgo"
    "log"
)

func registerCommands(discord *discordgo.Session) {
    commands := []discordgo.ApplicationCommand{
        {
            Name:        "multi",
            Description: "multiple profiles",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Type:        discordgo.ApplicationCommandOptionAttachment,
                    Name:        "csvfile",
                    Description: "CSV file containing user details",
                    Required:    true,
                },
            },
        },
        {
            Name:        "single",
            Description: "single profile",
            Options: []*discordgo.ApplicationCommandOption{
                {Type: discordgo.ApplicationCommandOptionString, Name: "username", Description: "VPN username", Required: true},
                {Type: discordgo.ApplicationCommandOptionString, Name: "password", Description: "VPN password", Required: true},
                {Type: discordgo.ApplicationCommandOptionString, Name: "name", Description: "Full name", Required: true},
                {Type: discordgo.ApplicationCommandOptionString, Name: "handle", Description: "Discord handle", Required: true},
            },
        },
        {Name: "display", Description: "Display all profiles"},
        {Name: "conns", Description: "Display active connections"},
        {
            Name:        "delete",
            Description: "Delete users by ID",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Type:        discordgo.ApplicationCommandOptionString,
                    Name:        "userids",
                    Description: "Comma-separated list of user IDs to delete",
                    Required:    true,
                },
            },
        },
    }

    for _, command := range commands {
        _, err := discord.ApplicationCommandCreate(discord.State.User.ID, "", &command)
        if err != nil {
            log.Fatalf("Failed to register command %s: %v", command.Name, err)
        }
    }
}
