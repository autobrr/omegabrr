package processor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_processTitle(t *testing.T) {
	type args struct {
		title        string
		matchRelease bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test_01",
			args: args{
				title:        "The Quick Brown Fox (2022)",
				matchRelease: false,
			},
			want: []string{"The?Quick?Brown?Fox"},
		},
		{
			name: "test_02",
			args: args{
				title:        "The Matrix     -        Reloaded (2929)",
				matchRelease: false,
			},
			want: []string{"The?Matrix*Reloaded"},
		},
		{
			name: "test_03",
			args: args{
				title:        "The Matrix -(Test)- Reloaded (2929)",
				matchRelease: false,
			},
			want: []string{"The?Matrix*Test*Reloaded"},
		},
		{
			name: "test_04",
			args: args{
				title:        "The Marvelous Mrs. Maisel",
				matchRelease: false,
			},
			want: []string{"The?Marvelous?Mrs*Maisel"},
		},
		{
			name: "test_05",
			args: args{
				title:        "Arrr!! The Title (2020)",
				matchRelease: false,
			},
			want: []string{"Arrr*The?Title"},
		},
		{
			name: "test_06",
			args: args{
				title:        "Whose Line Is It Anyway? (US)",
				matchRelease: false,
			},
			want: []string{"Whose?Line?Is?It?Anyway", "Whose?Line?Is?It?Anyway*US", "Whose?Line?Is?It?Anyway?"},
		},
		{
			name: "test_07",
			args: args{
				title:        "MasterChef (US)",
				matchRelease: false,
			},
			want: []string{"MasterChef*US", "MasterChef"},
		},
		{
			name: "test_08",
			args: args{
				title:        "Brooklyn Nine-Nine",
				matchRelease: false,
			},
			want: []string{"Brooklyn?Nine?Nine"},
		},
		{
			name: "test_09",
			args: args{
				title:        "S.W.A.T.",
				matchRelease: false,
			},
			want: []string{"S?W?A?T?", "S?W?A?T"},
		},
		{
			name: "test_10",
			args: args{
				title:        "The Handmaid's Tale",
				matchRelease: false,
			},
			want: []string{"The?Handmaid?s?Tale", "The?Handmaids?Tale"},
		},
		{
			name: "test_11",
			args: args{
				title:        "The Handmaid's Tale (US)",
				matchRelease: false,
			},
			want: []string{"The?Handmaid?s?Tale*US", "The?Handmaids?Tale*US", "The?Handmaid?s?Tale", "The?Handmaids?Tale"},
		},
		{
			name: "test_12",
			args: args{
				title:        "Monsters, Inc.",
				matchRelease: false,
			},
			want: []string{"Monsters*Inc?", "Monsters*Inc"},
		},
		{
			name: "test_13",
			args: args{
				title:        "Hello Tomorrow!",
				matchRelease: false,
			},
			want: []string{"Hello?Tomorrow?", "Hello?Tomorrow"},
		},
		{
			name: "test_14",
			args: args{
				title:        "Be Cool, Scooby-Doo!",
				matchRelease: false,
			},
			want: []string{"Be?Cool*Scooby?Doo?", "Be?Cool*Scooby?Doo"},
		},
		{
			name: "test_15",
			args: args{
				title:        "Scooby-Doo! Mystery Incorporated",
				matchRelease: false,
			},
			want: []string{"Scooby?Doo*Mystery?Incorporated"},
		},
		{
			name: "test_16",
			args: args{
				title:        "Master.Chef (US)",
				matchRelease: false,
			},
			want: []string{"Master?Chef*US", "Master?Chef"},
		},
		{
			name: "test_17",
			args: args{
				title:        "Whose Line Is It Anyway? (US)",
				matchRelease: false,
			},
			want: []string{"Whose?Line?Is?It?Anyway*US", "Whose?Line?Is?It?Anyway?", "Whose?Line?Is?It?Anyway"},
		},
		{
			name: "test_18",
			args: args{
				title:        "90 Day Fiancé: Pillow Talk",
				matchRelease: false,
			},
			want: []string{"90?Day?Fianc*Pillow?Talk"},
		},
		{
			name: "test_19",
			args: args{
				title:        "進撃の巨人",
				matchRelease: false,
			},
			want: []string{"進撃の巨人"},
		},
		{
			name: "test_20",
			args: args{
				title:        "呪術廻戦 0: 東京都立呪術高等専門学校",
				matchRelease: false,
			},
			want: []string{"呪術廻戦?0*東京都立呪術高等専門学校"},
		},
		{
			name: "test_21",
			args: args{
				title:        "-!",
				matchRelease: false,
			},
			want: []string{"-!"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ElementsMatch(t, tt.want, processTitle(tt.args.title, tt.args.matchRelease), tt.name)
		})
	}
}
