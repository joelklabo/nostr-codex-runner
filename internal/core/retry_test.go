package core

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetrySuccessFirst(t *testing.T) {
	err := retry(context.Background(), 3, func() error { return nil })
	if err != nil {
		t.Fatalf("unexpected err %v", err)
	}
}

func TestRetryExhaust(t *testing.T) {
	err := retry(context.Background(), 2, func() error { return errors.New("fail") })
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestRetryCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := retry(ctx, 5, func() error { return errors.New("fail") })
	if err == nil {
		t.Fatalf("expected error on cancel")
	}
}
