package main

import (
    "fmt"
    "github.com/PuerkitoBio/goquery"
    "github.com/bwmarrin/discordgo"
    "io"
    "net/http"
    "net/url"
    "strings"
)

func loginToPfSense(client *http.Client) (bool, error) {
    loginPage, err := client.Get(BaseURL + LoginPath)
    if err != nil {
        return false, err
    }
    defer loginPage.Body.Close()

    csrfToken, err := getCSRFToken(loginPage)
    if err != nil {
        return false, err
    }

    loginPayload := url.Values{
        "__csrf_magic": {csrfToken},
        "usernamefld":  {AdminUsername},
        "passwordfld":  {AdminPassword},
        "login":        {"Sign In"},
    }

    headers := map[string]string{
        "Content-Type": "application/x-www-form-urlencoded",
        "Referer":      BaseURL + LoginPath,
    }

    loginResp, err := postFormWithHeaders(client, BaseURL+LoginPath, loginPayload, headers)
    if err != nil {
        return false, err
    }
    defer loginResp.Body.Close()

    loginBody, err := io.ReadAll(loginResp.Body)
    if err != nil {
        return false, err
    }

    return strings.Contains(string(loginBody), "System Information"), nil
}

func createUser(client *http.Client, username, password, description string) error {
    userCreationPage, err := client.Get(BaseURL + UserManagerPath)
    if err != nil {
        return fmt.Errorf("failed to fetch user creation page: %w", err)
    }
    defer userCreationPage.Body.Close()

    csrfToken, err := getCSRFToken(userCreationPage)
    if err != nil {
        return fmt.Errorf("failed to retrieve CSRF token: %w", err)
    }

    createUserPayload := url.Values{
        "__csrf_magic":      {csrfToken},
        "usernamefld":       {username},
        "passwordfld1":      {password},
        "passwordfld2":      {password},
        "descr":             {description},
        "expires":           {""},
        "webguicss":         {"pfSense.css"},
        "webguifixedmenu":   {""},
        "webguihostnamemenu":{""},
        "dashboardcolumns":  {"2"},
        "caref":             {"629e3f042ee41"},
        "keytype":           {"RSA"},
        "keylen":            {"2048"},
        "ecname":            {"prime256v1"},
        "digest_alg":        {"sha256"},
        "lifetime":          {"3650"},
        "authorizedkeys":    {""},
        "ipsecpsk":          {""},
        "act":               {""},
        "userid":            {""},
        "privid":            {""},
        "certid":            {""},
        "utype":             {"user"},
        "oldusername":       {""},
        "save":              {"Save"},
    }

    headers := map[string]string{
        "Content-Type": "application/x-www-form-urlencoded",
        "Referer":      BaseURL + UserManagerPath,
    }

    response, err := postFormWithHeaders(client, BaseURL+UserManagerPath, createUserPayload, headers)
    if err != nil {
        return fmt.Errorf("request to create user failed: %w", err)
    }
    defer response.Body.Close()

    return nil
}

func getCSRFToken(resp *http.Response) (string, error) {
    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return "", err
    }
    token, exists := doc.Find("input[name='__csrf_magic']").Attr("value")
    if !exists {
        return "", fmt.Errorf("CSRF token not found")
    }
    return token, nil
}

func postFormWithHeaders(client *http.Client, url string, formData url.Values, headers map[string]string) (*http.Response, error) {
    req, err := http.NewRequest("POST", url, strings.NewReader(formData.Encode()))
    if err != nil {
        return nil, err
    }
    for key, value := range headers {
        req.Header.Set(key, value)
    }
    return client.Do(req)
}

func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: message,
        },
    })
}

func getUserIDByUsername(s *discordgo.Session, guildID, username string) (string, error) {
    members, err := s.GuildMembersSearch(guildID, username, 1)
    if err != nil {
        return "", fmt.Errorf("failed to search for user %s: %w", username, err)
    }

    if len(members) == 0 {
        return "", fmt.Errorf("user %s not found in guild", username)
    }

    return members[0].User.ID, nil
}

func notifyUserOnDiscord(s *discordgo.Session, userID, username, password string) error {
    channel, err := s.UserChannelCreate(userID)
    if err != nil {
        return fmt.Errorf("failed to create DM channel for user ID %s: %w", userID, err)
    }

    message := fmt.Sprintf(
        "Hi, I am here to deliver your VPN credentials. Please import the attached profile to either OpenVPN or Pritunl. If your username indicates the profile is for a workshop (WS), you will lose access after the workshop ends. DM @chxuru on Discord if you have any questions.\n\n"+
            "Credentials\nUsername: ||%s||\nPassword: ||%s||\n\n"+
            "https://drive.google.com/file/d/12ncVQsFdZktoYm3BkBKam7g78uhTnG2r/view",
        username, password,
    )

    _, err = s.ChannelMessageSend(channel.ID, message)
    if err != nil {
        return fmt.Errorf("failed to send message to user ID %s: %w", userID, err)
    }

    return nil
}

func deleteUser(client *http.Client, ids []string) error {
    userManagerPage, err := client.Get(BaseURL + UserManagerPath)
    if err != nil {
        return fmt.Errorf("failed to fetch user manager page: %w", err)
    }
    defer userManagerPage.Body.Close()

    csrfToken, err := getCSRFToken(userManagerPage)
    if err != nil {
        return fmt.Errorf("failed to retrieve CSRF token: %w", err)
    }

    deleteUserPayload := url.Values{
        "__csrf_magic": {csrfToken},
        "dellall":      {"dellall"},
    }
    for _, id := range ids {
        deleteUserPayload.Add("delete_check[]", strings.TrimSpace(id))
    }

    headers := map[string]string{
        "Content-Type": "application/x-www-form-urlencoded",
        "Referer":      BaseURL + UserManagerPath,
    }

    response, err := postFormWithHeaders(client, BaseURL+"/system_usermanager.php", deleteUserPayload, headers)
    if err != nil {
        return fmt.Errorf("request to delete users failed: %w", err)
    }
    defer response.Body.Close()

    if response.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to delete users, status: %s", response.Status)
    }

    return nil
}

func terminateSession(client *http.Client, port, clientIPPort string) error {
    terminateURL := fmt.Sprintf("%s/status_openvpn.php?action=kill&port=%s&remipp=%s&client_id=", BaseURL, port, clientIPPort)
    fmt.Println("Terminate URL:", terminateURL)
    req, err := http.NewRequest("GET", terminateURL, nil)
    if err != nil {
        return fmt.Errorf("failed to create terminate request: %w", err)
    }

    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to execute terminate request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("terminate request failed with status: %s", resp.Status)
    }

    return nil
}
