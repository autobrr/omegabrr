package processor

import (
	"fmt"
	"regexp"
	"strings"
)

func processTitle(title string, matchRelease bool) []string {
	// Checking if the title is empty.
	if strings.TrimSpace(title) == "" {
		return nil
	}

	// cleans year like (2020) from arr title
	var re = regexp.MustCompile(`(?m)\s(\(\d+\))`)
	title = re.ReplaceAllString(title, "")

	t := NewTitleSlice()

	// Regex patterns
	replaceRegexp := regexp.MustCompile(`[^[:alnum:]]`)
	replaceNotApostropheRegexp := regexp.MustCompile(`[^[:alnum:]']`)
	questionmarkRegexp := regexp.MustCompile(`[?]{2,}`)
	regionCodeRegexp := regexp.MustCompile(`\(.+\)$`)
	parenthesesEndRegexp := regexp.MustCompile(`\)$`)

	// title with all non-alphanumeric characters replaced by "?"
	apostropheTitle := parenthesesEndRegexp.ReplaceAllString(title, "")
	apostropheTitle = replaceRegexp.ReplaceAllString(apostropheTitle, "?")
	apostropheTitle = questionmarkRegexp.ReplaceAllString(apostropheTitle, "*")

	t.Add(apostropheTitle, matchRelease)
	t.Add(strings.TrimRight(apostropheTitle, "?* "), matchRelease)

	// title with apostrophes removed and all non-alphanumeric characters replaced by "?"
	noApostropheTitle := parenthesesEndRegexp.ReplaceAllString(title, "")
	noApostropheTitle = replaceNotApostropheRegexp.ReplaceAllString(noApostropheTitle, "?")
	noApostropheTitle = strings.ReplaceAll(noApostropheTitle, "'", "")
	noApostropheTitle = questionmarkRegexp.ReplaceAllString(noApostropheTitle, "*")

	t.Add(noApostropheTitle, matchRelease)
	t.Add(strings.TrimRight(noApostropheTitle, "?* "), matchRelease)

	// title with regions in parentheses removed and all non-alphanumeric characters replaced by "?"
	removedRegionCode := regionCodeRegexp.ReplaceAllString(title, "")
	removedRegionCode = strings.TrimRight(removedRegionCode, " ")
	removedRegionCode = replaceRegexp.ReplaceAllString(removedRegionCode, "?")
	removedRegionCode = questionmarkRegexp.ReplaceAllString(removedRegionCode, "*")

	t.Add(removedRegionCode, matchRelease)
	t.Add(strings.TrimRight(removedRegionCode, "?* "), matchRelease)

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
