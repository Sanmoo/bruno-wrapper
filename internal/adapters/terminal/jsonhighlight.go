package terminal

import (
	"strings"

	"github.com/muesli/termenv"
)

func highlightJSON(raw string) string {
	if termenv.EnvNoColor() {
		return raw
	}

	output := termenv.DefaultOutput()
	keyColor := termenv.ANSIBlue
	stringColor := termenv.ANSIGreen
	numberColor := termenv.ANSIWhite
	boolColor := termenv.ANSIWhite
	nullColor := termenv.ANSIBrightBlack
	structColor := termenv.ANSIWhite

	var result strings.Builder
	i := 0
	for i < len(raw) {
		c := raw[i]

		switch c {
		case ' ', '\t', '\n', '\r':
			result.WriteByte(c)
			i++
			continue
		case '{', '}', '[', ']', ':', ',':
			result.WriteString(output.String(string(c)).Foreground(structColor).Bold().String())
			i++
			continue
		case '"':
			start := i
			i++
			for i < len(raw) {
				if raw[i] == '\\' {
					i += 2
					continue
				}
				if raw[i] == '"' {
					i++
					break
				}
				i++
			}
			token := raw[start:i]
			content := token[1 : len(token)-1]

			isKey := false
			j := i
			for j < len(raw) && (raw[j] == ' ' || raw[j] == '\t') {
				j++
			}
			if j < len(raw) && raw[j] == ':' {
				isKey = true
			}

			var style termenv.Style
			if isKey {
				style = output.String(token).Foreground(keyColor).Bold()
			} else if content == "true" || content == "false" || content == "null" {
				style = output.String(token).Foreground(nullColor)
			} else if isNumber(content) {
				style = output.String(token).Foreground(numberColor)
			} else {
				style = output.String(token).Foreground(stringColor)
			}
			result.WriteString(style.String())
			continue
		default:
			start := i
			for i < len(raw) && raw[i] != ' ' && raw[i] != '\t' && raw[i] != '\n' && raw[i] != '\r' && raw[i] != ',' && raw[i] != '}' && raw[i] != ']' {
				i++
			}
			token := raw[start:i]
			if token == "" {
				continue
			}

			var style termenv.Style
			if token == "true" || token == "false" {
				style = output.String(token).Foreground(boolColor)
			} else if token == "null" {
				style = output.String(token).Foreground(nullColor)
			} else {
				style = output.String(token).Foreground(numberColor)
			}
			result.WriteString(style.String())
		}
	}

	return result.String()
}

func isNumber(s string) bool {
	if len(s) == 0 {
		return false
	}
	if s[0] == '-' {
		s = s[1:]
	}
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c >= '0' && c <= '9' {
			continue
		}
		if c == '.' || c == 'e' || c == 'E' || c == '+' || c == '-' {
			continue
		}
		return false
	}
	return true
}
