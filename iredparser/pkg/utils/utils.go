// Package utils provides some utils for application
package utils

import (
	"bytes"
	"fmt"
	"io"
	"iredparser/pkg/errors"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetMemoryBytes(memWithSuffix string) (int64, error) {
	if memWithSuffix == "0" {
		return 0, nil
	}
	memorySuffix := []string{"Bytes", "KB", "MB", "GB"}
	suffixInd := -1
	for i, suff := range memorySuffix {
		if strings.HasSuffix(memWithSuffix, suff) {
			suffixInd = i
			break
		}
	}
	if suffixInd == -1 {
		if memWithSuffix == "Unlimited" {
			return -1, nil
		}
		log.Printf("unknown memory size suffix: %q\n", memWithSuffix)
		return -1, errors.ErrInvalidMemorySuffix
	}

	usedMemoryStr := strings.TrimSpace(strings.TrimSuffix(memWithSuffix, memorySuffix[suffixInd]))
	usedMemory, err := strconv.ParseFloat(usedMemoryStr, 64)
	if err != nil {
		log.Fatalf("invalid memory value: %q, %s\n", usedMemoryStr, err)
	}
	return int64(usedMemory * math.Pow(1000, float64(suffixInd))), nil
}

func GetLoginErrorMessage(body io.ReadCloser) error {
	content, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("utils: cannot read bytes from http-response: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("utils: cannot create document from http-response: %w", err)
	}

	p := doc.Find(".note-error").Find("p")
	errMessage := strings.TrimSpace(p.Clone().Find("strong").Remove().End().Text())

	switch {
	case strings.Contains(errMessage, "required"):
		return errors.ErrLoginRequired
	case strings.Contains(errMessage, "must be an valid email"):
		return errors.ErrInvalidUsername
	case strings.Contains(errMessage, "or password is incorrect"):
		return errors.ErrIncorrectCredentials
	}

	return nil
}
