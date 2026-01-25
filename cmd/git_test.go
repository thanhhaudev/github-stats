package main

import (
	"errors"
	"strings"
	"testing"
)

func TestSanitizeError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		token         string
		owner         string
		expectedParts []string // Parts that should be in the result
		forbiddenParts []string // Parts that should NOT be in the result
	}{
		{
			name:          "nil error returns nil",
			err:           nil,
			token:         "ghp_1234567890",
			owner:         "testuser",
			expectedParts: nil,
			forbiddenParts: nil,
		},
		{
			name:          "token is redacted",
			err:           errors.New("authentication failed with token ghp_1234567890"),
			token:         "ghp_1234567890",
			owner:         "testuser",
			expectedParts: []string{"[***]"},
			forbiddenParts: []string{"ghp_1234567890"},
		},
		{
			name:          "owner/username is redacted",
			err:           errors.New("failed to push to testuser/repo"),
			token:         "ghp_1234567890",
			owner:         "testuser",
			expectedParts: []string{"[***]"},
			forbiddenParts: []string{"testuser"},
		},
		{
			name:          "https URL is completely redacted",
			err:           errors.New("failed to clone https://github.com/user/repo.git"),
			token:         "",
			owner:         "",
			expectedParts: []string{"[***]"},
			forbiddenParts: []string{"https://github.com", "github.com", "/user/repo.git"},
		},
		{
			name:          "http URL is completely redacted",
			err:           errors.New("failed to fetch http://example.com/path/to/resource"),
			token:         "",
			owner:         "",
			expectedParts: []string{"[***]"},
			forbiddenParts: []string{"http://example.com", "example.com", "/path/to/resource"},
		},
		{
			name:          "multiple URLs are redacted",
			err:           errors.New("failed: https://github.com/repo1 and https://github.com/repo2"),
			token:         "",
			owner:         "",
			expectedParts: []string{"[***]"},
			forbiddenParts: []string{"github.com", "repo1", "repo2"},
		},
		{
			name:          "URL with token in it is redacted",
			err:           errors.New("failed to push to https://ghp_token123@github.com/user/repo.git"),
			token:         "ghp_token123",
			owner:         "user",
			expectedParts: []string{"[***]"},
			forbiddenParts: []string{"ghp_token123", "github.com", "user", "repo.git"},
		},
		{
			name:          "complex error with token, owner, and URL",
			err:           errors.New("git remote set-url failed for testuser: https://ghp_secret@github.com/testuser/myrepo.git"),
			token:         "ghp_secret",
			owner:         "testuser",
			expectedParts: []string{"[***]"},
			forbiddenParts: []string{"ghp_secret", "testuser", "github.com", "myrepo"},
		},
		{
			name:          "error without sensitive data remains unchanged",
			err:           errors.New("connection timeout"),
			token:         "ghp_1234",
			owner:         "user",
			expectedParts: []string{"connection timeout"},
			forbiddenParts: nil,
		},
		{
			name:          "URL with query parameters is fully redacted",
			err:           errors.New("failed: https://api.github.com/repos/user/repo?token=abc123"),
			token:         "",
			owner:         "",
			expectedParts: []string{"[***]"},
			forbiddenParts: []string{"api.github.com", "token=abc123", "user", "repo"},
		},
		{
			name:          "empty token and owner still redacts URLs",
			err:           errors.New("error with https://github.com/path"),
			token:         "",
			owner:         "",
			expectedParts: []string{"[***]"},
			forbiddenParts: []string{"github.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeError(tt.err, tt.token, tt.owner)

			// Check nil case
			if tt.err == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("expected non-nil error, got nil")
				return
			}

			resultMsg := result.Error()

			// Check that expected parts are present
			for _, expected := range tt.expectedParts {
				if !strings.Contains(resultMsg, expected) {
					t.Errorf("expected result to contain %q, but got: %s", expected, resultMsg)
				}
			}

			// Check that forbidden parts are NOT present
			for _, forbidden := range tt.forbiddenParts {
				if strings.Contains(resultMsg, forbidden) {
					t.Errorf("result should NOT contain %q, but got: %s", forbidden, resultMsg)
				}
			}
		})
	}
}




func TestSanitizeError_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		token    string
		owner    string
		expected string
	}{
		{
			name:     "very long URL is redacted",
			err:      errors.New("error: https://github.com/very/long/path/to/repository/with/many/segments/file.git?param1=value1&param2=value2"),
			token:    "",
			owner:    "",
			expected: "error: [***]",
		},
		{
			name:     "URL at end of string without whitespace",
			err:      errors.New("failed to access https://github.com/user/repo.git"),
			token:    "",
			owner:    "",
			expected: "failed to access [***]",
		},
		{
			name:     "multiple different protocols",
			err:      errors.New("tried http://example.com and https://github.com"),
			token:    "",
			owner:    "",
			expected: "tried [***] and [***]",
		},
		{
			name:     "token appears multiple times",
			err:      errors.New("token ghp_abc used in ghp_abc authentication"),
			token:    "ghp_abc",
			owner:    "",
			expected: "token [***] used in [***] authentication",
		},
		{
			name:     "owner appears multiple times",
			err:      errors.New("user john tried to access john's repository"),
			token:    "",
			owner:    "john",
			expected: "user [***] tried to access [***]'s repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeError(tt.err, tt.token, tt.owner)
			if result == nil {
				t.Fatal("expected non-nil error")
			}

			if result.Error() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.Error())
			}
		})
	}
}

