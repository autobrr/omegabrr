package processor

import (
	"testing"
)

func TestProcessTitle(t *testing.T) {
	title := "The Quick Brown Fox (2022)" // Removes year from title
	expected := []string{"The?Quick?Brown?Fox"}
	result := processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "The Matrix - Reloaded (2929)" // Handle hyphens with whitespace on each side
	expected = []string{"The?Matrix???Reloaded", "The?Matrix*Reloaded"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "The Marvelous Mrs. Maisel" // Handle titles with ". "
	expected = []string{"The?Marvelous?Mrs??Maisel", "The?Marvelous?Mrs?Maisel"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "Arrr!! The Title (2020)" // replace 2 or more special characters with a *
	expected = []string{"Arrr!!?The?Title", "Arrr*The?Title"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "Whose Line Is It Anyway? (US)" // Handle parentheses and a question mark
	expected = []string{"Whose?Line?Is?It?Anyway", "Whose?Line?Is?It?Anyway??US"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "Brooklyn Nine-Nine" // Handle hyphens
	expected = []string{"Brooklyn?Nine?Nine"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "S.W.A.T." // Titles ending with a dot
	expected = []string{"S?W?A?T?", "S?W?A?T"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "The Handmaid's Tale" // Titles with apostrophe
	expected = []string{"The?Handmaid's?Tale", "The?Handmaids?Tale"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "Scooby-Doo! Mystery Incorporated" // Handle multiple special characters
	expected = []string{"Scooby?Doo??Mystery?Incorporated", "Scooby?Doo!?Mystery?Incorporated"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "Master.Chef (US)" // with matchRelease set to true
	expected = []string{"*Master?Chef?US*", "*Master?Chef*"}
	result = processTitle(title, true)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}
}

// It returns true if the two slices have the same elements, regardless of order
func stringSlicesContainSameElements(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	m := make(map[string]int)
	for _, x := range a {
		m[x]++
	}
	for _, x := range b {
		if m[x] == 0 {
			return false
		}
		m[x]--
	}
	return true
}
