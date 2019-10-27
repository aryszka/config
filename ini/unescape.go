package ini

import "fmt"

// TODO:
// - common escape sequences

var (
	escapeChars        = []rune{'\\'}
	escapedSingleQuote = []rune{'\\', '\''}
	escapedDoubleQuote = []rune{'\\', '"'}
	escapedNonQuote    = []rune{'\\', '\'', '"', '\n', '=', '[', ']', '#'}
)

func errUnexpectedEscapeSequence(text string) error {
	return fmt.Errorf("unexpected escape sequence: %s", text)
}

func errUnexpectedQuoteSequence(quote string) error {
	return fmt.Errorf("unexpected quote sequence: %s", quote)
}

func charsContain(chars []rune, char rune) bool {
	for i := range chars {
		if chars[i] == char {
			return true
		}
	}

	return false
}

func unescape(escapeChars, escapedChars, text []rune) ([]rune, error) {
	var (
		result  []rune
		escaped bool
	)

	for _, c := range text {
		switch {
		case escaped:
			result = append(result, c)
			escaped = false
		case charsContain(escapeChars, c):
			escaped = true
		default:
			result = append(result, c)
		}
	}

	if escaped {
		return nil, errUnexpectedEscapeSequence(string(text))
	}

	return result, nil
}

func unquote(quote string) (string, error) {
	if len(quote) < 2 {
		return "", errUnexpectedQuoteSequence(quote)
	}

	chars := []rune(quote)
	switch {
	case chars[0] == '\'' && chars[len(chars)-1] == '\'':
		result, err := unescape(escapeChars, escapedSingleQuote, chars[1:len(chars)-1])
		return string(result), err
	case chars[0] == '"' && chars[len(chars)-1] == '"':
		result, err := unescape(escapeChars, escapedDoubleQuote, chars[1:len(chars)-1])
		return string(result), err
	default:
		return "", errUnexpectedQuoteSequence(quote)
	}
}

func unescapeNonQuote(text string) (string, error) {
	result, err := unescape(escapeChars, escapedNonQuote, []rune(text))
	return string(result), err
}
