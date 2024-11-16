package main

import (
    "log"

    "github.com/bwmarrin/discordgo"
)

func handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
    if i.GuildID == "" {
        log.Println("Ignoring DM interaction")
        return
    }

    if i.Type == discordgo.InteractionApplicationCommand {
        switch i.ApplicationCommandData().Name {
        case "multi":
            handleProfileCommand(s, i)
        case "single":
            handleSingleProfileCommand(s, i)
        case "display":
            handleDisplayCommand(s, i)
        case "conns":
            handleConnsCommand(s, i)
        case "delete":
            handleDeleteCommand(s, i)
        case "status":
            handleStatusCommand(s, i)
        case "kill":
            handleKillCommand(s, i)
        case "kamino":
            handleKaminoCommand(s, i)
        }
    }
}

func handleKaminoCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
    subcommand := i.ApplicationCommandData().Options[0].Name
    options := i.ApplicationCommandData().Options[0].Options

    switch subcommand {
    case "add":
        if len(options) < 2 {
            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Content: "Error: Missing required options. Please provide both username and discord handle.",
                },
            })
            log.Println("Error: Missing required options for /kamino add command.")
            return
        }

        username := options[0].StringValue()
        discordHandle := options[1].StringValue()
        createUserAndAddToGroup(s, i, username, discordHandle)

    case "delete":
        if len(options) < 1 {
            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Content: "Error: Missing required option. Please provide the username to delete.",
                },
            })
            log.Println("Error: Missing required option for /kamino delete command.")
            return
        }

        username := options[0].StringValue()
        deleteKaminoUser(s, i, username)
    }
}
