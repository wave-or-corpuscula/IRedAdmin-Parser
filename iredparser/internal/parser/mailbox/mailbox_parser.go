// Package mailboxparser parses mailboxes (async)
package mailboxparser

import (
	"bytes"
	"context"
	"fmt"
	"iredparser/internal/parser"
	"iredparser/internal/parser/client"
	"iredparser/pkg/utils"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type MailboxParser struct {
	client  *client.Client
	workers int
}

func NewMailboxParser(client *client.Client, workers int) *MailboxParser {
	return &MailboxParser{client: client, workers: workers}
}

func (p *MailboxParser) getPagesAmount(ctx context.Context, server string, domain parser.Domain) (int, error) {
	body, err := p.client.GetFromServer(ctx, server, parser.DomainUsersPath+domain.Name)
	if err != nil {
		return -1, fmt.Errorf("failed to get domain page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return -1, fmt.Errorf("error while parsing pages html: %w", err)
	}

	pages := 0
	doc.Find(".pages").Each(func(i int, selection *goquery.Selection) {
		spans := selection.Find("a")
		usersAmountStr := spans.Last().Text()
		pages, _ = strconv.Atoi(usersAmountStr)
	})

	return pages, nil
}

func (p *MailboxParser) Parse(ctx context.Context, server string, domain parser.Domain) ([]*parser.Mailbox, error) {
	pages, err := p.getPagesAmount(ctx, server, domain)
	if err != nil {
		return nil, err
	}

	return p.parsePages(ctx, server, domain, pages)
}

func (p *MailboxParser) parsePages(ctx context.Context, server string, domain parser.Domain, pages int) ([]*parser.Mailbox, error) {
	jobs := make(chan string)
	results := make(chan []*parser.Mailbox)

	var wg sync.WaitGroup

	for i := 0; i < p.workers; i++ {
		wg.Go(func() {
			for pageURL := range jobs {
				boxes, err := p.parsePage(
					ctx,
					pageURL,
				)
				if err != nil {
					continue // TODO: Logging parsing errors
				}

				results <- boxes
			}
		})
	}

	baseURL := parser.CreateBaseURL(server)

	go func() {
		defer close(jobs)
		for page := range pages {
			jobs <- fmt.Sprintf("%s%s%s%s%d", baseURL, parser.DomainUsersPath, domain.Name, parser.DomainUsersPagesPath, page+1)
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var mailboxes []*parser.Mailbox

	for boxes := range results {
		mailboxes = append(mailboxes, boxes...)
	}

	return mailboxes, nil
}

func (p *MailboxParser) parsePage(ctx context.Context, pageURL string) ([]*parser.Mailbox, error) {
	body, err := p.client.Get(ctx, pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get mailboxes page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse mailboxes pabe: %w", err)
	}

	var parseErrors []error
	var mailboxes []*parser.Mailbox

	doc.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
		mailbox, err := p.parsePageMailboxes(row)
		if err != nil {
			parseErrors = append(parseErrors, err)
			return
		}

		mailboxes = append(mailboxes, mailbox)
	})

	return mailboxes, nil
}

func (p *MailboxParser) parsePageMailboxes(row *goquery.Selection) (*parser.Mailbox, error) {
	displayName := strings.TrimSpace(row.Find("td").Eq(1).Text())
	mailAddress := strings.TrimSpace(row.Find("td").Eq(2).Text())

	quotaField := strings.TrimSpace(row.Find("td").Eq(5).Find(".color-grey a").Text())
	if len(quotaField) == 0 {
		quotaField = strings.TrimSpace(row.Find("td").Eq(5).Text())
	}

	disabled := row.Find("td").Eq(1).Find(".color-red").Size() > 0
	isAdmin := row.Find("td").Eq(1).Find(".color-blue").Size() > 0

	usedQuota := strings.Split(quotaField, "/")
	if len(usedQuota) != 2 {
		return nil, fmt.Errorf("invalid quota field: %q, %s", quotaField, mailAddress)
	}
	usedMemoryWithSuffix, quotaWithSuffix := strings.TrimSpace(usedQuota[0]), strings.TrimSpace(usedQuota[1])

	usedMemory, err := utils.GetMemoryBytes(usedMemoryWithSuffix)
	if err != nil {
		return nil, fmt.Errorf("invalid used memory value: %s", usedMemoryWithSuffix)
	}

	quota, err := utils.GetMemoryBytes(quotaWithSuffix)
	if err != nil {
		quota = -1
	}

	return &parser.Mailbox{
		Disabled:        disabled,
		IsAdmin:         isAdmin,
		DisplayName:     displayName,
		Address:         mailAddress,
		QuotaBytes:      quota,
		UsedMemoryBytes: usedMemory,
	}, nil
}
