package nlp

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/abadojack/whatlanggo"
	"github.com/kljensen/snowball"
)

func Lang(text ...string) string {
	whitelist := whatlanggo.Options{
		Whitelist: map[whatlanggo.Lang]bool{
			whatlanggo.Rus: true,
			whatlanggo.Eng: true,
		},
	}
	lang := whatlanggo.DetectLangWithOptions(strings.Join(text, " "), whitelist)
	return lang.Iso6391()
}

func Tokens(stem bool, text ...string) []string {
	lang := ""
	if stem {
		lang = Lang(text...)
	}
	words := bytes.FieldsFunc([]byte(strings.Join(text, " ")), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	for i := range words {
		words[i] = bytes.ToLower(words[i])
	}
	stemmed := make([]string, 0, len(words))
	switch lang {
	case "en":
		for i := range words {
			if stem, err := snowball.Stem(string(words[i]), "english", true); err == nil {
				stemmed = append(stemmed, stem)
			}
		}
	case "ru":
		for i := range words {

			if stem, err := snowball.Stem(string(words[i]), "russian", true); err == nil {
				stemmed = append(stemmed, stem)
			}
		}
	default:
		for i := range words {
			stemmed = append(stemmed, string(words[i]))
		}
	}

	return stemmed
}
