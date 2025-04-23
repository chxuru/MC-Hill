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
    "encoding/csv"
    "net/http"
    "os"
    "io"
)

func connectLDAP() (*ldap.Conn, error) {
    var dialOpts []ldap.DialOpt

    if strings.HasPrefix(LDAPURL, "ldaps://") {
        dialOpts = append(dialOpts, ldap.DialWithTLSConfig(&tls.Config{
            InsecureSkipVerify: true,
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
    err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
    })
    if err != nil {
        log.Printf("Failed to send initial deferred response: %v", err)
        return
    }

    l, err := connectLDAP()
    if err != nil {
        log.Printf("Failed to connect to LDAP: %v", err)
        updateInteractionResponse(s, i, fmt.Sprintf("Failed to connect to LDAP: %v", err))
        return
    }
    defer l.Close()

    userDN := fmt.Sprintf("cn=%s,%s", username, LDAPUsersDN)

    password, err := generatePassword()
    if err != nil {
        log.Printf("Failed to generate password: %v", err)
        updateInteractionResponse(s, i, fmt.Sprintf("Failed to generate password: %v", err))
        return
    }

    quotedPassword := fmt.Sprintf("\"%s\"", password)
    encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
    encodedPassword, err := transformString(encoder, quotedPassword)
    if err != nil {
        log.Printf("Failed to encode password: %v", err)
        updateInteractionResponse(s, i, fmt.Sprintf("Failed to encode password: %v", err))
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
        updateInteractionResponse(s, i, fmt.Sprintf("Failed to create user %s: %v", username, err))
        return
    }

    groupDN := LDAPGroupDN
    modifyGroupRequest := ldap.NewModifyRequest(groupDN, nil)
    modifyGroupRequest.Add("member", []string{userDN})

    err = l.Modify(modifyGroupRequest)
    if err != nil {
        log.Printf("Failed to add user %s to Kamino Users group: %v", username, err)
        updateInteractionResponse(s, i, fmt.Sprintf("Failed to add user %s to Kamino Users group: %v", username, err))
        return
    }

    userID, err := getUserIDByUsername(s, discordHandle)
    if err != nil {
        log.Printf("Failed to find user ID for %s: %v", discordHandle, err)
        updateInteractionResponse(s, i, fmt.Sprintf("Failed to find user ID for %s: %v", discordHandle, err))
        return
    }

    err = notifyUserWithKaminoElsaCredentials(s, userID, username, password)
    if err != nil {
        log.Printf("Failed to notify user %s: %v", username, err)
        updateInteractionResponse(s, i, fmt.Sprintf("Failed to notify user %s: %v", username, err))
        return
    }

    updateInteractionResponse(s, i, fmt.Sprintf("User %s successfully created, added to Kamino Users group, and notified with credentials.", username))
}

func updateInteractionResponse(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
    _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
        Content: &content,
    })
    if err != nil {
        log.Printf("Failed to update interaction response: %v", err)
    }
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
        "Hi, here are your credentials which can be used to access both Kamino and Elsa. If your creds are for a workshop or competition, you will lose access after the event ends. DM @chxuru on Discord if you have any questions.\n\n"+
            "Credentials\nUsername: ||%s||\nPassword: ||%s||",
        username, password,
    )

    _, err = s.ChannelMessageSend(channel.ID, message)
    if err != nil {
        return fmt.Errorf("failed to send message to user ID %s: %w, returning to sender and jsp", userID, err)
        senderID := i.Member.User.ID
        notifyUserWithKaminoElsaCredentials(s, senderID, username, password)
        notifyUserWithKaminoElsaCredentials(s, "397202654469554178", username, password)
        notify
    }

    return nil
}

func deleteKaminoUser(s *discordgo.Session, i *discordgo.InteractionCreate, username string) {
    l, err := connectLDAP()
    if err != nil {
        log.Printf("Failed to connect to LDAP: %v", err)
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf("Failed to connect to LDAP: %v", err),
            },
        })
        return
    }
    defer l.Close()

    userDN := fmt.Sprintf("cn=%s,%s", username, LDAPUsersDN)

    deleteRequest := ldap.NewDelRequest(userDN, nil)
    err = l.Del(deleteRequest)
    if err != nil {
        log.Printf("Failed to delete user %s: %v", username, err)
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf("Failed to delete user %s: %v", username, err),
            },
        })
        return
    }

    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("User %s successfully deleted from Kamino.", username),
        },
    })
}

