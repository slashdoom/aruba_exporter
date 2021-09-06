package util

import (
	"fmt"
	"regexp"
	"strings"
	"strconv"
)

// Str2float64 converts a string to float64
func Str2float64(str string) float64 {
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return -1
	}
	return value
}

func StandardizeMacAddr(str string) string {
	macRegex := regexp.MustCompile(`([0-9A-Fa-f]{2})[\.\-:]?([0-9A-Fa-f]{2})[\.\-:]?([0-9A-Fa-f]{2})[\.\-:]?([0-9A-Fa-f]{2})[\.\-:]?([0-9A-Fa-f]{2})[\.\-:]?([0-9A-Fa-f]{2})`)
	if matches := macRegex.FindStringSubmatch(str); matches != nil {
		mac := fmt.Sprintf("%s-%s-%s-%s-%s-%s", 
			matches[1],
			matches[2],
			matches[3],
			matches[4],
			matches[5],
			matches[6]) 
		return strings.ToUpper(mac)
	}
	return ""
}