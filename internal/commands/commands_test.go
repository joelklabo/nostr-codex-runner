package commands

import "testing"

func TestParseCommands(t *testing.T) {
	cases := []struct {
		in   string
		name string
		args string
	}{
		{"/new start fresh", "new", "start fresh"},
		{"/use abc123", "use", "abc123"},
		{"/status", "status", ""},
		{"/help", "help", ""},
		{"/shell ls -la", "shell", "ls -la"},
		{"shell ls -la", "shell", "ls -la"},
		{"free text prompt", "run", "free text prompt"},
	}
	for _, tc := range cases {
		cmd := Parse(tc.in)
		if cmd.Name != tc.name {
			t.Fatalf("%q expected name %s got %s", tc.in, tc.name, cmd.Name)
		}
		if cmd.Args != tc.args {
			t.Fatalf("%q expected args %q got %q", tc.in, tc.args, cmd.Args)
		}
	}
}
