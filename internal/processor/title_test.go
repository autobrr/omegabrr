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

	title = "The Matrix     -        Reloaded (2929)" // Handle hyphens with whitespace on each side
	expected = []string{"The?Matrix*Reloaded"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "The Matrix -(Test)- Reloaded (2929)" // Handle hyphens and parentheses with whitespace on each side
	expected = []string{"The?Matrix*Test*Reloaded"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "The Marvelous Mrs. Maisel" // Handle titles with ". "
	expected = []string{"The?Marvelous?Mrs*Maisel"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "Arrr!! The Title (2020)" // replace 2 or more special characters with a *
	expected = []string{"Arrr*The?Title"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "Whose Line Is It Anyway? (US)" // Handle parentheses and a question mark in the title
	expected = []string{"Whose?Line?Is?It?Anyway", "Whose?Line?Is?It?Anyway*US", "Whose?Line?Is?It?Anyway?"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "MasterChef (US)" // Handle parentheses without a question mark
	expected = []string{"MasterChef*US", "MasterChef"}
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
	expected = []string{"The?Handmaid?s?Tale", "The?Handmaids?Tale"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "Monsters, Inc." // Handle commas and special character ending
	expected = []string{"Monsters*Inc?", "Monsters*Inc"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "Hello Tomorrow!" // Handle commas and special character ending
	expected = []string{"Hello?Tomorrow", "Hello?Tomorrow?"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "Be Cool, Scooby-Doo!" // Handle multiple special characters
	expected = []string{"Be?Cool*Scooby?Doo?", "Be?Cool*Scooby?Doo"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "Scooby-Doo! Mystery Incorporated" // Handle multiple special characters
	expected = []string{"Scooby?Doo*Mystery?Incorporated"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "Master.Chef (US)" // with matchRelease set to true
	expected = []string{"*Master?Chef*US*", "*Master?Chef*"}
	result = processTitle(title, true)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "Whose Line Is It Anyway? (US)" // Handle parentheses and a question mark in the title with matchRelease enabled
	expected = []string{"*Whose?Line?Is?It?Anyway*", "*Whose?Line?Is?It?Anyway*US*", "*Whose?Line?Is?It?Anyway?*"}
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
