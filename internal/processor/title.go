package processor

import (
	"fmt"
	"regexp"
	"strings"
)

func processTitle(title string, matchRelease bool) []string {
	// replace - : _
	if title == "" || title == " " {
		return nil
	}

	// cleans year like (2020) from arr title
	var re = regexp.MustCompile(`(?m)\s(\(\d+\))`)
	title = re.ReplaceAllString(title, "")

	t := NewTitleSlice()

	//titles = append(titles, rls.MustNormalize(title))
	t.Add(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(title, " ", "?"), ",", ""), "-", "?"), matchRelease)

	if strings.Contains(title, ". ") {
		t.Add(strings.ReplaceAll(title, ". ", "??"), matchRelease)

		strip := strings.ReplaceAll(title, ". ", " ")
		replace := strings.ReplaceAll(strip, " ", "?")
		t.Add(replace, matchRelease)
	}

	if re, err := regexp.Compile(`\s\(([A-Za-z]{2})\)$`); err == nil {
		regex := regexp.MustCompile(`[?]{2,}`)
		if re.MatchString(title) {
			replace := re.ReplaceAllString(title, "*$1")
			replace = strings.ReplaceAll(replace, " ", "?")
			replace = regex.ReplaceAllString(replace, "*")
			t.Add(replace, matchRelease)
		}
	}

	if strings.ContainsAny(title, "!?-,:.&'") {
		regex := regexp.MustCompile(`[?]{2,}`)
		title = strings.TrimRight(title, "!")
		title = strings.TrimRight(title, ".")
		title = strings.ReplaceAll(title, "!", "?")
		title = strings.ReplaceAll(title, ".", "?")
		title = strings.ReplaceAll(title, "'", "")
		title = strings.ReplaceAll(title, "-", "?")
		title = strings.ReplaceAll(title, ":", "?")
		title = strings.ReplaceAll(title, ",", "")
		title = strings.ReplaceAll(title, "?*", "*")
		title = strings.ReplaceAll(title, "&", "and")
		title = strings.ReplaceAll(title, " ", "?")
		replace := strings.ReplaceAll(title, " ", "?")
		replace = regex.ReplaceAllString(replace, "*")
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
