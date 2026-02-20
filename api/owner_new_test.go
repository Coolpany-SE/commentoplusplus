package main

import (
	"testing"
)

func TestOwnerNewBasics(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	if _, err := ownerNew("test@example.com", "Test", "hunter2", false); err != nil {
		t.Errorf("unexpected error when creating new owner: %v", err)
		return
	}
}

func TestOwnerNewClash(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	if _, err := ownerNew("test@example.com", "Test", "hunter2", false); err != nil {
		t.Errorf("unexpected error when creating new owner: %v", err)
		return
	}

	if _, err := ownerNew("test@example.com", "Test", "hunter2", false); err == nil {
		t.Errorf("expected error not found when creating with clashing email")
		return
	}
}

func TestOwnerNewEmpty(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	if _, err := ownerNew("test@example.com", "", "hunter2", false); err == nil {
		t.Errorf("expected error not found when passing empty name")
		return
	}

	if _, err := ownerNew("", "", "", false); err == nil {
		t.Errorf("expected error not found when passing empty everything")
		return
	}
}
