package main

import (
	"errors"
	"iredparser/internal/controller"
	"iredparser/internal/database"
	"iredparser/internal/parser/client"
	"log"
	"os"

	domainparser "iredparser/internal/parser/domain"
	mailboxparser "iredparser/internal/parser/mailbox"
	authservice "iredparser/internal/services/auth_service"
	passwordservice "iredparser/internal/services/password_service"
	syncservice "iredparser/internal/sync"
	syncdomain "iredparser/internal/sync/domain"
	syncmailbox "iredparser/internal/sync/mailbox"
	apperrors "iredparser/pkg/errors"
)

func main() {
	httpClient, err := client.NewClient()
	if err != nil {
		log.Fatalln(err)
	}

	authService := authservice.NewAuthService(httpClient)

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

	passwordService := passwordservice.NewPasswordService(httpClient)

	ctrl := controller.NewCLIController(httpClient, db, authService, syncService, passwordService, os.Stdout)

	rootCmd := ctrl.InitCommands()

	var clierr *apperrors.IRedError

	if err := controller.Execute(rootCmd); err != nil {
		if errors.As(err, &clierr) {
			ctrl.SendIRedError(clierr)
		} else {
			ctrl.SendError(
				string(apperrors.ErrTypeCLI),
				int(apperrors.ErrCodeCLIUnknown),
				err,
			)
		}
	}
}
