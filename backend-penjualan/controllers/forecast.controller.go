package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ForecastResponse godoc
// @Description Response for forecast upload (includes historical + predictions)
type ForecastResponse struct {
	Historical []map[string]interface{} `json:"historical"`
    Forecast   []map[string]interface{} `json:"forecast"`
}

type UploadForecastController struct {
	// Nggak butuh DB buat forecast, cuma exec Python
}

func NewForecastController() *UploadForecastController {
	return &UploadForecastController{}
}

// UploadHandler godoc
// @Summary Upload CSV and get forecast
// @Description Upload CSV for sales forecast using Prophet model. CSV must have 'date' (YYYY-MM-DD) and 'projected_quantity' (or 'value') columns.
// @Tags forecast
// @Accept multipart/form-data
// @Produce json
// @Param csvFile formData file true "CSV file with historical data"
// @Param periods formData int false "Number of days to forecast (default 30)"
// @Success 200 {object} controllers.ForecastResponse
// @Failure 400 {object} map[string]string "Invalid CSV or input"
// @Failure 500 {object} map[string]string "Server error (e.g., Python exec failed)"
// @Router /api/v1/forecast/upload [post]
func (ctrl *UploadForecastController) UploadHandler(c *gin.Context) {
	// Parse multipart
	file, err := c.FormFile("csvFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No CSV file provided"})
		return
	}

	// Save ke temp file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("upload_%d.csv", time.Now().UnixNano()))
	out, err := os.Create(tempFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp file"})
		return
	}
	defer out.Close()
	defer os.Remove(tempFile) // Cleanup

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer f.Close()

	_, err = io.Copy(out, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save CSV"})
		return
	}

	// Periods dari form
	periodsStr := c.PostForm("periods")
	periods, _ := strconv.Atoi(periodsStr)
	if periods == 0 {
		periods = 30
	}

	// Input JSON buat Python
	input := map[string]interface{}{
		"csv_path": tempFile,
		"periods":  periods,
	}
	inputJSON, _ := json.Marshal(input)

	// Exec Python script
	cmd := exec.Command("py", "-3.14", "./ml/predict.py")
	cmd.Stdin = bytes.NewBuffer(inputJSON)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// LOG DETAIL KE CONSOLE (untuk kamu lihat di terminal)
		fmt.Printf("=== PYTHON EXECUTION FAILED ===\n")
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Full output (stdout + stderr):\n%s\n", string(output))
		fmt.Printf("================================\n")

		// KIRIM DETAIL ERROR KE CLIENT (Swagger/Postman akan tampilkan ini)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Prediction failed",
			"details": string(output), // Ini yang paling penting: isi error dari Python
			"cmd_err": fmt.Sprintf("%v", err),
		})
		return
	}

	// Jika sukses
	var resp ForecastResponse
	err = json.Unmarshal(output, &resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to parse prediction response",
			"details": err.Error(),
			"raw_output": string(output),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}