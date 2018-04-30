package mail

import (
	"fmt"
	"testing"

	gomail "gopkg.in/gomail.v2"
)

var (
	testMsg  = "hello"
	testSubj = "subj"
	testRec  = "someone"

	testHost = "host:22"
	testUser = "user"
	testPass = "pass"
	testFrom = "fromAddress"
)

func TestInit(t *testing.T) {
	err := Send(testRec, testSubj, testMsg)
	if err != nil {
		t.Errorf("Send should've returned without error, instead got: %v", err)
	}

	err = Init(testHost, testUser, testPass, testFrom)
	if err != nil {
		t.Errorf("Init should've succeeded, failed with: %v", err)
	}

	err = checkVars()
	if err != nil {
		t.Errorf("incorrect variables: %v", err)
	}

	err = Init(testHost, testUser, testPass, testFrom)
	if err == nil {
		t.Errorf("Init should have failed the second time, didn't")
	}
}

func TestInitNoAuth(t *testing.T) {
	// De-initialize
	initialized = false
	fromAddr = ""
	dialer = gomail.Dialer{}

	err := InitNoAuth(testHost, testFrom)
	if err != nil {
		t.Errorf("InitNoAuth should've succeeded, failed with: %v", err)
	}

	testUser = ""
	testPass = ""

	err = checkVars()
	if err != nil {
		t.Errorf("incorrect variables: %v", err)
	}

	err = InitNoAuth(testHost, testFrom)
	if err == nil {
		t.Errorf("Init should have failed the second time, didn't")
	}
}

func checkVars() error {
	if !initialized {
		return fmt.Errorf("Initialized should be true")
	}

	if fromAddr != testFrom {
		return fmt.Errorf("fromAddr mismatch, should be %q, got %q", testFrom, fromAddr)
	}

	resHost := fmt.Sprintf("%s:%d", dialer.Host, dialer.Port)

	if resHost != testHost {
		return fmt.Errorf("Host mismatch: should be %q, got %q", testHost, fmt.Sprintf("%s:%d", dialer.Host, dialer.Port))
	}

	return nil
}
