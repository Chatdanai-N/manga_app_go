package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Testsuite struct {
	name           string
	username       string
	userId         string
	secretKey      string
	httpStatusCode int
	expected       string
}

func generateTestToken(t *testing.T, test Testsuite) string {
	t.Helper()

	if test.name == "Status 401 Unauthorized invalid token or expired token" {
		token := jwt.New(jwt.SigningMethodNone)
		tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if err != nil {
			t.Errorf("Failed to sign unsafe token: %s", err)
			return ""
		}
		return tokenStr
	}

	claims := jwt.MapClaims{
		"username": test.username,
		"sub":      test.userId,
		"exp":      time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(test.secretKey))
	if err != nil {
		t.Errorf("Failed to sign token: %s", err)
		return ""
	}

	return tokenString

}

func TestAuthMiddleware(t *testing.T) {

	tests := []Testsuite{
		{
			name:           "Status 200 Success",
			username:       "user",
			userId:         "1234",
			secretKey:      "my-secert-key",
			httpStatusCode: http.StatusOK,
			expected:       `{"user_id":"1234","username":"user"}`,
		},
		{
			name:           "Status 401 Unauthorized missing header",
			username:       "user",
			userId:         "1234",
			secretKey:      "my-secret-key",
			httpStatusCode: http.StatusUnauthorized,
			expected:       `{"error":"Authorization header is required"}`,
		},
		{
			name:           "Status 401 Unauthorized invalid bearer token",
			username:       "user",
			userId:         "1234",
			secretKey:      "my-secret-key",
			httpStatusCode: http.StatusUnauthorized,
			expected:       `{"error":"Authorization header format must be Bearer \u003ctoken\u003e"}`,
		},
		{
			name:           "Status 401 Unauthorized invalid token or expired token",
			username:       "user",
			userId:         "1234",
			secretKey:      "my-secret-key",
			httpStatusCode: http.StatusUnauthorized,
			expected:       `{"error":"Invalid or expired token"}`,
		},
	}

	for _, test := range tests {
		gin.SetMode(gin.TestMode)

		r := gin.New()
		r.Use(AuthMiddleware(test.secretKey))

		r.GET("/test", func(c *gin.Context) {

			username, _ := c.Get("username")
			userId, _ := c.Get("user_id")
			if test.name == `Status 200 Success` {
				c.JSON(http.StatusOK, gin.H{
					"username": username,
					"user_id":  userId,
				})
			}
		})

		tokenStr := generateTestToken(t, test)

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)

		authHeaderValue := "Bearer " + tokenStr

		switch test.name {
		case "Status 401 Unauthorized invalid bearer token":
			authHeaderValue = "Baerer " + tokenStr
		case "Status 401 Unauthorized missing header":
			authHeaderValue = ""
		}

		if authHeaderValue != "" {
			req.Header.Add("Authorization", authHeaderValue)
		}

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != test.httpStatusCode {
			t.Errorf("handler returned wrong status code: got %v want %v", w.Code, test.httpStatusCode)
		}

		if w.Body.String() != test.expected {
			t.Errorf("handler returned unexpected body: got %v want %v", w.Body.String(), test.expected)
		}
	}

}
