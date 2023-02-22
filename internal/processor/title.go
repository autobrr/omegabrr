package processor

import (
	"fmt"
	"regexp"
	"strings"
)

func processTitle(title string, matchRelease bool) []string {
	// replace - : _
	if strings.TrimSpace(title) == "" {
		return nil
	}

	// cleans year like (2020) from arr title
	var re = regexp.MustCompile(`(?m)\s(\(\d+\))`)
	title = re.ReplaceAllString(title, "")

	t := NewTitleSlice()

	// Replace all occurrences of " ", ",", and "-" with "?"
	replaceAll := strings.NewReplacer(" ", "?", ",", "?", "-", "?")
	t.Add(replaceAll.Replace(title), matchRelease)

	// If title contains ". ", replace all occurrences with "??"
	if strings.Contains(title, ". ") {
		t.Add(strings.ReplaceAll(title, ". ", "??"), matchRelease)

		// Replace all occurrences of " " with "?"
		strip := strings.ReplaceAll(title, ". ", " ")
		replace := strings.ReplaceAll(strip, " ", "?")
		t.Add(replace, matchRelease)
	}

	if regexp.MustCompile(`[[:punct:]]`).MatchString(title) {
		// Regex patterns for matching "???" and all non-alphanumeric characters
		regexQuestionmark := regexp.MustCompile(`[?]{2,}`)
		regexReplace := regexp.MustCompile(`[^[:alnum:]]|\(.*?\)`)

		// Remove trailing ".", "!", and " " characters, and replace & with "and"
		title = strings.TrimRight(title, ".! ")
		title = strings.ReplaceAll(title, "'", "")
		title = strings.ReplaceAll(title, "’", "")
		title = strings.ReplaceAll(title, ",", "")
		title = strings.ReplaceAll(title, "&", "and")

		// Replace all non-alphanumeric characters with "?", and all occurrences of "???" with "*"
		replace := regexReplace.ReplaceAllString(title, "?")
		replace = regexQuestionmark.ReplaceAllString(replace, "*")
		t.Add(strings.TrimRight(replace, "?*"), matchRelease)
	}

	// If title ends with eg. (US), remove it completely. Trim any leftover whitespace at the end.
	parenRegexp := regexp.MustCompile(`\((.*?)\)`)
	matches := parenRegexp.FindAllStringSubmatch(title, -1)
	if len(matches) == 1 {
		parenContent := matches[0][0]
		title = strings.ReplaceAll(title, parenContent, "")

		// Trim any trailing whitespace from the modified title
		title = strings.TrimSpace(title)

		// Remove a trailing "?" if it exists
		if title[len(title)-1:] == "?" {
			title = strings.TrimRight(title, "?")
		}

		// Replace all spaces with "?" and add the modified title to the slice
		replace := strings.ReplaceAll(title, " ", "?")
		t.Add(replace, matchRelease)
	}

	return t.Titles()
}

type Titles struct {
	tm map[string]struct{}
}

func NewTitleSlice() *Titles {
	ts := Titles{
		tm: map[string]struct{}{},
	}
	return &ts
}

func (ts *Titles) Add(title string, matchRelease bool) {
	if matchRelease {
		title = fmt.Sprintf("*%v*", title)
	}

	_, ok := ts.tm[title]
	if !ok {
		ts.tm[title] = struct{}{}
	}
}

func (ts *Titles) Titles() []string {
	titles := []string{}
	for key := range ts.tm {
		titles = append(titles, key)
	}
	return titles
}
