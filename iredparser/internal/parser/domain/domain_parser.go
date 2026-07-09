// Package domainparser parses domains
package domainparser

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"iredparser/internal/parser"
	"iredparser/internal/parser/client"
	"iredparser/pkg/utils"

	"github.com/PuerkitoBio/goquery"
)

type DomainParser struct {
	client *client.Client
}

func NewDomainParser(client *client.Client) *DomainParser {
	return &DomainParser{client: client}
}

func (p *DomainParser) Parse(ctx context.Context) ([]*parser.Domain, error) {
	body, err := p.client.GetFromBase(ctx, parser.DomainsListPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse domains: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse html body: %w", err)
	}

	var domains []*parser.Domain
	var parseErrors []error

	doc.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
		domain, err := parseRow(row)
		if err != nil {
			parseErrors = append(parseErrors, err)
			return
		}

		domains = append(domains, domain)
	})

	if len(parseErrors) > 0 {
		return nil, errors.Join(parseErrors...)
	}

	return domains, nil
}

func parseRow(row *goquery.Selection) (*parser.Domain, error) {
	domain := strings.TrimSpace(row.Find("td").Eq(1).Text())
	displayName := strings.TrimSpace(row.Find("td").Eq(2).Text())

	memoryField := strings.TrimSpace(row.Find("td").Eq(3).Text())
	usedQuota := strings.Split(memoryField, "/")
	if len(usedQuota) != 2 {
		return nil, fmt.Errorf("invalid quota format: %q, %s", memoryField, domain)
	}
	usedMemoryWithSuffix, quotaStr := strings.TrimSpace(usedQuota[0]), strings.TrimSpace(usedQuota[1])
	usedMemory, err := utils.GetMemoryBytes(usedMemoryWithSuffix)
	if err != nil {
		return nil, err
	}

	quota, err := utils.GetMemoryBytes(quotaStr)
	if err != nil {
		return nil, err
	}

	domainData := parser.Domain{
		Disabled:        true, // Fix сделать парсинг активности домена
		Name:            domain,
		DisplayName:     displayName,
		QuotaBytes:      quota,
		UsedMemoryBytes: usedMemory,
	}

	return &domainData, nil
}
