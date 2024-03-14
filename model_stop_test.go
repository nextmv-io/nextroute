// © 2019-present nextmv.io inc

package nextroute_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/nextmv-io/nextroute"
	"github.com/nextmv-io/nextroute/common"
)

func TestModelStop_ToEarliestStart(t *testing.T) {
	model, err := nextroute.NewModel()
	if err != nil {
		t.Fatal(err)
	}

	s1, err := model.NewStop(common.NewInvalidLocation())
	if err != nil {
		t.Fatal(err)
	}

	windows := [][2]float64{
		{60, 120},
		{240, 360},
		{360, 420},
		{480, 540},
	}
	windowsAsTime := make([][2]time.Time, len(windows))
	for i, w := range windows {
		windowsAsTime[i] = [2]time.Time{
			model.Epoch().Add(time.Duration(w[0]) * model.DurationUnit()),
			model.Epoch().Add(time.Duration(w[1]) * model.DurationUnit()),
		}
	}
	err = s1.SetWindows(windowsAsTime)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		t    float64
		want float64
	}{
		{t: 0, want: 60},
		{t: 120, want: 240},
		{t: 130, want: 240},
		{t: 240, want: 240},
		{t: 300, want: 300},
		{t: 480, want: 480},
		{t: 539, want: 539},
		{t: 540, want: 540},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v -> %v", tt.t, tt.want), func(_ *testing.T) {
			earliest := s1.ToEarliestStartValue(tt.t)
			if earliest != tt.want {
				t.Errorf("got %v, want %v, at %v", earliest, tt.want, tt.t)
			}
		})
	}
}
