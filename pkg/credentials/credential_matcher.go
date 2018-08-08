package credentials

import (
	"encoding/json"
	"fmt"
	"github.com/onsi/gomega/types"
	"reflect"
)

type CredentialMatcher struct {
	expected   string
	checkField string
}

func (c CredentialMatcher) Match(actual interface{}) (bool, error) {
	actualCredentials, err := extractCredentialField(actual, c.checkField)
	if err != nil {
		return false, err
	}
	expectedCredentials, err := extractCredentialField(c.expected, c.checkField)
	if err != nil {
		return false, err
	}
	return reflect.DeepEqual(actualCredentials, expectedCredentials), nil
}

func (c CredentialMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Credentials do not match (field: %s)\nExpected: %v\nActual: %v", c.checkField, c.expected, actual)
}
func (c CredentialMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Credentials are unchanged (field: %s)\nActual: %v", c.checkField, actual)
}

func extractCredentialField(from interface{}, field string) (interface{}, error) {
	text := from.(string)
	var fromJson map[string]interface{}
	err := json.Unmarshal([]byte(text), &fromJson)
	if err != nil {
		return nil, err
	}
	credentials := fromJson["credentials"].(map[string]interface{})
	if field == "" {
		return credentials, nil
	}
	return credentials[field], nil
}

func haveTheSameCredentialsAs(expected string) types.GomegaMatcher {
	return CredentialMatcher{expected: expected}
}

func haveTheSameCredentialFieldAs(expected string, field string) types.GomegaMatcher {
	return CredentialMatcher{expected: expected, checkField: field}
}
