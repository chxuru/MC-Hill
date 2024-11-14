package main

import (
    "fmt"
    "log"
    "github.com/go-ldap/ldap/v3"
    "crypto/tls"
    "github.com/bwmarrin/discordgo"
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

func AddUserToGroup(username string) error {
    l, err := connectLDAP()
    if err != nil {
        return err
    }
    defer l.Close()
    fmt.Println("LDAP_INSECURE_TLS:", LDAPInsecureTLS)
    userDN := fmt.Sprintf("%s=%s,%s", LDAPUserAttribute, username, LDAPUsersDN)
    modifyRequest := ldap.NewModifyRequest(LDAPGroupDN, nil)
    modifyRequest.Add("member", []string{userDN})
    fmt.Println("LDAP_INSECURE_TLS:", LDAPInsecureTLS)
    err = l.Modify(modifyRequest)
    if err != nil {
        return fmt.Errorf("failed to add user to Kamino group: %v", err)
    }

    log.Printf("User %s added to Kamino group", username)
    return nil
}

func DeleteUserFromGroup(username string) error {
    l, err := connectLDAP()
    if err != nil {
        return err
    }
    defer l.Close()

    userDN := fmt.Sprintf("%s=%s,%s", LDAPUserAttribute, username, LDAPUsersDN)
    modifyRequest := ldap.NewModifyRequest(LDAPGroupDN, nil)
    modifyRequest.Delete("member", []string{userDN})

    err = l.Modify(modifyRequest)
    if err != nil {
        return fmt.Errorf("failed to remove user from Kamino group: %v", err)
    }

    log.Printf("User %s removed from Kamino group", username)
    return nil
}

func handleKaminoAddCommand(s *discordgo.Session, i *discordgo.InteractionCreate, username string) {
    fmt.Println("LDAP_INSECURE_TLS:", LDAPInsecureTLS)
    err := AddUserToGroup(username)
    var response string
    if err != nil {
        response = fmt.Sprintf("Failed to add user %s to Kamino group: %v", username, err)
    } else {
        response = fmt.Sprintf("User %s successfully added to Kamino group.", username)
    }
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: response,
        },
    })
}

func handleKaminoDeleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate, username string) {
    err := DeleteUserFromGroup(username)
    var response string
    if err != nil {
        response = fmt.Sprintf("Failed to remove user %s from Kamino group: %v", username, err)
    } else {
        response = fmt.Sprintf("User %s successfully removed from Kamino group.", username)
    }
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: response,
        },
    })
}
