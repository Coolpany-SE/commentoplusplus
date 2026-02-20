package main

import (
	"testing"
)

func TestOwnerUpdateBasics(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	ownerHex, _ := ownerNew("test@example.com", "Test", "hunter2", false)

	if err := ownerUpdate(ownerHex, "updated@example.com", "UpdatedName", true); err != nil {
		t.Errorf("unexpected error updating owner: %v", err)
		return
	}

	o, err := ownerGetByOwnerHex(ownerHex)
	if err != nil {
		t.Errorf("unexpected error getting owner: %v", err)
		return
	}

	if o.Email != "updated@example.com" {
		t.Errorf("expected email=updated@example.com got email=%s", o.Email)
		return
	}

	if o.Name != "UpdatedName" {
		t.Errorf("expected name=UpdatedName got name=%s", o.Name)
		return
	}

	if o.ConfirmedEmail != true {
		t.Errorf("expected confirmedEmail=true got confirmedEmail=%v", o.ConfirmedEmail)
		return
	}
}

func TestOwnerUpdateNonExistent(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	if err := ownerUpdate("non-existent-hex", "test@example.com", "Test", false); err == nil {
		t.Errorf("expected error when updating non-existent owner")
		return
	}
}

func TestOwnerUpdateEmpty(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	ownerHex, _ := ownerNew("test@example.com", "Test", "hunter2", false)

	if err := ownerUpdate(ownerHex, "", "Test", false); err == nil {
		t.Errorf("expected error when passing empty email")
		return
	}

	if err := ownerUpdate(ownerHex, "test@example.com", "", false); err == nil {
		t.Errorf("expected error when passing empty name")
		return
	}

	if err := ownerUpdate("", "test@example.com", "Test", false); err == nil {
		t.Errorf("expected error when passing empty ownerHex")
		return
	}
}

func TestOwnerUpdateByAnotherOwner(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	ownerHex1, _ := ownerNew("test1@example.com", "Test1", "hunter2", false)
	ownerNew("test2@example.com", "Test2", "hunter2", false)

	// Owner2 updates Owner1's details
	if err := ownerUpdate(ownerHex1, "modified@example.com", "ModifiedName", true); err != nil {
		t.Errorf("unexpected error when another owner updates: %v", err)
		return
	}

	o, _ := ownerGetByOwnerHex(ownerHex1)
	if o.Email != "modified@example.com" {
		t.Errorf("expected email=modified@example.com got email=%s", o.Email)
		return
	}

	if o.Name != "ModifiedName" {
		t.Errorf("expected name=ModifiedName got name=%s", o.Name)
		return
	}
}