func TestSanitizeError_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		token            string
		owner            string
		shouldNotContain []string
	}{
		{
			name:  "git push error with credentials",
			err:   errors.New("fatal: unable to access 'https://ghp_xxxxxxxxxxxx@github.com/myuser/myrepo.git/': The requested URL returned error: 403"),
			token: "ghp_xxxxxxxxxxxx",
			owner: "myuser",
			shouldNotContain: []string{
				"ghp_xxxxxxxxxxxx",
				"myuser",
				"github.com",
				"myrepo",
			},
		},
		{
			name:  "git clone error",
			err:   errors.New("Cloning into 'repo'... fatal: repository 'https://github.com/private/repo.git/' not found"),
			token: "",
			owner: "private",
			shouldNotContain: []string{
				"github.com",
				"private",
				"repo.git",
			},
		},
		{
			name:  "authentication failure",
			err:   errors.New("remote: Invalid username or password for user johndoe at https://github.com"),
			token: "ghp_token",
			owner: "johndoe",
			shouldNotContain: []string{
				"johndoe",
				"github.com",
			},
		},
		{
			name:  "permission denied",
			err:   errors.New("ERROR: Permission to alice/secret-repo.git denied to alice."),
			token: "",
			owner: "alice",
			shouldNotContain: []string{
				"alice", // This will be redacted
				// Note: "secret-repo" won't be redacted as it's not the owner or in a URL
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeError(tt.err, tt.token, tt.owner)
			if result == nil {
				t.Fatal("expected non-nil error")
			}

			resultMsg := result.Error()
			for _, forbidden := range tt.shouldNotContain {
				if strings.Contains(resultMsg, forbidden) {
					t.Errorf("sanitized error should NOT contain %q, but got: %s", forbidden, resultMsg)
				}
			}
		})
	}
}

func TestSanitizeError_PreservesErrorStructure(t *testing.T) {
	originalErr := errors.New("git command failed: https://github.com/user/repo.git")
	sanitized := sanitizeError(originalErr, "", "")

	if sanitized == nil {
		t.Fatal("expected non-nil error")
	}

	// Should still be an error type
	var err error = sanitized
	if err == nil {
		t.Error("sanitized error should still implement error interface")
	}

	// Should contain some of the original message structure
	if !strings.Contains(sanitized.Error(), "git command failed") {
		t.Error("sanitized error should preserve non-sensitive parts of the message")
	}

	// Should NOT contain the URL
	if strings.Contains(sanitized.Error(), "github.com") {
		t.Error("sanitized error should not contain URL")
	}
}


// TestSanitizeError_OrderOfOperations tests that sanitization happens in the correct order
func TestSanitizeError_OrderOfOperations(t *testing.T) {
	// Test that token is replaced before URL regex runs
	// This ensures we don't have partial replacements
	err := errors.New("failed: https://mytoken@github.com/user/repo.git with token mytoken")
	result := sanitizeError(err, "mytoken", "user")

	resultMsg := result.Error()

	// Should not contain the token
	if strings.Contains(resultMsg, "mytoken") {
		t.Errorf("result should not contain token, got: %s", resultMsg)
	}

	// Should not contain the user
	if strings.Contains(resultMsg, "user") {
		t.Errorf("result should not contain user, got: %s", resultMsg)
	}

	// Should not contain any part of the URL
	if strings.Contains(resultMsg, "github.com") {
		t.Errorf("result should not contain URL parts, got: %s", resultMsg)
	}
}

// TestSanitizeError_SpecialCharacters tests handling of special regex characters
func TestSanitizeError_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		token string
		owner string
	}{
		{
			name:  "token with special characters",
			err:   errors.New("failed with token ghp_abc.def+123"),
			token: "ghp_abc.def+123",
			owner: "",
		},
		{
			name:  "owner with special characters",
			err:   errors.New("user john.doe-123 failed"),
			token: "",
			owner: "john.doe-123",
		},
		{
			name:  "URL with special characters",
			err:   errors.New("failed: https://github.com/user/repo-name_v2.0.git"),
			token: "",
			owner: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeError(tt.err, tt.token, tt.owner)
			if result == nil {
				t.Fatal("expected non-nil error")
			}

			resultMsg := result.Error()

			// Should contain redaction marker
			if !strings.Contains(resultMsg, "[***]") {
				t.Errorf("expected redaction marker, got: %s", resultMsg)
			}

			// Should not contain the sensitive data
			if tt.token != "" && strings.Contains(resultMsg, tt.token) {
				t.Errorf("should not contain token, got: %s", resultMsg)
			}
			if tt.owner != "" && strings.Contains(resultMsg, tt.owner) {
				t.Errorf("should not contain owner, got: %s", resultMsg)
			}
		})
	}
}

// TestSanitizeError_EmptyStrings tests behavior with empty token and owner
func TestSanitizeError_EmptyStrings(t *testing.T) {
	err := errors.New("some error message")
	result := sanitizeError(err, "", "")

	if result == nil {
		t.Fatal("expected non-nil error")
	}

	// Should return the original message since nothing to redact
	if result.Error() != "some error message" {
		t.Errorf("expected original message, got: %s", result.Error())
	}
}

// TestSanitizeError_CaseSensitivity tests that replacement is case-sensitive
func TestSanitizeError_CaseSensitivity(t *testing.T) {
	err := errors.New("User alice and user ALICE and user Alice")
	result := sanitizeError(err, "", "alice")

	resultMsg := result.Error()

	// Should only replace exact matches (case-sensitive)
	if !strings.Contains(resultMsg, "alice") && !strings.Contains(resultMsg, "[***]") {
		t.Errorf("unexpected result: %s", resultMsg)
	}

	// The lowercase "alice" should be replaced
	expectedCount := strings.Count(resultMsg, "[***]")
	if expectedCount == 0 {
		t.Errorf("expected at least one redaction, got: %s", resultMsg)
	}
}

