package storage

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func ChapterDir(num string) string {
	// For plain numeric chapters ("78", "78.5"), keep the legacy ch-NNN format
	// so existing on-disk directories remain valid.
	f, err := strconv.ParseFloat(num, 64)
	if err == nil {
		whole := int(f)
		frac := f - math.Trunc(f)
		if frac == 0 {
			return fmt.Sprintf("ch-%03d", whole)
		}
		fracPart := strings.TrimPrefix(fmt.Sprintf("%.1f", frac), "0.")
		return fmt.Sprintf("ch-%03d-%s", whole, fracPart)
	}
	// Season-encoded or other non-numeric: slugify
	return Slugify(num)
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
