package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type testSuit struct {
	name           string
	mockBehavior   func(mock sqlmock.Sqlmock)
	expectedStatus int
	expectedBody   string
	parameterId    string
	reqBody        interface{}
}

func init() {
	// ปิด log ของ Gin ตอนรัน test
	gin.SetMode(gin.TestMode)
}

func TestNewMangaSystemParameterHandler(t *testing.T) {

	// arrange
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	handler := NewMangaSystemParameterHandler(gormDB)
	if handler == nil {
		t.Errorf("NewMangaSystemParameterHandler returned nil")
	}

}

func TestMangaSystemParameterHandler_GetMangaSystemParameters_Success(t *testing.T) {
	testSuits := []testSuit{
		{
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
		{
			name: "Database Error - 400 Bad Request",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				expectedError := errors.New("database connection failed")
				mock.ExpectQuery(`^SELECT \* FROM "manga_system_parameter"`).
					WillReturnError(expectedError)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"database connection failed"}`,
		},
	}

	for _, ts := range testSuits {
		t.Run(ts.name, func(t *testing.T) {
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
			ts.mockBehavior(mock)

			// สร้าง Handler พร้อมฉีด gormDB เข้าไป
			handler := &MangaSystemParameterHandler{DB: gormDB}

			// ตั้งค่า Gin Router
			r := gin.New()
			r.GET("/system-parameter", handler.GetMangaSystemParameters)

			// 2. Act: จำลอง HTTP Request
			req, _ := http.NewRequest("GET", "/system-parameter", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			// 3. Assert: ตรวจสอบ Status Code และ Body
			if w.Code != ts.expectedStatus {
				t.Errorf("expected status %d, got %d", ts.expectedStatus, w.Code)
			}

			if w.Body.String() != ts.expectedBody {
				t.Errorf("expected body %s, got %s", ts.expectedBody, w.Body.String())
			}

			// ตรวจสอบว่า Query ที่ตั้ง mock ไว้ถูกเรียกใช้ครบถ้วน
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMangaSystemParameterHandler_GetMangaSystemParameterById(t *testing.T) {

	testSuits := []testSuit{

		{
			name: "Record Not Found - 404 Not Found",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				expectedSQL := `SELECT * FROM "manga_system_parameter" WHERE "manga_system_parameter"."parameter_id" = $1 ORDER BY "manga_system_parameter"."parameter_id" LIMIT $2`
				mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
					WithArgs(999, 1).
					WillReturnError(gorm.ErrRecordNotFound) // $1 = 999, $2 = 1 (LIMIT 1 ของ GORM First)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Record not found"}`,
			parameterId:    "999",
		},
		{
			name: "Invalid ID Format - 400 Bad Request",
			mockBehavior: func(mock sqlmock.Sqlmock) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid ID format"}`,
			parameterId:    "A",
		},
		{
			name: "Database Error - 500 Internal Server Error",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				expectedSQL := `SELECT * FROM "manga_system_parameter" WHERE "manga_system_parameter"."parameter_id" = $1 ORDER BY "manga_system_parameter"."parameter_id" LIMIT $2`
				mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
					WithArgs(999, 1).
					WillReturnError(errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
			parameterId:    "999",
		},
		{
			name: "Success - 200 OK",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				expectedSQL := `SELECT * FROM "manga_system_parameter" WHERE "manga_system_parameter"."parameter_id" = $1 ORDER BY "manga_system_parameter"."parameter_id" LIMIT $2`
				rows := sqlmock.NewRows([]string{"parameter_id", "group_code", "code", "value", "order_number"}).
					AddRow(1, "mock group", "mock code", "100", 1)
				mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"parameterId":1,"groupCode":"mock group","code":"mock code","value":"100","orderNumber":1}`,
			parameterId:    "1",
		},
	}
	for _, ts := range testSuits {
		t.Run(ts.name, func(t *testing.T) {

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
			ts.mockBehavior(mock)

			// สร้าง Handler พร้อมฉีด gormDB เข้าไป
			handler := &MangaSystemParameterHandler{DB: gormDB}

			// ตั้งค่า Gin Router
			r := gin.New()
			r.GET("/system-parameters/:id", handler.GetMangaSystemParameterById)

			// 2. Act:
			req, _ := http.NewRequest("GET", "/system-parameters/"+ts.parameterId, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			// 3. Assert
			if w.Code != ts.expectedStatus {
				t.Errorf("expected status %d, got %d", ts.expectedStatus, w.Code)
			}

			if w.Body.String() != ts.expectedBody {
				t.Errorf("expected body %s, got %s", ts.expectedBody, w.Body.String())
			}

			// ตรวจสอบว่า Mock SQL Query ถูกเรียกครบถ้วนตาม Expectation หรือไม่
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMangaSystemParameterHandler_CreateMangaSystemParameter(t *testing.T) {

	testSuits := []testSuit{
		{
			name: "Success - 201 Created",
			reqBody: RequestMangaParameter{
				GroupCode:   "SYSTEM",
				Code:        "MAXLIMIT",
				Value:       "100",
				OrderNumber: 1,
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				expectedSQL := `INSERT INTO "manga_system_parameter"`

				mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
					WithArgs("SYSTEM", "MAXLIMIT", "100", 1).
					WillReturnRows(sqlmock.NewRows([]string{"parameter_id"}).AddRow(1))
				mock.ExpectCommit()
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"parameterId":1,"groupCode":"SYSTEM","code":"MAXLIMIT","value":"100","orderNumber":1}`,
		},
		{
			name: "Invalid Body Request - 400 Bad Request",
			reqBody: RequestMangaParameter{
				Code:        "MAXLIMIT",
				Value:       "100",
				OrderNumber: 1,
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Key: 'RequestMangaParameter.GroupCode' Error:Field validation for 'GroupCode' failed on the 'required' tag"}`,
		},
		{
			name: "Failed to create record - 500 Internal Server Error",
			reqBody: RequestMangaParameter{
				GroupCode:   "SYSTEM",
				Code:        "MAXLIMIT",
				Value:       "100",
				OrderNumber: 1,
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				expectedSQL := `INSERT INTO "manga_system_parameter"`

				mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
					WithArgs("SYSTEM", "MAXLIMIT", "100", 1).
					WillReturnError(errors.New("Failed to create record"))

				//เมื่อเกิด Error ตัว GORM จะสั่ง Rollback (ห้ามใช้ ExpectCommit)
				mock.ExpectRollback()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to create record"}`,
		},
	}
	for _, ts := range testSuits {
		t.Run(ts.name, func(t *testing.T) {
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
			ts.mockBehavior(mock)

			// สร้าง Handler พร้อมฉีด gormDB เข้าไป
			handler := &MangaSystemParameterHandler{DB: gormDB}
			// ตั้งค่า Gin Router
			r := gin.New()
			r.POST("/system-parameters", handler.CreateMangaSystemParameter)

			// แปลง reqBody ให้เป็น []byte
			var jsonBytes []byte
			if str, ok := ts.reqBody.(string); ok {
				jsonBytes = []byte(str)
			} else {
				jsonBytes, _ = json.Marshal(ts.reqBody)
			}

			// 2. Act:
			req, _ := http.NewRequest("POST", "/system-parameters", bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			// 3. Assert
			if w.Code != ts.expectedStatus {
				t.Errorf("expected status %d, got %d", ts.expectedStatus, w.Code)
			}

			if w.Body.String() != ts.expectedBody {
				t.Errorf("expected body %s, got %s", ts.expectedBody, w.Body.String())
			}

			// ตรวจสอบว่า Mock SQL Query ถูกเรียกครบถ้วนตาม Expectation หรือไม่
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMangaSystemParameterHandler_PutMangaSystemParameter(t *testing.T) {
	testSuits := []testSuit{
		{
			name: "Invalid Body Request - 400 Bad Request",
			reqBody: RequestMangaParameter{
				Code:        "MAXLIMIT",
				Value:       "100",
				OrderNumber: 1,
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Key: 'RequestMangaParameter.GroupCode' Error:Field validation for 'GroupCode' failed on the 'required' tag"}`,
		},
		{
			name: "Record Not Found - 404 Not Found",
			reqBody: RequestMangaParameter{
				ParameterId: 999,
				GroupCode:   "SYSTEM",
				Code:        "TEST",
				Value:       "100",
				OrderNumber: 999,
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				expectedSQL := `SELECT * FROM "manga_system_parameter" WHERE "manga_system_parameter"."parameter_id" = $1 ORDER BY "manga_system_parameter"."parameter_id" LIMIT $2`
				mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
					WithArgs(999, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Record not found"}`,
		},
		{
			name: "Internal Server Error - 500 Internal Server Error",
			reqBody: RequestMangaParameter{
				ParameterId: 999,
				GroupCode:   "SYSTEM",
				Code:        "TEST",
				Value:       "100",
				OrderNumber: 999,
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				expectedSQL := `SELECT * FROM "manga_system_parameter" WHERE "manga_system_parameter"."parameter_id" = $1 ORDER BY "manga_system_parameter"."parameter_id" LIMIT $2`
				mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
					WithArgs(999, 1).
					WillReturnError(errors.New("Failed to connect database"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
		{
			name: "Failed to create record - 500 Internal Server Error",
			reqBody: RequestMangaParameter{
				ParameterId: 999,
				GroupCode:   "SYSTEM",
				Code:        "MAX_LIMIT",
				Value:       "100",
				OrderNumber: 1,
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				// 1.mock select
				expectedSQL := `SELECT * FROM "manga_system_parameter" WHERE "manga_system_parameter"."parameter_id" = $1 ORDER BY "manga_system_parameter"."parameter_id" LIMIT $2`

				rows := sqlmock.NewRows([]string{"parameter_id", "group_code", "code", "value", "order_number"}).
					AddRow(999, "SYSTEM", "MAX_LIMIT", "100", 1)

				mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
					WithArgs(999, 1).
					WillReturnRows(rows)

				// 2.mock update
				mock.ExpectBegin()

				expectedUpdateSql := `UPDATE "manga_system_parameter" SET "group_code"=\$1,"code"=\$2,"value"=\$3,"order_number"=\$4 WHERE "parameter_id" = \$5`

				mock.ExpectExec(expectedUpdateSql).
					WithArgs("SYSTEM", "MAX_LIMIT", "100", 1, uint64(999)).
					WillReturnError(errors.New("Failed to connect database"))

				mock.ExpectRollback()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to update record"}`,
		},
		{
			name: "Success to Update record - 200 OK",
			reqBody: RequestMangaParameter{
				ParameterId: 999,
				GroupCode:   "SYSTEM",
				Code:        "MAX_LIMIT",
				Value:       "101",
				OrderNumber: 1,
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				// 1.mock select
				expectedSQL := `SELECT * FROM "manga_system_parameter" WHERE "manga_system_parameter"."parameter_id" = $1 ORDER BY "manga_system_parameter"."parameter_id" LIMIT $2`

				rows := sqlmock.NewRows([]string{"parameter_id", "group_code", "code", "value", "order_number"}).
					AddRow(999, "SYSTEM", "MAX_LIMIT", "100", 1)

				mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
					WithArgs(999, 1).
					WillReturnRows(rows)

				// 2.mock update
				mock.ExpectBegin()

				expectedUpdateSql := `UPDATE "manga_system_parameter" SET "group_code"=\$1,"code"=\$2,"value"=\$3,"order_number"=\$4 WHERE "parameter_id" = \$5`
				mock.ExpectExec(expectedUpdateSql).
					WithArgs("SYSTEM", "MAX_LIMIT", "101", 1, uint64(999)).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"parameterId":999,"groupCode":"SYSTEM","code":"MAX_LIMIT","value":"101","orderNumber":1}`,
		},
	}

	for _, ts := range testSuits {
		t.Run(ts.name, func(t *testing.T) {
			// 1.Arrange:
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
			ts.mockBehavior(mock)

			// สร้าง Handler พร้อมฉีด gormDB เข้าไป
			handler := &MangaSystemParameterHandler{DB: gormDB}
			// ตั้งค่า Gin Router
			r := gin.New()
			r.PUT("/system-parameter", handler.PutMangaSystemParameter)

			// แปลง reqBody ให้เป็น []byte
			var jsonBytes []byte
			if str, ok := ts.reqBody.(string); ok {
				jsonBytes = []byte(str)
			} else {
				jsonBytes, _ = json.Marshal(ts.reqBody)
			}
			// 2.Ack
			req, _ := http.NewRequest("PUT", "/system-parameter", bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			// 3.Assert
			if w.Code != ts.expectedStatus {
				t.Errorf("expected status %d, got %d", ts.expectedStatus, w.Code)
			}

			if w.Body.String() != ts.expectedBody {
				t.Errorf("expected body %s, got %s", ts.expectedBody, w.Body.String())
			}

			// ตรวจสอบว่า Mock SQL Query ถูกเรียกครบถ้วนตาม Expectation หรือไม่
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMangaSystemParameterHandler_PatchMangaSystemParameter(t *testing.T) {
	testSuits := []testSuit{
		{
			name: "Invalid ID format - 400 Bad Request",
			reqBody: RequestMangaParameter{
				Value: "100",
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid ID format"}`,
			parameterId:    "Hello",
		},
		{
			name: "Invalid Body Request - 400 Bad Request",
			reqBody: RequestMangaParameter{
				ParameterId: 999,
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Key: 'RequestPatchMangaSystemParameter.Value' Error:Field validation for 'Value' failed on the 'required' tag"}`,
			parameterId:    "999",
		},
		{
			name: "Record not Found - 404 Not Found",
			reqBody: RequestMangaParameter{
				Value: "100",
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				expectedSQL := `SELECT * FROM "manga_system_parameter" WHERE "manga_system_parameter"."parameter_id" = $1 ORDER BY "manga_system_parameter"."parameter_id" LIMIT $2`
				mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
					WithArgs(999, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Record not found"}`,
			parameterId:    "999",
		},
		{
			name: "Internal server error - 500 Internal server error",
			reqBody: RequestMangaParameter{
				Value: "100",
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				expectedSQL := `SELECT * FROM "manga_system_parameter" WHERE "manga_system_parameter"."parameter_id" = $1 ORDER BY "manga_system_parameter"."parameter_id" LIMIT $2`

				mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
					WithArgs(999, 1).
					WillReturnError(errors.New("error connection database"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
			parameterId:    "999",
		},
		{
			name: "Failed to Update record - 500 Internal server error",
			reqBody: RequestMangaParameter{
				Value: "100",
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				// 1. mock select statement
				rows := sqlmock.NewRows([]string{"parameter_id", "group_code", "code", "value", "order_number"}).
					AddRow(999, "SYSTEM", "MAX_LIMIT", "100", 1)

				expectedSQL := `SELECT * FROM "manga_system_parameter" WHERE "manga_system_parameter"."parameter_id" = $1 ORDER BY "manga_system_parameter"."parameter_id" LIMIT $2`
				mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
					WithArgs(999, 1).
					WillReturnRows(rows)

				// 2. mock update statement
				mock.ExpectBegin()

				expectedUpdateSQL := `UPDATE "manga_system_parameter" SET "value"=$1 WHERE "parameter_id" = $2`
				mock.ExpectExec(regexp.QuoteMeta(expectedUpdateSQL)).
					WithArgs("100", uint64(999)).
					WillReturnError(errors.New("error connection database"))

				mock.ExpectRollback()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to update record"}`,
			parameterId:    "999",
		},
		{
			name: "Success to update record - 200 OK",
			reqBody: RequestMangaParameter{
				Value: "101",
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				// 1. mock select statement
				rows := sqlmock.NewRows([]string{"parameter_id", "group_code", "code", "value", "order_number"}).
					AddRow(999, "SYSTEM", "MAX_LIMIT", "100", 1)

				expectedSQL := `SELECT * FROM "manga_system_parameter" WHERE "manga_system_parameter"."parameter_id" = $1 ORDER BY "manga_system_parameter"."parameter_id" LIMIT $2`
				mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
					WithArgs(999, 1).
					WillReturnRows(rows)
				// 2. mock update statement
				mock.ExpectBegin()

				expectedUpdateSQL := `UPDATE "manga_system_parameter" SET "value"=$1 WHERE "parameter_id" = $2`
				mock.ExpectExec(regexp.QuoteMeta(expectedUpdateSQL)).
					WithArgs("101", uint64(999)).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"parameterId":999,"groupCode":"SYSTEM","code":"MAX_LIMIT","value":"101","orderNumber":1}`,
			parameterId:    "999",
		},
	}

	for _, ts := range testSuits {
		t.Run(ts.name, func(t *testing.T) {
			// 1.Arrage
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
			ts.mockBehavior(mock)

			// สร้าง Handler พร้อมฉีด gormDB เข้าไป
			handler := &MangaSystemParameterHandler{DB: gormDB}

			// ตั้งค่า Gin Router
			r := gin.New()
			r.PATCH("/system-parameter/:id", handler.PatchMangaSystemParameter)

			var jsonBytes []byte
			if str, ok := ts.reqBody.(string); ok {
				jsonBytes = []byte(str)
			} else {
				jsonBytes, _ = json.Marshal(ts.reqBody)
			}

			// 2. Act:
			req, _ := http.NewRequest("PATCH", "/system-parameter/"+ts.parameterId, bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			// 3.Assert
			if w.Code != ts.expectedStatus {
				t.Errorf("expected status %d, got %d", ts.expectedStatus, w.Code)
			}

			if w.Body.String() != ts.expectedBody {
				t.Errorf("expected body %s, got %s", ts.expectedBody, w.Body.String())
			}

			// ตรวจสอบว่า Mock SQL Query ถูกเรียกครบถ้วนตาม Expectation หรือไม่
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}

}

func TestMangaSystemParameterHandler_DeleteMangaSystemParameterById(t *testing.T) {
	testSuits := []testSuit{
		{
			name: "Invalid ID format - 400 Bad Request",
			mockBehavior: func(mock sqlmock.Sqlmock) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid ID format"}`,
			parameterId:    "Hello",
		},
		{
			name: "Invalid ID format - 500 Internal Server Error",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				expectedSQL := `DELETE FROM "manga_system_parameter" WHERE "manga_system_parameter"."parameter_id" = $1`
				mock.ExpectExec(regexp.QuoteMeta(expectedSQL)).
					WithArgs(uint64(999)).
					WillReturnError(gorm.ErrInvalidTransaction)

				mock.ExpectRollback()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
			parameterId:    "999",
		},
		{
			name: "Record not found - 404 not found",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				expectedSQL := `DELETE FROM "manga_system_parameter" WHERE "manga_system_parameter"."parameter_id" = $1`
				mock.ExpectExec(regexp.QuoteMeta(expectedSQL)).
					WithArgs(uint64(999)).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Record not found"}`,
			parameterId:    "999",
		},
		{
			name: "Delete Success - 200 Successfully deleted record",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				expectedSQL := `DELETE FROM "manga_system_parameter" WHERE "manga_system_parameter"."parameter_id" = $1`
				mock.ExpectExec(regexp.QuoteMeta(expectedSQL)).
					WithArgs(uint64(999)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Successfully deleted record"}`,
			parameterId:    "999",
		},
	}

	for _, ts := range testSuits {
		t.Run(ts.name, func(t *testing.T) {
			// 1.Arrage
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
			ts.mockBehavior(mock)

			// สร้าง Handler พร้อมฉีด gormDB เข้าไป
			handler := &MangaSystemParameterHandler{DB: gormDB}

			// ตั้งค่า Gin Router
			r := gin.New()
			r.DELETE("/system-parameter/:id", handler.DeleteMangaSystemParameterById)
			// 2.Ack
			req, _ := http.NewRequest("DELETE", "/system-parameter/"+ts.parameterId, nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			// 3.Assert
			if w.Code != ts.expectedStatus {
				t.Errorf("expected status %d, got %d", ts.expectedStatus, w.Code)
			}

			if w.Body.String() != ts.expectedBody {
				t.Errorf("expected body %s, got %s", ts.expectedBody, w.Body.String())
			}

			// ตรวจสอบว่า Mock SQL Query ถูกเรียกครบถ้วนตาม Expectation หรือไม่
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
