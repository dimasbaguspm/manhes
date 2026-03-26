package storage

import (
	"fmt"
	"math"
	"strings"
)

func ChapterDir(num float64) string {
	whole := int(num)
	frac := num - math.Trunc(num)
	if frac == 0 {
		return fmt.Sprintf("ch-%03d", whole)
	}
	fracPart := strings.TrimPrefix(fmt.Sprintf("%.1f", frac), "0.")
	return fmt.Sprintf("ch-%03d-%s", whole, fracPart)
}

func PageFile(idx int, ext string) string {
	return fmt.Sprintf("%03d%s", idx, ext)
}

func Slugify(title string) string {
	s := strings.ToLower(title)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, s)
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}
