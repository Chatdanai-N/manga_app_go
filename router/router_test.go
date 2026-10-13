package router

import (
	"manga_app/config"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {

	// 1.Arrange
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

	// load config
	envContent := `DB_HOST=localhost
DB_PORT=5432
DB_USER=root
DB_PASSWORD=secret
DB_NAME=testdb
DB_SEARCH_PATH=public
PORT=8080`
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	if err != nil {

		t.Fatalf("failed to create temp env file: %v", err)
	}

	// ย้าย Working Directory ชั่วคราวไปที่ Directory ที่มีไฟล์ .env
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)
	// 2.Act
	cfg, err := config.LoadConfig()

	// 3.Assert
	if err != nil {
		t.Fatalf("expected no error , got %v", err)
	}

	if cfg == nil {
		t.Fatal("expected cfg not to be nil")
	}

	// 2.Ack
	router := SetupRouter(gormDB, cfg)

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
