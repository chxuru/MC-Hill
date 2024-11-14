package main

import (
    "crypto/tls"
    "fmt"
    "log"
    "os/exec"
    "strings"
    "github.com/go-ldap/ldap/v3"
    "github.com/bwmarrin/discordgo"
    "golang.org/x/text/encoding/unicode"
    "golang.org/x/text/transform"
    "bytes"
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

func generatePassword() (string, error) {
    wordsCmd := exec.Command("bw", "generate", "--passphrase", "--words", "3", "--separator", "-")
    wordsOutput, err := wordsCmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to generate words for password: %v", err)
    }
    words := strings.ReplaceAll(strings.TrimSpace(string(wordsOutput)), "-", "")

    numberCmd := exec.Command("bash", "-c", "bw generate --length 5 --number | head -c 1")
    numberOutput, err := numberCmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to generate number for password: %v", err)
    }
    number := strings.TrimSpace(string(numberOutput))

    specialCmd := exec.Command("bash", "-c", "bw generate --length 5 --special | head -c 1")
    specialOutput, err := specialCmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to generate special character for password: %v", err)
    }
    specialChar := strings.TrimSpace(string(specialOutput))

    newPassword := words + number + specialChar
    return newPassword, nil
}

func createUserAndAddToGroup(s *discordgo.Session, i *discordgo.InteractionCreate, username string, discordHandle string) {
    l, err := connectLDAP()
    if err != nil {
        log.Printf("Failed to connect to LDAP: %v", err)
        return
    }
    defer l.Close()

    userDN := fmt.Sprintf("cn=%s,%s", username, LDAPUsersDN)

    password, err := generatePassword()
    if err != nil {
        log.Printf("Failed to generate password: %v", err)
        return
    }

    quotedPassword := fmt.Sprintf("\"%s\"", password)
    encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
    encodedPassword, err := transformString(encoder, quotedPassword)
    if err != nil {
        log.Printf("Failed to encode password: %v", err)
        return
    }

    addRequest := ldap.NewAddRequest(userDN, nil)
    addRequest.Attribute("objectClass", []string{"top", "person", "organizationalPerson", "user"})
    addRequest.Attribute("cn", []string{username})
    addRequest.Attribute("sAMAccountName", []string{username})
    addRequest.Attribute("unicodePwd", []string{encodedPassword})
    addRequest.Attribute("userPrincipalName", []string{username + "@sdc.cpp"})
    addRequest.Attribute("displayName", []string{username})
    addRequest.Attribute("userAccountControl", []string{"512"})

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

    userID, err := getUserIDByUsername(s, i.GuildID, discordHandle)
    if err != nil {
        log.Printf("Failed to find user ID for %s: %v", discordHandle, err)
        return
    }

    err = notifyUserWithKaminoElsaCredentials(s, userID, username, password)
    if err != nil {
        log.Printf("Failed to notify user %s: %v", username, err)
        return
    }

    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("User %s successfully created, added to Kamino Users group, and notified with credentials.", username),
        },
    })
}

func transformString(t transform.Transformer, s string) (string, error) {
    var buf bytes.Buffer
    writer := transform.NewWriter(&buf, t)
    _, err := writer.Write([]byte(s))
    if err != nil {
        return "", err
    }
    writer.Close()
    return buf.String(), nil
}

func notifyUserWithKaminoElsaCredentials(s *discordgo.Session, userID, username, password string) error {
    channel, err := s.UserChannelCreate(userID)
    if err != nil {
        return fmt.Errorf("failed to create DM channel for user ID %s: %w", userID, err)
    }

    message := fmt.Sprintf(
        "Hi, here are your credentials which can be used to access both Kamino and Elsa. DM @chxuru on Discord if you have any questions.\n\n"+
            "Credentials\nUsername: ||%s||\nPassword: ||%s||",
        username, password,
    )

    _, err = s.ChannelMessageSend(channel.ID, message)
    if err != nil {
        return fmt.Errorf("failed to send message to user ID %s: %w", userID, err)
    }

    return nil
}
