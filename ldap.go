package main

import (
    "crypto/tls"
    "fmt"
    "log"
    "github.com/go-ldap/ldap/v3"
    "github.com/bwmarrin/discordgo"
    "strings"
)

func connectLDAP() (*ldap.Conn, error) {
    var dialOpts []ldap.DialOpt

    if strings.HasPrefix(LDAPURL, "ldaps://") {
        dialOpts = append(dialOpts, ldap.DialWithTLSConfig(&tls.Config{
            InsecureSkipVerify: LDAPInsecureTLS,
            MinVersion:         tls.VersionTLS12,
        }))
    } else {
        return nil, fmt.Errorf("only ldaps:// connections are supported")
    }

    l, err := ldap.DialURL(LDAPURL, dialOpts...)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to LDAP server: %v", err)
    }

    err = l.Bind(LDAPBindDN, LDAPBindPassword)
    if err != nil {
        l.Close()
        return nil, fmt.Errorf("failed to bind to LDAP server: %v", err)
    }

    return l, nil
}

func createUserAndAddToGroup(s *discordgo.Session, i *discordgo.InteractionCreate, username string) {
    l, err := connectLDAP()
    if err != nil {
        log.Printf("Failed to connect to LDAP: %v", err)
        return
    }
    defer l.Close()

    userDN := fmt.Sprintf("cn=%s,%s", username, LDAPUsersDN)

    addRequest := ldap.NewAddRequest(userDN, nil)
    addRequest.Attribute("objectClass", []string{"top", "person", "organizationalPerson", "user"})
    addRequest.Attribute("cn", []string{username})
    addRequest.Attribute("sAMAccountName", []string{username})
    addRequest.Attribute("userPassword", []string{"testpassword123!"})
    addRequest.Attribute("userPrincipalName", []string{username + "@sdc.cpp"})
    addRequest.Attribute("displayName", []string{username})

    err = l.Add(addRequest)
    if err != nil {
        log.Printf("Failed to create user %s: %v", username, err)
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf("Failed to create user %s: %v", username, err),
            },
        })
        return
    }

    modifyRequest := ldap.NewModifyRequest(userDN, nil)
    modifyRequest.Replace("userAccountControl", []string{"512"})

    err = l.Modify(modifyRequest)
    if err != nil {
        log.Printf("Failed to enable user %s: %v", username, err)
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf("Failed to enable user %s: %v", username, err),
            },
        })
        return
    }

    groupDN := LDAPGroupDN
    modifyGroupRequest := ldap.NewModifyRequest(groupDN, nil)
    modifyGroupRequest.Add("member", []string{userDN})

    err = l.Modify(modifyGroupRequest)
    if err != nil {
        log.Printf("Failed to add user %s to Kamino Users group: %v", username, err)
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf("Failed to add user %s to Kamino Users group: %v", username, err),
            },
        })
        return
    }

    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("User %s successfully created and added to Kamino Users group.", username),
        },
    })
}

func handleKaminoAddCommand(s *discordgo.Session, i *discordgo.InteractionCreate, username string) {
    createUserAndAddToGroup(s, i, username)
}

func handleKaminoDeleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate, username string) {
    response := fmt.Sprintf("Deletion of user %s from Kamino Users group is not currently implemented.", username)
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: response,
        },
    })
}
