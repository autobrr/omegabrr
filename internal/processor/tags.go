package processor

import (
	"golift.io/starr"
)

func processTags(tags []*starr.Tag, titleTags []int, includeTags []string, excludeTags []string) bool {

	if len(includeTags) == 0 && len(excludeTags) == 0 {
		return true
	}

	tagLabels := []string{}
	ret := false

	// match tag id's with labels
	for _, movieTag := range titleTags {
		for _, tag := range tags {
			if movieTag == tag.ID {
				tagLabels = append(tagLabels, tag.Label)
			}
		}
	}

	// check included tags and set ret to true if we have a match
	for _, includeTag := range includeTags {
		for _, label := range tagLabels {
			if includeTag == label {
				ret = true
				continue
			}
		}
	}

	// check exclude tags and return false on first match
	for _, excludeTag := range excludeTags {
		for _, label := range tagLabels {
			if excludeTag == label {
				return false
			}
		}
	}

	return ret
}
