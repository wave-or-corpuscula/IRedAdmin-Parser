// Package mailboxparser parses mailboxes (async)
package mailboxparser

import (
	"bytes"
	"context"
	"fmt"
	"iredparser/internal/parser"
	"iredparser/internal/parser/client"
	apperrors "iredparser/pkg/errors"
	"iredparser/pkg/utils"
	"log"
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

	pagesSpan := doc.Find(".pages")
	spans := pagesSpan.Find("a")
	usersAmountStr := spans.Last().Text()
	pages, err := strconv.Atoi(usersAmountStr)
	if err != nil {
		return -1, apperrors.ErrInvalidPageValue
	}

	if pages <= 0 {
		return -1, apperrors.ErrInvalidPageValue.Wrap("pages amount cannt be negative or 0")
	}

	return pages, nil
}

func (p *MailboxParser) Parse(ctx context.Context, server string, domain parser.Domain) (*parser.ParseMailboxesResult, error) {
	pages, err := p.getPagesAmount(ctx, server, domain)
	if err != nil {
		return nil, err
	}

	return p.parsePages(ctx, server, domain, pages)
}

func (p *MailboxParser) parsePages(ctx context.Context, server string, domain parser.Domain, pages int) (*parser.ParseMailboxesResult, error) {
	jobs := make(chan string)
	resultsCh := make(chan *parser.ParseMailboxesResult)

	var wg sync.WaitGroup
	var workerErrors []error

	for i := 0; i < p.workers; i++ {
		wg.Go(func() {
			for pageURL := range jobs {
				result, err := p.parsePage(
					ctx,
					pageURL,
				)
				if err != nil {
					workerErrors = append(workerErrors, err)
					return
				}

				resultsCh <- result
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
		close(resultsCh)
	}()

	results := &parser.ParseMailboxesResult{}

	for res := range resultsCh {
		results.Extend(res)
	}

	if len(workerErrors) != 0 {
		return nil, apperrors.NewMultiError(workerErrors)
	}

	return results, nil
}

func (p *MailboxParser) parsePage(ctx context.Context, pageURL string) (*parser.ParseMailboxesResult, error) {
	body, err := p.client.Get(ctx, pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get mailboxes page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse mailboxes page: %w", err)
	}

	var mailboxes []*parser.Mailbox
	var parseErrors []error
	rows := doc.Find("tbody tr")

	rows.Each(func(i int, row *goquery.Selection) {
		mailbox, err := p.parsePageMailboxes(row)
		if err != nil {
			log.Println("error in", pageURL, "in position:", i)
			parseErrors = append(parseErrors, err)
			return
		}

		mailboxes = append(mailboxes, mailbox)
	})

	result := &parser.ParseMailboxesResult{
		Mailboxes: mailboxes,
		Total:     rows.Length(),
		Errors:    parseErrors,
	}

	return result, nil
}

func (p *MailboxParser) parsePageMailboxes(row *goquery.Selection) (*parser.Mailbox, error) {
	displayName := strings.TrimSpace(row.Find("td").Eq(1).Text())
	mailAddress := strings.TrimSpace(row.Find("td").Eq(0).Find("input[name='mail']").AttrOr("value", ""))

	if len(mailAddress) == 0 {
		return nil, apperrors.ErrEmptyMailAddress
	}

	quotaField := strings.TrimSpace(row.Find("td").Eq(5).Find(".color-grey a").Text())
	if len(quotaField) == 0 {
		quotaField = strings.TrimSpace(row.Find("td").Eq(5).Text())
	}

	disabled := row.Find("td").Eq(1).Find(".color-red").Size() > 0
	isAdmin := row.Find("td").Eq(1).Find(".color-blue").Size() > 0

	usedQuota := strings.Split(quotaField, "/")
	if len(usedQuota) != 2 {
		return nil, apperrors.ErrInvalidQuotaFormat.Wrapf("%q, %s", quotaField, mailAddress)
	}
	usedMemoryWithSuffix, quotaWithSuffix := strings.TrimSpace(usedQuota[0]), strings.TrimSpace(usedQuota[1])

	usedMemory, err := utils.GetMemoryBytes(usedMemoryWithSuffix)
	if err != nil {
		return nil, apperrors.ErrInvalidMemoryValue.Wrapf("%q", usedMemoryWithSuffix)
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
