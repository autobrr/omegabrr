package processor

import (
	"testing"

	"golift.io/starr"
)

func Test_containsTag(t *testing.T) {
	type args struct {
		tags      []*starr.Tag
		titleTags []int
		checkTags []string
	}

	tags := []*starr.Tag{
		{
			ID:    1,
			Label: "Want",
		},
		{
			ID:    2,
			Label: "exclude-me",
		},
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test_1",
			args: args{
				tags:      tags,
				titleTags: []int{},
				checkTags: []string{"Want"},
			},
			want: false,
		},
		{
			name: "test_2",
			args: args{
				tags:      tags,
				titleTags: []int{1},
				checkTags: []string{"Want"},
			},
			want: true,
		},
		{
			name: "test_3",
			args: args{
				tags:      tags,
				titleTags: []int{1},
				checkTags: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsTag(tt.args.tags, tt.args.titleTags, tt.args.checkTags); got != tt.want {
				t.Errorf("containsTag() = %v, want %v", got, tt.want)
			}
		})
	}
}
