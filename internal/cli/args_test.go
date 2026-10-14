package cli

import (
	"flag"
	"io"
	"reflect"
	"testing"
)

func newTestFlagSet() (*flag.FlagSet, *string, *bool) {
	f := flag.NewFlagSet("test", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	box := f.String("box", "", "")
	force := f.Bool("force", false, "")
	return f, box, force
}

func TestPositional(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		want      []string
		wantBox   string
		wantForce bool
	}{
		{"No arguments.", nil, nil, "", false},
		{"Flags before the arguments.", []string{"-box", "b", "a1", "a2"}, []string{"a1", "a2"}, "b", false},
		{"Flags after the arguments.", []string{"a1", "-box", "b", "a2", "-force"}, []string{"a1", "a2"}, "b", true},
		{"Only arguments.", []string{"a1"}, []string{"a1"}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, box, force := newTestFlagSet()
			if err := f.Parse(tt.args); err != nil {
				t.Fatalf("unexpected err occurred: %v", err)
			}

			got, err := positional(f)
			if err != nil {
				t.Fatalf("unexpected err occurred: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("args = %q, want %q", got, tt.want)
			}
			if *box != tt.wantBox || *force != tt.wantForce {
				t.Errorf("box = %q, force = %t, want %q, %t", *box, *force, tt.wantBox, tt.wantForce)
			}
		})
	}

	t.Run("Fail on an unknown flag after the arguments.", func(t *testing.T) {
		f, _, _ := newTestFlagSet()
		if err := f.Parse([]string{"a1", "-nope"}); err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if _, err := positional(f); err == nil {
			t.Fatalf("error was expected, but not occured.")
		}
	})
}
