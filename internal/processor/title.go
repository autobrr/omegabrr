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
	t.Add(strings.ReplaceAll(title, " ", "?"), matchRelease)

	//fmt.Println(rls.MustClean(title))
	//fmt.Println(rls.MustNormalize(title))

	// Replaces whitespace with "?" and adds title and alternative title without ending "?".
	if strings.HasSuffix(title, "?") {
		// create version without the question mark
		altTitle := strings.TrimSuffix(title, "?")
		// add both versions
		t.Add(strings.ReplaceAll(title, " ", "?"), matchRelease)
		t.Add(strings.ReplaceAll(altTitle, " ", "?"), matchRelease)
	} else {
		t.Add(strings.ReplaceAll(title, " ", "?"), matchRelease)
	}

	if strings.Contains(title, ". ") {
		t.Add(strings.ReplaceAll(title, ". ", "??"), matchRelease)

		strip := strings.ReplaceAll(title, ". ", " ")
		replace := strings.ReplaceAll(strip, " ", "?")
		t.Add(replace, matchRelease)
	}

	if strings.ContainsAny(title, "-") {
		strip := strings.ReplaceAll(title, "-", "?")
		replace := strings.ReplaceAll(strip, " ", "?")
		replace = strings.ReplaceAll(replace, "!", "?")
		replace = strings.ReplaceAll(replace, "!?", "*")
		replace = strings.ReplaceAll(replace, "??", "*")
		t.Add(replace, matchRelease)
	}

	// Replaces all exclamation marks "!" in a title with a question mark "?" except for the last ! if the title ends with a "!"
	// Then it replaces any consecutive "??" with a single "*" and any consecutive "**" with a single "*".
	if strings.ContainsAny(title, "!") {
		last := title[len(title)-1:]
		if last == "!" {
			title = strings.TrimRight(title, "!")
		} else {
			title = strings.ReplaceAll(title, "!", "?")
		}
		replace := strings.ReplaceAll(title, " ", "?")
		replace = strings.ReplaceAll(replace, "-", "?")
		replace = strings.ReplaceAll(replace, "?*", "*")
		replace = strings.ReplaceAll(replace, "*?", "*")
		replace = strings.ReplaceAll(replace, "??", "*")
		replace = strings.ReplaceAll(replace, "**", "*")
		t.Add(replace, matchRelease)
	}

	if strings.ContainsAny(title, ":") {
		strip := strings.ReplaceAll(title, ":", "")
		replace := strings.ReplaceAll(strip, " ", "?")
		t.Add(replace, matchRelease)

		split := strings.Split(title, ":")
		if len(split) > 1 {
			first := strings.ReplaceAll(split[0], " ", "?")
			second := strings.ReplaceAll(strings.Trim(split[1], " "), " ", "?")
			part := fmt.Sprintf("%v*%v", first, second)

			t.Add(part, matchRelease)
		}

	}

	if strings.ContainsAny(title, "&") {
		t.Add(strings.ReplaceAll(title, " ", "?"), matchRelease)

		strip := strings.ReplaceAll(title, "&", "and")
		replace := strings.ReplaceAll(strip, " ", "?")
		t.Add(replace, matchRelease)
	}

	// Replaces "." characters in a title with "?" except when the title ends with ".", in which case the "." is removed.
	if strings.ContainsAny(title, ".") {
		strip := strings.TrimRight(title, ".")
		replace := strings.ReplaceAll(strip, ".", "?")
		replace = strings.ReplaceAll(replace, " ", "?")
		t.Add(replace, matchRelease)
	}

	// Strips titles ending with parentheses like (US) to *US, and removes a consecutive question mark if present
	if re, err := regexp.Compile(`\s\(([A-Za-z]{2})\)$`); err == nil {
		if re.MatchString(title) {
			replace := re.ReplaceAllString(title, "*$1")
			replace = strings.Replace(replace, "*?", "*", 1)
			replace = strings.Replace(replace, "?*", "*", 1)
			replace = strings.ReplaceAll(replace, " ", "?")
			replace = strings.ReplaceAll(replace, "* ", "*")
			replace = strings.ReplaceAll(replace, " *", "*")
			t.Add(replace, matchRelease)
		}
	}

	// Strip apostrophes from titles
	if strings.ContainsAny(title, "'") {
		strip := strings.ReplaceAll(title, "'", "")
		replace := strings.ReplaceAll(strip, " ", "?")
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
