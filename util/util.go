package util

import (
	"fmt"
	"regexp"
	"strings"
	"strconv"
)

// Str2float64 converts a string to float64
func Str2float64(s string) float64 {
	ns := strings.Replace(s, ",", "", -1)
	value, err := strconv.ParseFloat(ns, 64)
	if err != nil {
		return -1
	}
	return value
}

// StandardizeMacAddr converts MAC address into IEEE 802 format (e.g. 01-23-45-67-89-AB)
func StandardizeMacAddr(s string) string {
	r := regexp.MustCompile(`([0-9A-Fa-f]{2})[\.\-:]?([0-9A-Fa-f]{2})[\.\-:]?([0-9A-Fa-f]{2})[\.\-:]?([0-9A-Fa-f]{2})[\.\-:]?([0-9A-Fa-f]{2})[\.\-:]?([0-9A-Fa-f]{2})`)
	if matches := r.FindStringSubmatch(s); matches != nil {
		m := fmt.Sprintf("%s-%s-%s-%s-%s-%s", 
			matches[1],
			matches[2],
			matches[3],
			matches[4],
			matches[5],
			matches[6]) 
		return strings.ToUpper(m)
	}
	return ""
}