package processor

import (
	"testing"

	"golift.io/starr"
)

func Test_processTags(t *testing.T) {
	type args struct {
		tags        []*starr.Tag
		movieTags   []int
		includeTags []string
		excludeTags []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test_1",
			args: args{
				tags: []*starr.Tag{{
					ID:    1,
					Label: "Want",
				}},
				movieTags:   []int{1},
				includeTags: []string{"Want"},
				excludeTags: nil,
			},
			want: true,
		},
		{
			name: "test_2",
			args: args{
				tags: []*starr.Tag{
					{
						ID:    1,
						Label: "Want",
					},
					{
						ID:    2,
						Label: "exclude-me",
					},
				},
				movieTags:   []int{1, 2},
				includeTags: nil,
				excludeTags: []string{"exclude-me"},
			},
			want: false,
		},
		{
			name: "test_3",
			args: args{
				tags: []*starr.Tag{
					{
						ID:    1,
						Label: "Want",
					},
					{
						ID:    2,
						Label: "exclude-me",
					},
				},
				movieTags:   []int{},
				includeTags: []string{"Want"},
			},
			want: false,
		},
		{
			name: "test_4",
			args: args{
				tags: []*starr.Tag{{
					ID:    1,
					Label: "Want",
				}},
				movieTags:   []int{},
				includeTags: nil,
				excludeTags: nil,
			},
			want: true,
		},
		{
			name: "test_5",
			args: args{
				tags: []*starr.Tag{
					{
						ID:    1,
						Label: "Want",
					},
					{
						ID:    2,
						Label: "exclude-me",
					},
				},
				movieTags:   []int{1, 2},
				includeTags: []string{"Want"},
				excludeTags: []string{"exclude-me"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := processTags(tt.args.tags, tt.args.movieTags, tt.args.includeTags, tt.args.excludeTags); got != tt.want {
				t.Errorf("processTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
