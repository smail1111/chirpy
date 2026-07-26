package auth_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smail1111/chirpy/internal/auth"
)

func TestPasswordAuth(t *testing.T) {
	cases := []struct {
		password string
		attempt  string
		expected bool
	}{
		{
			password: "1234",
			attempt:  "1234",
			expected: true,
		},
		{
			password: "0000",
			attempt:  "1234",
			expected: false,
		},
		{
			password: "applebannanacatdog",
			attempt:  "applebannanacatdog",
			expected: true,
		},
		{
			password: "Password",
			attempt:  "password",
			expected: false,
		},
	}

	for i, c := range cases {
		hash, er := auth.HashPassword(c.password)

		if er != nil {
			t.Errorf("Test %d failed -> %s", i, er.Error())
			continue
		}
		got, er := auth.CheckPasswordHash(c.attempt, hash)

		if er != nil {
			t.Errorf("Test %d failed -> %s", i, er.Error())
			continue
		}

		if got != c.expected {
			t.Errorf("Test %d failed -> %v != %v", i, got, c.expected)
		}
	}
}

func TestJwtAuth(t *testing.T) {
	cases := []struct {
		secret      string
		attempt     string
		expiresIn   time.Duration
		delay       time.Duration
		expectError bool
	}{
		{
			secret:      "1234",
			attempt:     "1234",
			expiresIn:   time.Minute,
			delay:       0,
			expectError: false,
		},
		{
			secret:      "0000",
			attempt:     "1234",
			expiresIn:   time.Minute,
			delay:       0,
			expectError: true,
		},
		{
			secret:      "1234",
			attempt:     "1234",
			expiresIn:   time.Second,
			delay:       time.Second,
			expectError: true,
		},
		{
			secret:      "abcd1234",
			attempt:     "abcd1234",
			expiresIn:   time.Second * 5,
			delay:       time.Second,
			expectError: false,
		},
	}

	for i, c := range cases {
		id, er := uuid.NewUUID()
		if er != nil {
			t.Errorf("Test %d Failed -> %s", i, er.Error())
			continue
		}

		str, er := auth.MakeJWT(id, c.secret, c.expiresIn)
		if er != nil {
			t.Errorf("Test %d Failed -> %s", i, er.Error())
			continue
		}

		time.Sleep(c.delay)

		got, er := auth.ValidateJWT(str, c.attempt)
		if (er != nil) != c.expectError {
			t.Errorf("Test %d Failed -> %s", i, er.Error())
			continue
		}
		if !c.expectError && got != id {
			t.Errorf("Test %d Falied -> %v != %v", i, got, id)
		}
	}
}

func TestGetBearerToken(t *testing.T) {
	cases := []struct {
		header      http.Header
		expectError bool
		expected    string
	}{
		{
			header:      http.Header{"Authorization": []string{"Bearer 1234"}},
			expectError: false,
			expected:    "1234",
		},
		{
			header:      http.Header{},
			expectError: true,
		},
	}

	for i, c := range cases {
		got, er := auth.GetBearerToken(c.header)

		if er != nil != c.expectError {
			t.Errorf("Test %d Failed -> %s", i, er.Error())
			continue
		}

		if !c.expectError && got != c.expected {
			t.Errorf("Test %d Falied -> %s != %s", i, got, c.expected)
		}
	}
}
