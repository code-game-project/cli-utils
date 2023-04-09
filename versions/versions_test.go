package versions

import (
	"errors"
	"fmt"
	"testing"
)

func Test_IsCompatible(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want bool
	}{
		{"1.2.3", "2.2.3", false},
		{"2.2.3", "1.2.3", false},
		{"0.2.3", "0.3.3", false},
		{"0.2.3", "0.1.3", false},
		{"1.2.3", "1.1.3", true},
		{"1.2.3", "1.3.3", false},
		{"0.2", "0.2.5", true},
		{"1.2.3", "1.2.4", true},
		{"0.2.3", "0.2.4", true},
		{"1.2.3", "1.2.3", true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s - %s", tt.a, tt.b), func(t *testing.T) {
			a := MustParse(tt.a)
			b := MustParse(tt.b)
			got := a.IsCompatible(b)
			if got != tt.want {
				t.Errorf("IsCompatible = %t, want %t", got, tt.want)
			}
		})
	}
}

func Test_Compare(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want int
	}{
		{"1.2.3", "2.2.3", 1},
		{"2.2.3", "1.2.3", -1},
		{"0.2.3", "0.3.3", 1},
		{"0.2.3", "0.1.3", -1},
		{"1.2.3", "1.1.3", -1},
		{"1.2.3", "1.3.3", 1},
		{"0.2", "0.2.5", 1},
		{"1.2.3", "1.2.4", 1},
		{"0.2.3", "0.2.4", 1},
		{"1.2.3", "1.2.3", 0},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s - %s", tt.a, tt.b), func(t *testing.T) {
			a := MustParse(tt.a)
			b := MustParse(tt.b)
			got := Compare(a, b)
			if got != tt.want {
				t.Errorf("Compare = %d, want %d", got, tt.want)
			}
		})
	}
}

func Test_FindCompatibleInMap(t *testing.T) {
	tests := []struct {
		name      string
		versions  map[string]string
		cgVersion string
		want      string
		wantErr   error
	}{
		{name: "exact match", versions: map[string]string{
			"0.1": "1.2",
			"0.3": "1.4",
			"0.5": "1.6",
			"0.7": "1.8",
			"1.1": "2.2",
			"1.5": "2.6",
			"1.7": "2.8",
		}, cgVersion: "0.5", want: "1.6"},
		{name: "next lowest minor", versions: map[string]string{
			"1.1": "1.2",
			"1.3": "1.4",
			"1.5": "2.6",
			"1.7": "2.8",
		}, cgVersion: "1.6", want: "2.6"},
		{name: "minor too low (major = 0)", versions: map[string]string{
			"0.1": "1.2",
			"0.3": "1.4",
			"1.1": "2.2",
			"1.5": "2.6",
			"1.7": "2.8",
		}, cgVersion: "0.5", wantErr: ErrNoCompatibleVersion},
		{name: "minor too high", versions: map[string]string{
			"0.1": "1.2",
			"0.3": "1.4",
			"0.6": "1.6",
			"0.7": "1.8",
			"1.1": "2.2",
			"1.5": "2.6",
			"1.7": "2.8",
		}, cgVersion: "0.5", wantErr: ErrNoCompatibleVersion},
		{name: "major mismatch", versions: map[string]string{
			"0.1": "1.2",
			"0.3": "1.4",
			"1.1": "2.2",
			"1.5": "2.6",
			"1.7": "2.8",
		}, cgVersion: "2.5", wantErr: ErrNoCompatibleVersion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cgVersion := MustParse(tt.cgVersion)
			got, err := FindCompatibleInMap(cgVersion, tt.versions)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("IsCompatible err '%v', want err '%v'", err, tt.wantErr)
			}
			if tt.wantErr == nil && got.String() != tt.want {
				t.Errorf("IsCompatible = %v, want %v", got.String(), tt.want)
			}
		})
	}
}
