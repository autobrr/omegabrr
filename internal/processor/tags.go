package processor

import (
	"golift.io/starr"
)

func containsTag(tags []*starr.Tag, titleTags []int, checkTags []string) bool {
	tagLabels := []string{}

	// match tag id's with labels
	for _, movieTag := range titleTags {
		for _, tag := range tags {
			tag := tag
			if movieTag == tag.ID {
				tagLabels = append(tagLabels, tag.Label)
			}
		}
	}

	// log the tagLabels and checkTags
	//log.Printf("tagLabels: %v, checkTags: %v", tagLabels, checkTags)

	// check included tags and set ret to true if we have a match
	for _, includeTag := range checkTags {
		for _, label := range tagLabels {
			if includeTag == label {
				return true
			}
		}
	}

	return false
}
