package check

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestEnvChecker(t *testing.T) {
	const key = "CHECK_TEST_ENV"
	_ = os.Unsetenv(key)
	c := EnvChecker{}
	res := c.Check(DepInput{Name: key, Type: "env"})
	if res.Status != "MISSING" {
		t.Fatalf("expected missing when unset, got %s", res.Status)
	}
	_ = os.Setenv(key, "ok")
	defer func() { _ = os.Unsetenv(key) }()
	res = c.Check(DepInput{Name: key, Type: "env"})
	if res.Status != "OK" {
		t.Fatalf("expected OK when set, got %s", res.Status)
	}
}

func TestFileChecker(t *testing.T) {
	c := FileChecker{}
	res := c.Check(DepInput{Name: filepath.Join("this", "does", "not", "exist"), Type: "file"})
	if res.Status != "MISSING" {
		t.Fatalf("expected missing for absent file, got %s", res.Status)
	}
	res = c.Check(DepInput{Name: "check.go", Type: "file"})
	if res.Status != "OK" {
		t.Fatalf("expected OK for existing file, got %s", res.Status)
	}
}

func TestURLChecker(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer s.Close()

	c := URLChecker{}
	res := c.Check(DepInput{Name: s.URL, Type: "url"})
	if res.Status != "OK" {
		t.Fatalf("expected OK for reachable url, got %s", res.Status)
	}

	res = c.Check(DepInput{Name: "http://127.0.0.1:1", Type: "url"})
	if res.Status != "MISSING" {
		t.Fatalf("expected MISSING for bad url, got %s", res.Status)
	}
}

func TestPortChecker(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	defer func() { _ = ln.Close() }()

	c := PortChecker{}
	res := c.Check(DepInput{Name: addr, Type: "port"})
	if res.Status != "OK" {
		t.Fatalf("expected OK for open port, got %s", res.Status)
	}

	res = c.Check(DepInput{Name: "127.0.0.1:9", Type: "port"})
	if res.Status != "MISSING" {
		t.Fatalf("expected MISSING for closed port, got %s", res.Status)
	}
}

func TestRelayChecker(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	defer func() { _ = ln.Close() }()

	c := RelayChecker{}
	res := c.Check(DepInput{Name: "wss://" + addr, Type: "relay"})
	if res.Status != "OK" {
		t.Fatalf("expected OK for reachable relay, got %s", res.Status)
	}
	res = c.Check(DepInput{Name: "127.0.0.1:9", Type: "relay"})
	if res.Status != "MISSING" {
		t.Fatalf("expected MISSING for unreachable relay, got %s", res.Status)
	}
}

func TestDirWriteChecker(t *testing.T) {
	td := t.TempDir()
	c := DirWriteChecker{}
	res := c.Check(DepInput{Name: td, Type: "dirwrite"})
	if res.Status != "OK" {
		t.Fatalf("expected OK for writable temp dir, got %s", res.Status)
	}
	res = c.Check(DepInput{Name: "/nonexistent-path-hopefully", Type: "dirwrite"})
	if res.Status != "MISSING" {
		t.Fatalf("expected MISSING for nonexistent dir, got %s", res.Status)
	}
}
