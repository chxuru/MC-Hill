package main

import (
    "log"

    "github.com/bwmarrin/discordgo"
)

func handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
    const allowedChannelID = "1304231659746627634"
    allowedChannelIDs := []string{"1364677031576600837", "1304231659746627634"}

    allowed := false
    for _, id := range allowedChannelIDs {
        if i.ChannelID == id {
            allowed = true
            break
        }
    }
    
    if allowed == false {
        err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "You can't run commands here",
            },
        })
        if err != nil {
            log.Printf("Error responding to interaction: %v", err)
        }
        return
    }

    if i.Type == discordgo.InteractionApplicationCommand {
        switch i.ApplicationCommandData().Name {
        /*depreciated
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
        case "kill":
            handleKillCommand(s, i)*/
        case "status":
            handleStatusCommand(s, i)
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

    case "add-bulk":
        if len(options) < 1 {
            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Content: "Error: Missing required option. Please upload a CSV file.",
                },
            })
            log.Println("Error: Missing required CSV file for /kamino add-bulk command.")
            return
        }
    
        attachmentID := options[0].Value.(string)
        attachment := i.ApplicationCommandData().Resolved.Attachments[attachmentID]
        if attachment == nil {
            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Content: "Error: Could not retrieve the uploaded CSV file.",
                },
            })
            log.Println("Error: Could not retrieve the uploaded CSV file.")
            return
        }
    
        log.Printf("Processing CSV file: %s", attachment.URL)
        processBulkAdd(s, i, attachment.URL)

    case "delete":
        if len(options) < 1 {
            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Content: "Error: Missing required option. Please provide a comma-separated list of usernames.",
                },
            })
            log.Println("Error: Missing required usernames for /kamino delete-bulk command.")
            return
        }

        usernames := options[0].StringValue()
        log.Printf("Processing bulk delete for usernames: %s", usernames)
        processBulkDelete(s, i, usernames)

    case "display":
        listKaminoUsers(s, i)
    }
}
