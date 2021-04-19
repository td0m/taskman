package date

import (
	"reflect"
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	today := time.Now().Truncate(time.Hour * 24)
	tests := []struct {
		name    string
		args    string
		want    RepeatableDate
		wantErr bool
	}{
		{"Case Insensitive", "Today", RepeatableDate{Type: Once, Value: today}, false},
		// TODO: Add more test cases
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDate(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
