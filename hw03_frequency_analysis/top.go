package hw03frequencyanalysis

import (
	"regexp"
	"slices"
	"strings"
)

var pattern = regexp.MustCompile(`[[:punct:]]?([а-яА-Яa-zA-Z-0-9]+)[[:punct:]]?`)

type wordFrequency struct {
	word string
	freq int
}

func Top10(text string) []string {
	parts := strings.Fields(text)

	if len(parts) == 0 {
		return nil
	}

	words := make(map[string]int)
	for _, part := range parts {
		if part == "" || part == "-" {
			continue
		}
		match := pattern.FindStringSubmatch(part)
		if len(match) > 1 {
			part = match[1]
		}
		words[strings.ToLower(part)]++
	}

	frequencies := make([]*wordFrequency, 0, len(words))
	for word, count := range words {
		frequencies = append(frequencies, &wordFrequency{word: word, freq: count})
	}

	slices.SortFunc(frequencies, func(a, b *wordFrequency) int {
		if a.freq == b.freq {
			return strings.Compare(a.word, b.word)
		}
		return b.freq - a.freq
	})

	if len(frequencies) > 10 {
		frequencies = frequencies[:10]
	}
	out := make([]string, 0, len(frequencies))
	for _, wf := range frequencies {
		out = append(out, wf.word)
	}
	return out
}
