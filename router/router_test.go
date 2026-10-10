package router

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {

	gin.SetMode(gin.TestMode)
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// เชื่อม GORM เข้ากับ Mock Database
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	router := SetupRouter(gormDB)

	return router, mock
}

func TestSetupRouter_HealthCheck(t *testing.T) {
	router, _ := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("health check returned wrong status code: got %v want %v", w.Code, http.StatusOK)
	}

	expectedResults := `{"status":"UP"}`
	if w.Body.String() != expectedResults {
		t.Errorf("health check returned unexpected body: got %v want %v", w.Body.String(), expectedResults)
	}
}

func TestSetupRouter_GetParameterById_Route(t *testing.T) {
	router, mock := setupTestRouter(t)

	expectedSQL := `SELECT * FROM "manga_system_parameter" WHERE "manga_system_parameter"."parameter_id" = $1 ORDER BY "manga_system_parameter"."parameter_id" LIMIT $2`
	rows := sqlmock.NewRows([]string{"parameter_id", "group_code", "code", "value", "order_number"}).
		AddRow(1, "SYSTEM", "MAX_LIMIT", "100", 1)

	mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
		WithArgs(uint64(1), 1).
		WillReturnRows(rows)

	// จำลองยิง HTTP Request ไปที่ Route จริง
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/parameter/1", nil)

	router.ServeHTTP(w, req)

	// ตรวจสอบผลลัพธ์
	if w.Code != http.StatusOK {
		t.Errorf("GetParameterById returned wrong status code: got %v want %v", w.Code, http.StatusOK)
	}

	expectedBody := `{"parameterId":1,"groupCode":"SYSTEM","code":"MAX_LIMIT","value":"100","orderNumber":1}`
	if w.Body.String() != expectedBody {
		t.Errorf("GetParameterById returned unexpected body: got %v want %v", w.Body.String(), expectedBody)
	}

}
