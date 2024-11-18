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
                {Type: discordgo.ApplicationCommandOptionString, Name: "name", Description: "Full name", Required: true},
                {Type: discordgo.ApplicationCommandOptionString, Name: "handle", Description: "Discord handle", Required: true},
            },
        },
        {Name: "display", Description: "Display all profiles"},
        {Name: "conns", Description: "Display active connections"},
        {Name: "status", Description: "Check bot status"},
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
        {
            Name:        "kill",
            Description: "Kill a VPN session by username",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Type:        discordgo.ApplicationCommandOptionString,
                    Name:        "username",
                    Description: "Username of the VPN connection to kill",
                    Required:    true,
                },
            },
        },
        {
            Name:        "kamino",
            Description: "Manage Kamino users",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Type:        discordgo.ApplicationCommandOptionSubCommand,
                    Name:        "add",
                    Description: "Add a Kamino user",
                    Options: []*discordgo.ApplicationCommandOption{
                        {
                            Type:        discordgo.ApplicationCommandOptionString,
                            Name:        "username",
                            Description: "username",
                            Required:    true,
                        },
                        {
                            Type:        discordgo.ApplicationCommandOptionString,
                            Name:        "handle",
                            Description: "Discord handle",
                            Required:    true,
                        },
                    },
                },
                {
                    Type:        discordgo.ApplicationCommandOptionSubCommand,
                    Name:        "add-bulk",
                    Description: "Add multiple Kamino users",
                    Options: []*discordgo.ApplicationCommandOption{
                        {
                            Type:        discordgo.ApplicationCommandOptionAttachment,
                            Name:        "csvfile",
                            Description: "CSV file with Username & Handle column headers",
                            Required:    true,
                        },
                    },
                },
                {
                    Type:        discordgo.ApplicationCommandOptionSubCommand,
                    Name:        "delete",
                    Description: "Delete Kamino users by username",
                    Options: []*discordgo.ApplicationCommandOption{
                        {
                            Type:        discordgo.ApplicationCommandOptionString,
                            Name:        "usernames",
                            Description: "Comma-separated list of usernames to delete",
                            Required:    true,
                        },
                    },
                },
                {
                    Type:        discordgo.ApplicationCommandOptionSubCommand,
                    Name:        "display",
                    Description: "Display all users in the Kamino Users group",
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
