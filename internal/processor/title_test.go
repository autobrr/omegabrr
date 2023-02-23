package processor

import (
	"testing"
)

func TestProcessTitle(t *testing.T) {
	title := "The Quick Brown Fox (2022)" // removes year from title
	expected := []string{"The?Quick?Brown?Fox"}
	result := processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "The Matrix - Reloaded (2929)"
	expected = []string{"The?Matrix?-?Reloaded", "The?Matrix*Reloaded"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "Arrr!!! The Title (2020)"
	expected = []string{"Arrr!!!?The?Title", "Arrr*The?Title"}
	result = processTitle(title, false)
	if !stringSlicesContainSameElements(result, expected) {
		t.Errorf("processTitle(%q, %t) = %v, expected %v", title, true, result, expected)
	}

	title = "Master.Chef (US)"
	expected = []string{"Master?Chef?US", "Master?Chef"}
	result = processTitle(title, false)
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
