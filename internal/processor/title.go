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
	//t.Add(title, matchRelease)
	//fmt.Println(rls.MustClean(title))
	//fmt.Println(rls.MustNormalize(title))

	if strings.Contains(title, ". ") {
		t.Add(strings.ReplaceAll(title, ". ", "??"), matchRelease)

		strip := strings.ReplaceAll(title, ". ", " ")
		replace := strings.ReplaceAll(strip, " ", "?")
		t.Add(replace, matchRelease)
	}

	if strings.ContainsAny(title, "-:!") {
		replace := strings.ReplaceAll(title, " ", "?")
		if strings.ContainsAny(title, "-") {
			replace = strings.ReplaceAll(replace, "-", "?")
		}
		// Replaces exclamation marks with a single asterisk "*" unless its at the end of title
		if strings.ContainsAny(title, "!- ") {
			replace := title
			replace = strings.TrimRight(replace, "!")
			re := regexp.MustCompile(`!+(?:.)`)
			replace = re.ReplaceAllStringFunc(replace, func(match string) string {
				return "*"
			})
			replace = strings.ReplaceAll(replace, "-", "?")
			replace = strings.ReplaceAll(replace, "?*", "*")
			replace = strings.ReplaceAll(replace, "*?", "*")
			replace = strings.ReplaceAll(replace, " ", "?")
			replace = strings.Trim(replace, " ?")
			if strings.Contains(replace, "?*") {
				replace = strings.ReplaceAll(replace, "?*", "*")
				replace = strings.ReplaceAll(replace, "*?", "*")
			}
			t.Add(replace, matchRelease)
		}
		if strings.ContainsAny(title, ":") {
			replace = strings.ReplaceAll(replace, ":", "?")
			split := strings.SplitN(title, ":", 2)
			if len(split) > 1 {
				part := fmt.Sprintf("%v*%v", strings.ReplaceAll(split[0], " ", "?"), strings.ReplaceAll(strings.Trim(split[1], " "), " ", "?"))
				t.Add(part, matchRelease)
			}
		}
		t.Add(replace, matchRelease)
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

	//	// Strip special characters from title endings
	//	suffixes := []string{"!", ".", "?", ":", ";", "\"", "'", "~", "@", "#", "%", "^", "*", "=", "+", "(", "[", "]", "{", "}", "<", ">", "/", "?", "|", "\\", ",", " \t\r\n\f._"}
	//	for _, suffix := range suffixes {
	//		title = strings.TrimSuffix(title, suffix)
	//	}

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
