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

    LDAPBindDN        = os.Getenv("LDAP_BIND_DN")
    LDAPURL           = os.Getenv("LDAP_URL")
    LDAPBindPassword  = os.Getenv("LDAP_BIND_PASSWORD")
    LDAPBaseDN        = os.Getenv("LDAP_BASE_DN")
    LDAPGroupDN       = os.Getenv("LDAP_GROUP_DN")
    LDAPUsersDN       = os.Getenv("LDAP_USERS_DN")
    LDAPUserAttribute = os.Getenv("LDAP_USER_ATTRIBUTE")
    LDAPInsecureTLS   = os.Getenv("LDAP_INSECURE_TLS") == "true"
)

var guildIDs = []string{
    "185912981576482816",
    "954324859876417557",
    "1300343195720486932",
    "514863658698801163",
}

func init() {
    requiredEnvVars := []string{"DISCORD_BOT_TOKEN", "PFSENSE_BASE_URL", "PFSENSE_ADMIN_USERNAME", "PFSENSE_ADMIN_PASSWORD", "LDAP_BIND_DN", "LDAP_URL", "LDAP_BIND_PASSWORD", "LDAP_BASE_DN", "LDAP_GROUP_DN", "LDAP_USERS_DN", "LDAP_USER_ATTRIBUTE"}
    for _, envVar := range requiredEnvVars {
        if os.Getenv(envVar) == "" {
            log.Fatalf("Environment variable %s is not set", envVar)
        }
    }
}
