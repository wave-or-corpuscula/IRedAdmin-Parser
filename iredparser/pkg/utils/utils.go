// Package utils provides some utils for application
package utils

import (
	"log"
	"math"
	"strconv"
	"strings"

	"iredparser/pkg/errors"
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
