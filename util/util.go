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

// Uptime2seconds converts uptime pieces to seconds in float64 
func Uptime2seconds(w string, d string, h string, m string, s string) float64 {
	f64w := Str2float64(w)*604800
	if f64w == -1 { f64w = 0 }
	f64d := Str2float64(d)*86400
	if f64d == -1 { f64d = 0 }
	f64h := Str2float64(h)*3600
	if f64h == -1 { f64h = 0 }
	f64m := Str2float64(m)*60
	if f64m == -1 { f64m = 0 }
	f64s := Str2float64(s)
	if f64s == -1 { f64s = 0 }
	return f64d+f64h+f64m+f64s
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