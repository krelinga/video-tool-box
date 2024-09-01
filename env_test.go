package main

import (
	"os"
	"testing"
)

const (
	testEnvKey   = "VTB_TEST_KEY"
	testEnvValue = "VTB_TEST_VALUE"
)

func TestGetEnvVarEmptyValue(t *testing.T) {
	if err := os.Setenv(testEnvKey, ""); err != nil {
		t.Fatalf("error setting env var: %s", err)
	}

	value, err := getEnvVar(testEnvKey)
	if err == nil {
		t.Error("expected an error")
	}
	if value != "" {
		t.Errorf("expected empty value, got %s", value)
	}
}

func TestGetEnvVarSetValue(t *testing.T) {
	if err := os.Setenv(testEnvKey, testEnvValue); err != nil {
		t.Fatalf("error setting env var: %s", err)
	}

	value, err := getEnvVar(testEnvKey)
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
	if value != testEnvValue {
		t.Errorf("expected %s, got %s", testEnvValue, value)
	}
}
