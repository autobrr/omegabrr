package processor

import (
	"golift.io/starr"
)

func processTags(tags []*starr.Tag, titleTags []int, includeTags []string, excludeTags []string) bool {

	tagLabels := []string{}

	// match tag id's with labels
	for _, movieTag := range titleTags {
		for _, tag := range tags {
			if movieTag == tag.ID {
				tagLabels = append(tagLabels, tag.Label)
			}
		}
	}

	if len(includeTags) > 0 {
		for _, includeTag := range includeTags {
			for _, label := range tagLabels {
				if includeTag == label {
					return true
				}
			}
		}
	} else if len(excludeTags) > 0 {
		for _, excludeTag := range excludeTags {
			for _, label := range tagLabels {
				if excludeTag == label {
					return false
				}
			}
		}

	}

	return false
}
