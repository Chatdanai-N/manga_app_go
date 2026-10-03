package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type expectedTestResult struct {
	name           string
	mockBehavior   func(mock sqlmock.Sqlmock)
	expectedStatus int
	expectedBody   string
}

func init() {
	// ปิด log ของ Gin ตอนรัน test
	gin.SetMode(gin.TestMode)
}

func TestMangaSystemParameterHandler_GetMangaSystemParameterById_Success(t *testing.T) {
	expectedTests := []expectedTestResult{
		expectedTestResult{
			name: "Success - Get Parameters List",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				// สร้าง mock rows สำหรับตาราง manga_system_parameter
				rows := sqlmock.NewRows([]string{"parameter_id", "group_code", "code", "value", "order_number"}).
					AddRow(1, "SYSTEM", "MAX_LIMIT", "100", 1).
					AddRow(2, "SYSTEM", "MIN_LIMIT", "10", 2)

				// ดักจับ Query SELECT * FROM "manga_system_parameter"
				mock.ExpectQuery(`^SELECT \* FROM "manga_system_parameter"`).
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"parameterId":1,"groupCode":"SYSTEM","code":"MAX_LIMIT","value":"100","orderNumber":1},{"parameterId":2,"groupCode":"SYSTEM","code":"MIN_LIMIT","value":"10","orderNumber":2}]`,
		},
	}

	for _, test := range expectedTests {
		t.Run(test.name, func(t *testing.T) {
			// 1. Arrange: สร้าง mock sql.DB และ GORM instance
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock: %v", err)
			}
			defer mockDB.Close()

			gormDB, err := gorm.Open(postgres.New(postgres.Config{
				Conn: mockDB,
			}), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm db: %v", err)
			}

			// เรียกใช้ mockBehavior ของแต่ละ case
			test.mockBehavior(mock)

			// สร้าง Handler พร้อมฉีด gormDB เข้าไป
			handler := &MangaSystemParameterHandler{DB: gormDB}

			// ตั้งค่า Gin Router
			r := gin.New()
			r.GET("/system-parameters", handler.GetMangaSystemParameters)

			// 2. Act: จำลอง HTTP Request
			req, _ := http.NewRequest("GET", "/system-parameters", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			// 3. Assert: ตรวจสอบ Status Code และ Body
			if w.Code != test.expectedStatus {
				t.Errorf("expected status %d, got %d", test.expectedStatus, w.Code)
			}

			if w.Body.String() != test.expectedBody {
				t.Errorf("expected body %s, got %s", test.expectedBody, w.Body.String())
			}

			// ตรวจสอบว่า Query ที่ตั้ง mock ไว้ถูกเรียกใช้ครบถ้วน
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMangaSystemParameterHandler_GetMangaSystemParameterById_Error(t *testing.T) {

	// 1. Arrange: สร้าง mock database
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer mockDB.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: mockDB,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	// บังคับให้ GORM Query ตาราง manga_system_parameter แล้วคาย Error กลับมา
	expectedError := errors.New("database connection failed")
	mock.ExpectQuery(`^SELECT \* FROM "manga_system_parameter"`).WillReturnError(expectedError)

	handler := &MangaSystemParameterHandler{DB: gormDB}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/system-parameters", handler.GetMangaSystemParameters)

	// 2. Act: ส่ง Request
	req, _ := http.NewRequest("GET", "/system-parameters", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 3. Assert: ตรวจสอบว่าได้ Status 400 และ Error Message ตรงกัน
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	expectedBody := `{"error":"database connection failed"}`
	if strings.TrimSpace(w.Body.String()) != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, w.Body.String())
	}

	// ตรวจสอบว่า Expectation ที่ตั้งไว้ถูกเรียกใช้จริง
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
