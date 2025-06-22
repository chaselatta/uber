package main

import (
	"io"
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *RunContext
		wantErr bool
	}{
		{
			name: "all flags and command",
			args: []string{"--root", "/tmp", "--verbose", "start", "foo", "bar"},
			want: &RunContext{
				Root:          "/tmp",
				Verbose:       true,
				Command:       "start",
				RemainingArgs: []string{"foo", "bar"},
			},
			wantErr: false,
		},
		{
			name:    "empty",
			args:    []string{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "root not parsed twice",
			args: []string{"--root", "/tmp", "--verbose", "start", "--root", "foo"},
			want: &RunContext{
				Root:          "/tmp",
				Verbose:       true,
				Command:       "start",
				RemainingArgs: []string{"--root", "foo"},
			},
			wantErr: false,
		},
		{
			name: "short verbose flag",
			args: []string{"-v", "stop"},
			want: &RunContext{
				Root:          "",
				Verbose:       true,
				Command:       "stop",
				RemainingArgs: []string{},
			},
			wantErr: false,
		},
		{
			name:    "missing command",
			args:    []string{"--root", "/tmp"},
			want:    nil,
			wantErr: true,
		},
		{
			name: "no flags, just command",
			args: []string{"start"},
			want: &RunContext{
				Root:          "",
				Verbose:       false,
				Command:       "start",
				RemainingArgs: []string{},
			},
			wantErr: false,
		},
		{
			name:    "unknown flag",
			args:    []string{"--unknown", "start"},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArgs(tt.args, io.Discard)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseArgs() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
