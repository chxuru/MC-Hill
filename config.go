package main

import (
    "log"
    "os"
)

var (
    DiscordBotToken   = os.Getenv("DISCORD_BOT_TOKEN")
    BaseURL           = os.Getenv("PFSENSE_BASE_URL")
    LoginPath         = "index.php"
    UserManagerPath   = "system_usermanager.php"
    OpenvpnStatusPath = "status_openvpn.php"
    AdminUsername     = os.Getenv("PFSENSE_ADMIN_USERNAME")
    AdminPassword     = os.Getenv("PFSENSE_ADMIN_PASSWORD")
    GuildID           = os.Getenv("DISCORD_GUILD_ID")
)

func init() {
    requiredEnvVars := []string{"DISCORD_BOT_TOKEN", "PFSENSE_BASE_URL", "PFSENSE_ADMIN_USERNAME", "PFSENSE_ADMIN_PASSWORD", "DISCORD_GUILD_ID"}
    for _, envVar := range requiredEnvVars {
        if os.Getenv(envVar) == "" {
            log.Fatalf("Environment variable %s is not set", envVar)
        }
    }
}
