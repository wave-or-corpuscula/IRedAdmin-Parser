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
	PasswordPath         = "/profile/user/password/"
)

var HTTPTimeoutSeconds = 20

func CreateBaseURL(serverName string) string {
	return fmt.Sprintf("https://%s/iredadmin", serverName)
}

func CreatePasswordPath(serverName string) string {
	return fmt.Sprintf("%s%s", CreateBaseURL(serverName), PasswordPath)
}

func CreateChangePasswordPath(serverName string, mailbox string) string {
	return fmt.Sprintf("%s%s%s", CreateBaseURL(serverName), PasswordPath, mailbox)
}
