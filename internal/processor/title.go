package processor

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/autobrr/omegabrr/internal/domain"
)

func processTitle(title string, cfg *domain.ArrConfig) []string {
	// replace - : _
	if title == "" || title == " " {
		return nil
	}

	// cleans year like (2020) from arr title
	var re = regexp.MustCompile(`(?m)\s(\(\d+\))`)
	title = re.ReplaceAllString(title, "")

	t := NewTitleSlice()

	//titles = append(titles, rls.MustNormalize(title))
	t.Add(strings.ReplaceAll(title, " ", "?"), cfg)

	//fmt.Println(rls.MustClean(title))
	//fmt.Println(rls.MustNormalize(title))

	if strings.Contains(title, ". ") {
		t.Add(strings.ReplaceAll(title, ". ", "??"), cfg)

		strip := strings.ReplaceAll(title, ". ", " ")
		replace := strings.ReplaceAll(strip, " ", "?")
		t.Add(replace, cfg)
	}

	if strings.ContainsAny(title, "-") {
		strip := strings.ReplaceAll(title, "-", "?")
		replace := strings.ReplaceAll(strip, " ", "?")
		t.Add(replace, cfg)
	}

	if strings.ContainsAny(title, "!") {
		strip := strings.ReplaceAll(title, "!", "?")
		replace := strings.ReplaceAll(strip, " ", "?")
		t.Add(replace, cfg)
	}

	if strings.ContainsAny(title, ":") {
		strip := strings.ReplaceAll(title, ":", "")
		replace := strings.ReplaceAll(strip, " ", "?")
		t.Add(replace, cfg)

		split := strings.Split(title, ":")
		if len(split) > 1 {
			first := strings.ReplaceAll(split[0], " ", "?")
			second := strings.ReplaceAll(strings.Trim(split[1], " "), " ", "?")
			part := fmt.Sprintf("%v*%v", first, second)

			t.Add(part, cfg)
		}

	}

	if strings.ContainsAny(title, "&") {
		t.Add(strings.ReplaceAll(title, " ", "?"), cfg)

		strip := strings.ReplaceAll(title, "&", "and")
		replace := strings.ReplaceAll(strip, " ", "?")
		t.Add(replace, cfg)
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

func (ts *Titles) Add(title string, cfg *domain.ArrConfig) {
	if cfg.MatchRelease {
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
