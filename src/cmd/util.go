package cmd

import "strings"

func splitShell(s string) []string {
	var args []string
	var cur strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for _, r := range s {
		switch {
		case (r == '"' || r == '\'') && !inQuotes:
			inQuotes = true
			quoteChar = r
		case r == quoteChar && inQuotes:
			inQuotes = false
		case r == ' ' && !inQuotes:
			if cur.Len() > 0 {
				args = append(args, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}

	if cur.Len() > 0 {
		args = append(args, cur.String())
	}
	return args
}
