package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
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
	// Config buat FastAPI URL (dari env atau hardcode)
	FastAPIURL string
}

func NewForecastController() *UploadForecastController {
	fastAPIURL := os.Getenv("ML_SERVICE_URL")
	if fastAPIURL == "" {
		fastAPIURL = "http://localhost:8000/predict" // Default
	}
	return &UploadForecastController{
		FastAPIURL: fastAPIURL,
	}
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
// @Failure 500 {object} map[string]string "Server error (e.g., ML service failed)"
// @Router /api/v1/forecast/upload [post]
func (ctrl *UploadForecastController) UploadHandler(c *gin.Context) {
	// Parse multipart
	file, err := c.FormFile("csvFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No CSV file provided"})
		return
	}

	// Periods dari form
	periodsStr := c.PostForm("periods")
	periods, _ := strconv.Atoi(periodsStr)
	if periods == 0 {
		periods = 30
	}

	// Buat multipart form buat FastAPI
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("csv_file", file.Filename) // FastAPI expect "csv_file"
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create form"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer f.Close()
	io.Copy(part, f)

	// Tambah periods
	writer.WriteField("periods", strconv.Itoa(periods))
	writer.Close()

	// HTTP POST ke FastAPI
	req, err := http.NewRequest("POST", ctrl.FastAPIURL, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "ML service unreachable",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Prediction failed",
			"details": string(bodyBytes),
			"status":  resp.StatusCode,
		})
		return
	}

	// Parse response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	var forecastResp ForecastResponse
	err = json.Unmarshal(bodyBytes, &forecastResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Failed to parse prediction",
			"details":   err.Error(),
			"raw":       string(bodyBytes),
		})
		return
	}

	c.JSON(http.StatusOK, forecastResp)
}