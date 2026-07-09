package main

import (
	"log"
	"os"

	"iredparser/internal/controller"
	"iredparser/internal/database"
	"iredparser/internal/parser/client"
	domainparser "iredparser/internal/parser/domain"
	mailboxparser "iredparser/internal/parser/mailbox"
	authservice "iredparser/internal/services/auth_service"
	syncservice "iredparser/internal/sync"
	syncdomain "iredparser/internal/sync/domain"
	syncmailbox "iredparser/internal/sync/mailbox"
)

func main() {
	httpClient := client.NewClient("")
	authService := authservice.NewAuthService()

	db, err := database.Connect(controller.DSN)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	mailParser := mailboxparser.NewMailboxParser(httpClient, controller.Workers)
	domainParser := domainparser.NewDomainParser(httpClient)

	mailboxSyncService := syncmailbox.NewMailboxSyncService(mailParser, db)
	domainSyncService := syncdomain.NewDomainSyncService(domainParser, db)

	syncService := syncservice.NewSyncService(mailboxSyncService, domainSyncService)

	ctrl := controller.NewCLIController(httpClient, db, authService, syncService, os.Stdout)

	rootCmd := ctrl.InitCommands()

	if err := controller.Execute(rootCmd); err != nil {
		ctrl.SendError(controller.ErrCli, err)
	}
}
