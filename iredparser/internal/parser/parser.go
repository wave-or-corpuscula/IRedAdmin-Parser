// Package parser provides parser for all data about domains and mailboxes in provided server
package parser

import (
	"fmt"
)

const (
	DomainsListPath      = "/domains"
	profilePath          = "/profile/user/general/"
	LoginPath            = "/login"
	DomainUsersPath      = "/users/"
	DomainUsersPagesPath = "/page/"
)

var HTTPTimeoutSeconds = 20

// type ServerConfig struct {
// 	ServerName string `json:"server_name"`
// 	Login      string `json:"login"`
// 	Password   string `json:"password"`
// }

func CreateBaseURL(serverName string) string {
	return fmt.Sprintf("https://%s/iredadmin", serverName)
}
