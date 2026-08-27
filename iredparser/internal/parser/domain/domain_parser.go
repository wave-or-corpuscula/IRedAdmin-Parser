// Package domainparser parses domains
package domainparser

import (
	"bytes"
	"context"
	"fmt"
	"iredparser/internal/parser"
	"iredparser/internal/parser/client"
	apperrors "iredparser/pkg/errors"
	"iredparser/pkg/utils"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type DomainParser struct {
	client *client.Client
}

func NewDomainParser(client *client.Client) *DomainParser {
	return &DomainParser{client: client}
}

func (p *DomainParser) Parse(ctx context.Context, server string) (*parser.ParseDomainResult, error) {
	body, err := p.client.GetFromServer(ctx, server, parser.DomainsListPath)
	if err != nil {
		return nil, fmt.Errorf("domains parser: failed to parse domains: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("domains parser: failed to parse html body: %w", err)
	}

	var domains []*parser.Domain
	var parseErrors []error

	rows := doc.Find("tbody tr")
	rows.Each(func(_ int, row *goquery.Selection) {
		domain, err := parseRow(row)
		if err != nil {
			parseErrors = append(parseErrors, err)
			return
		}

		domains = append(domains, domain)
	})

	result := &parser.ParseDomainResult{
		Domains: domains,
		Total:   rows.Length(),
		Errors:  parseErrors,
	}

	return result, nil
}

func parseRow(row *goquery.Selection) (*parser.Domain, error) {
	disabled := row.HasClass("disabled")

	domain := strings.TrimSpace(row.Find("td").Eq(1).Text())
	if domain == "" {
		return nil, apperrors.ErrEmptyDomain
	}
	displayName := strings.TrimSpace(row.Find("td").Eq(2).Text())

	memoryField := strings.TrimSpace(row.Find("td").Eq(3).Text())
	usedQuota := strings.Split(memoryField, "/")
	if len(usedQuota) != 2 {
		return nil, apperrors.ErrInvalidQuotaFormat.Wrapf("%q, %s", memoryField, domain)
	}
	usedMemoryWithSuffix, quotaStr := strings.TrimSpace(usedQuota[0]), strings.TrimSpace(usedQuota[1])
	usedMemory, err := utils.GetMemoryBytes(usedMemoryWithSuffix)
	if err != nil {
		return nil, err
	}

	if usedMemory == -1 {
		return nil, apperrors.ErrInvalidMemoryValue.Wrap("used memory cannot be unlimited")
	}

	quota, err := utils.GetMemoryBytes(quotaStr)
	if err != nil {
		return nil, err
	}

	users := strings.TrimSpace(row.Find("td").Last().Text())
	usersAmount, err := strconv.Atoi(users)
	if err != nil || usersAmount < 0 {
		return nil, apperrors.ErrDomainParsing.Wrapf("invalid users amount: %q", users)
	}

	domainData := parser.Domain{
		Disabled:        disabled,
		Name:            domain,
		DisplayName:     displayName,
		QuotaBytes:      quota,
		UsedMemoryBytes: usedMemory,
		UsersAmount:     usersAmount,
	}

	return &domainData, nil
}
