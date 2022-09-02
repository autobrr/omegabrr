package processor

import (
	"fmt"
	"github.com/moistari/rls"
	"regexp"
	"strings"
)

func processTitle(title string) []string {
	// replace - : _
	if title == "" || title == " " {
		return nil
	}

	// cleans year like (2020) from arr title
	var re = regexp.MustCompile(`(?m)\s(\(\d+\))`)
	title = re.ReplaceAllString(title, "")

	t := NewTitleSlice()

	//titles := []string{}

	//titles = append(titles, rls.MustNormalize(title))
	t.Add(strings.ReplaceAll(title, " ", "?"))

	//fmt.Println(rls.MustClean(title))
	//fmt.Println(rls.MustNormalize(title))

	if strings.Contains(title, ". ") {
		t.Add(strings.ReplaceAll(title, ". ", "??"))

		strip := strings.ReplaceAll(title, ". ", " ")
		replace := strings.ReplaceAll(strip, " ", "?")
		t.Add(replace)
	}

	if strings.ContainsAny(title, "-") {
		strip := strings.ReplaceAll(title, "-", "?")
		replace := strings.ReplaceAll(strip, " ", "?")
		t.Add(replace)
	}
	if strings.ContainsAny(title, ":") {
		strip := strings.ReplaceAll(title, ":", "")
		replace := strings.ReplaceAll(strip, " ", "?")
		t.Add(replace)

		split := strings.Split(title, ":")
		if len(split) > 1 {
			first := strings.ReplaceAll(split[0], " ", "?")
			second := strings.ReplaceAll(strings.Trim(split[1], " "), " ", "?")
			part := fmt.Sprintf("%v*%v", first, second)

			t.Add(part)
		}

	}
	if strings.ContainsAny(title, "&") {
		t.Add(strings.ReplaceAll(title, " ", "?"))

		strip := strings.ReplaceAll(title, "&", "and")
		replace := strings.ReplaceAll(strip, " ", "?")
		t.Add(replace)
	}

	return t.Titles()
}

type Titles struct {
	t []string
}

func NewTitleSlice() *Titles {
	ts := Titles{
		t: []string{},
	}
	return &ts
}

func (ts *Titles) Add(title string) {
	ts.t = append(ts.t, fmt.Sprintf("*%v*", title))
}

func (ts *Titles) Titles() []string {
	return ts.t
}

type TitleSlice []string

func (s TitleSlice) Add(title string) {
	s = append(s, fmt.Sprintf("*%v*", title))
}

func processTitleShows(title string) []string {
	// replace - : _
	if title == "" || title == " " {
		return nil
	}

	// Example: She-hulk: Attorney at law

	var re = regexp.MustCompile(`(?m)\s(\(\d+\))`)

	title = re.ReplaceAllString(title, "")

	//var titles []string
	titles := []string{title}
	//if strings.ContainsAny(title, "-:&") {
	//
	//}
	titles = append(titles, rls.MustNormalize(title))
	titles = append(titles, strings.ReplaceAll(title, " ", "?"))

	//fmt.Println(rls.MustClean(title))
	fmt.Println(rls.MustNormalize(title))

	if strings.ContainsAny(title, "-") {
		strip := strings.ReplaceAll(title, "-", "?")
		replace := strings.ReplaceAll(strip, " ", "?")
		titles = append(titles, replace)
	}
	if strings.ContainsAny(title, ":") {
		//titles = append(titles, strings.ReplaceAll(title, ":", ""))

		strip := strings.ReplaceAll(title, ":", "")
		replace := strings.ReplaceAll(strip, " ", "?")
		titles = append(titles, replace)

		split := strings.Split(title, ":")
		if len(split) > 1 {
			first := strings.ReplaceAll(split[0], " ", "?")
			second := strings.ReplaceAll(strings.Trim(split[1], " "), " ", "?")
			part := fmt.Sprintf("%v*%v*", first, second)

			titles = append(titles, part)
		}

	}
	if strings.ContainsAny(title, "&") {
		titles = append(titles, strings.ReplaceAll(title, " ", "?"))

		strip := strings.ReplaceAll(title, "&", "and")
		titles = append(titles, strip)

		replace := strings.ReplaceAll(strip, " ", "?")
		titles = append(titles, replace)
	}

	//var alts []string
	//for _, t := range titles {
	//	alts = append(alts, strings.ReplaceAll(t, " ", "?"))
	//}

	return titles
}