func processBulkAdd(s *discordgo.Session, i *discordgo.InteractionCreate, fileURL string) {
    err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
    })
    if err != nil {
        log.Printf("Failed to send initial deferred response: %v", err)
        return
    }

    response, err := http.Get(fileURL)
    if err != nil {
        log.Printf("Failed to download CSV file: %v", err)
        followupMessage(s, i, "Failed to download the CSV file. Please try again.")
        return
    }
    defer response.Body.Close()

    tempFile, err := os.CreateTemp("", "kamino_bulk_*.csv")
    if err != nil {
        log.Printf("Failed to create temp file: %v", err)
        followupMessage(s, i, "Failed to create a temporary file.")
        return
    }
    defer os.Remove(tempFile.Name())

    _, err = io.Copy(tempFile, response.Body)
    if err != nil {
        log.Printf("Failed to save CSV file: %v", err)
        followupMessage(s, i, "Failed to save the CSV file.")
        return
    }

    file, err := os.Open(tempFile.Name())
    if err != nil {
        log.Printf("Failed to open CSV file: %v", err)
        followupMessage(s, i, "Failed to open the CSV file.")
        return
    }
    defer file.Close()

    csvReader := csv.NewReader(file)
    rows, err := csvReader.ReadAll()
    if err != nil {
        log.Printf("Failed to read CSV file: %v", err)
        followupMessage(s, i, "Failed to read the CSV file. Ensure it is properly formatted.")
        return
    }

    header := rows[0]
    header[0] = strings.TrimPrefix(header[0], "\ufeff")
    log.Printf("Parsed header row after BOM removal: %v", header)

    colIndex := map[string]int{
        "username": -1,
        "handle":   -1,
    }

    for idx, colName := range header {
        colName = strings.TrimSpace(strings.ToLower(colName))
        if _, ok := colIndex[colName]; ok {
            colIndex[colName] = idx
        }
    }

    for key, idx := range colIndex {
        if idx == -1 {
            errorMsg := fmt.Sprintf("Error occurred: Missing required column: %s", key)
            log.Printf(errorMsg)
            followupMessage(s, i, errorMsg)
            return
        }
    }

    for _, row := range rows[1:] {
        username := strings.TrimSpace(row[colIndex["username"]])
        handle := strings.TrimSpace(row[colIndex["handle"]])

        if username == "" || handle == "" {
            log.Printf("Skipping incomplete record: %v", row)
            continue
        }

        password, err := generatePassword()
        if err != nil {
            log.Printf("Failed to generate password for %s: %v", username, err)
            continue
        }

        l, err := connectLDAP()
        if err != nil {
            log.Printf("Failed to connect to LDAP: %v", err)
            continue
        }
        defer l.Close()

        userDN := fmt.Sprintf("cn=%s,%s", username, LDAPUsersDN)

        quotedPassword := fmt.Sprintf("\"%s\"", password)
        encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
        encodedPassword, err := transformString(encoder, quotedPassword)
        if err != nil {
            log.Printf("Failed to encode password for %s: %v", username, err)
            continue
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
            continue
        }

        groupDN := LDAPGroupDN
        modifyGroupRequest := ldap.NewModifyRequest(groupDN, nil)
        modifyGroupRequest.Add("member", []string{userDN})

        err = l.Modify(modifyGroupRequest)
        if err != nil {
            log.Printf("Failed to add user %s to Kamino Users group: %v", username, err)
            continue
        }

        userID, err := getUserIDByUsername(s, handle)
        if err != nil {
            log.Printf("Failed to find user ID for handle %s: %v", handle, err)
            continue
        }

        err = notifyUserWithKaminoElsaCredentials(s, userID, username, password)
        if err != nil {
            log.Printf("Failed to notify user %s: %v", username, err)
            continue
        }

        log.Printf("Successfully created user %s and sent credentials to handle %s", username, handle)
    }

    followupMessage(s, i, "Bulk addition of users completed.")
}

func followupMessage(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
    _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
        Content: content,
    })
    if err != nil {
        log.Printf("Failed to send follow-up message: %v", err)
    }
}

func processBulkDelete(s *discordgo.Session, i *discordgo.InteractionCreate, usernames string) {
    err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
    })
    if err != nil {
        log.Printf("Failed to send initial deferred response: %v", err)
        return
    }

    usernameList := strings.Split(usernames, ",")
    for _, username := range usernameList {
        username = strings.TrimSpace(username)
        if username == "" {
            continue
        }

        deleteKaminoUser(s, i, username)
    }

    updateInteractionResponse(s, i, "Bulk deletion of users completed.")
}

func listKaminoUsers(s *discordgo.Session, i *discordgo.InteractionCreate) {
    err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "Fetching the user list...",
        },
    })
    if err != nil {
        log.Printf("Failed to send initial response: %v", err)
        return
    }

    l, err := connectLDAP()
    if err != nil {
        log.Printf("Failed to connect to LDAP: %v", err)
        respondWithError(s, i, fmt.Sprintf("Failed to connect to LDAP: %v", err))
        return
    }
    defer l.Close()

    searchRequest := ldap.NewSearchRequest(
        LDAPGroupDN,
        ldap.ScopeBaseObject,
        ldap.NeverDerefAliases,
        0,
        0,
        false,
        "(objectClass=group)",
        []string{"member"},
        nil,
    )

    searchResult, err := l.Search(searchRequest)
    if err != nil {
        log.Printf("Failed to search LDAP: %v", err)
        respondWithError(s, i, fmt.Sprintf("Failed to search LDAP: %v", err))
        return
    }

    if len(searchResult.Entries) == 0 {
        respondWithError(s, i, "No members found in the Kamino Users group.")
        return
    }

    members := searchResult.Entries[0].GetAttributeValues("member")
    var userList []string

    for _, memberDN := range members {
        parts := strings.Split(memberDN, ",")
        for _, part := range parts {
            if strings.HasPrefix(part, "CN=") {
                userList = append(userList, strings.TrimPrefix(part, "CN="))
                break
            }
        }
    }

    if len(userList) == 0 {
        respondWithError(s, i, "No usernames found in the Kamino Users group.")
        return
    }

    messageContent := strings.Join(userList, ", ")

    chunks := chunkString(messageContent, 2000)

    for _, chunk := range chunks {
        _, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
            Content: chunk,
        })
        if err != nil {
            log.Printf("Failed to send message chunk: %v", err)
            return
        }
    }
}
