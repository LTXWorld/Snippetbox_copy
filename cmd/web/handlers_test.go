package main

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"
)

// 这里的测试都是在用testutils_test.go中抽象出的逻辑代码使其配合表测试具体化

// 复用 testutils_test.go中的代码，端到端的测试GET /ping
func TestPing(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/ping")

	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
	}

	if string(body) != "OK" {
		t.Errorf("want body to equal %q", "OK")
	}
}

func TestShowSnippet(t *testing.T) {
	// 使用的是测试中的app（其包括了模仿的一些方法用于测试）
	app := newTestApplication(t)

	// Establish a new test server for running end-to-end tests
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody []byte
	}{
		{"Valid ID", "/snippet/1", http.StatusOK, []byte("An old silent pond...")},
		{"Non-existent ID", "/snippet/2", http.StatusNotFound, nil},
		{"Negative ID", "/snippet/-1", http.StatusNotFound, nil},
		{"Decimal ID", "/snippet/1.23", http.StatusNotFound, nil},
		{"String ID", "/snippet/foo", http.StatusNotFound, nil},
		{"Empty ID", "/snippet/", http.StatusNotFound, nil},
		{"Trailing slash", "/snippet/1/", http.StatusNotFound, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body to contain %q", tt.wantBody)
			}
		})
	}
}

// 先测试 GET /user/signup得到CSRF令牌，在此基础上测试 POST /user/signup
func TestSignupUser(t *testing.T) {
	// Make the end-to-end test
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// Make a GET /user/signup request extract the CSRF token
	_, _, body := ts.get(t, "/user/signup")
	csrfToken := extractCSRFToken(t, body)

	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantBody     []byte
	}{
		{"Valid submission", "Bob", "bob@example.com", "validPassword", csrfToken, http.StatusOK, []byte("Valid submission")},
		{"Empty name", "", "bob@example.com", "validPa$$word", csrfToken, http.StatusBadRequest, []byte("Empty name error message")},
		{"Empty email", "Bob", "", "validPassword", "csrfToken", http.StatusOK, []byte("Empty email error message")},
		{"Empty password", "Bob", "bob@example.com", "", "csrfToken", http.StatusOK, []byte("Empty password error message")},
		{"Invalid email (incomplete domain)", "Bob", "bob@example.", "validPassword", "csrfToken", http.StatusOK, []byte("Invalid email error message")},
		{"Invalid email (missing @)", "Bob", "bobexample.com", "validPassword", "csrfToken", http.StatusOK, []byte("Invalid email error message")},
		{"Invalid email (missing local part)", "Bob", "@example.com", "validPassword", "csrfToken", http.StatusOK, []byte("Invalid email error message")},
		{"Short password", "Bob", "bobexample.com", "password", "csrfToken", http.StatusOK, []byte("Short password error message")},
		{"Duplicate email", "Bob", "dupeexample.com", "validPassword", "csrfToken", http.StatusOK, []byte("Duplicate email error message")},
		{"Invalid CSRF Token", "", "", "", "wrongToken", http.StatusBadRequest, []byte("Invalid CSRF Token error message")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)
			code, _, body := ts.postForm(t, "/user/signup", form)
			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}
			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body %s to contain %q", body, tt.wantBody)
			}
		})
	}
}
