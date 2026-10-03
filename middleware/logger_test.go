package middleware

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type expectedResultTest struct {
	name           string
	method         string
	path           string
	handler        gin.HandlerFunc
	expectedStatus int
}

func TestInitLogger(t *testing.T) {
	// 1. Arrange: ดักจับ os.Stdout เพื่อเอาไว้ตรวจจับ Log Text ที่พิมพ์ออกมา
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	// 2. Act: เรียกใช้ InitLogger()
	logger := InitLogger()

	// 3. Assert: ตรวจสอบว่าได้ logger กลับมา และ slog.Default() เปลี่ยนค่าแล้ว
	if logger == nil {
		t.Fatal("expected logger to not be nil")
	}

	if slog.Default() != logger {
		t.Error("expected slog.Default() to be set to the created logger")
	}

	// ทดสอบยิง Log 2 แบบ (Info ต้องออก, Debug ต้องถูกกรองออกเพราะตั้งไว้ที่ LevelInfo)
	slog.Debug("this debug log should be ignored")
	slog.Info("test info log", "key", "value")

	// ปิด Pipe เพื่ออ่านข้อมูล Log ที่พิมพ์ออกทาง Stdout
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// 4. Assert Output Log Format & Level Filter
	// ต้องมีข้อความ Info log
	if !strings.Contains(output, `"msg":"test info log"`) {
		t.Errorf("expected log output to contain info message, got: %s", output)
	}

	// ต้องเป็น JSON Format
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Errorf("expected log level to be INFO in JSON format, got: %s", output)
	}

	// ต้องไม่มี Debug Log หลุดออกมา
	if strings.Contains(output, "this debug log should be ignored") {
		t.Errorf("did not expect debug log output, got: %s", output)
	}
}

func TestStructuredLogger(t *testing.T) {

	//data driven test
	tests := []expectedResultTest{
		expectedResultTest{
			name:   "Status 200 - Info Log",
			method: "GET",
			path:   "/success",
			handler: func(c *gin.Context) {
				c.Status(http.StatusOK)
			},
			expectedStatus: http.StatusOK,
		},
		expectedResultTest{
			name:   "Status 400 - Warn Log",
			method: "GET",
			path:   "/bad-request",
			handler: func(c *gin.Context) {
				c.Status(http.StatusBadRequest)
			},
			expectedStatus: http.StatusBadRequest,
		},
		expectedResultTest{
			name:   "Status 500 - Error Log",
			method: "POST",
			path:   "/internal-error",
			handler: func(c *gin.Context) {
				c.Status(http.StatusInternalServerError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		expectedResultTest{
			name:   "With Private Error - Append Error Attr",
			method: "GET",
			path:   "/error-with-msg",
			handler: func(c *gin.Context) {
				// บันทึก Error เข้าไปใน gin.Context เพื่อให้ c.Errors.ByType() ทำงาน
				_ = c.Error(errors.New("db connection failed")).SetType(gin.ErrorTypePrivate)
				c.Status(http.StatusInternalServerError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// 1.Arrange
			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			router.Use(StructuredLogger())
			// 2.Act
			router.Handle(test.method, test.path, test.handler)
			//จำลอง HTTP Request และ ResponseRecorder
			req, _ := http.NewRequest(test.method, test.path, nil)
			//กำหนด Header เช่น User-Agent เพื่อทดสอบว่า Middleware สามารถดึงค่าไปลง Log ได้
			req.Header.Set("User-Agent", "Go-Test-Agent")
			//สร้าง Response Recorder ทำหน้าที่เป็น "ตัวรับ Response จำลอง" แทน Browser หรือ Client
			w := httptest.NewRecorder()
			//สั่งให้ Gin Engine ประมวลผล Request นั้นทันที โดยไม่ต้องเปิด Port Server จริง แล้วเขียนผลลัพธ์ลงใน
			router.ServeHTTP(w, req)
			// 3. Assert ตรวจสอบ HTTP Status Code
			if w.Code != test.expectedStatus {
				t.Errorf("expected status %d, got %d", test.expectedStatus, w.Code)
			}
		})
	}
}
