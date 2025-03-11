package controllers

const (
	//host = "http://localhost:8080"
)

/*
func TestUpdateCounterOK(t *testing.T) {
	err := database.ResetTestDB()
	require.NoError(t, err)
	url := host + "/update/counter/test/1"
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "counter"},
		{Key: "metric", Value: "test"},
		{Key: "value", Value: "1"},
	}
	c.Request.Header.Set("Content-Type", "text/plain")

	UpdateMetricParams(c)
	UpdateMetricParams(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())

	db := database.GetTestDB()
	var val float64
	err = db.QueryRow(context.Background(), "SELECT value FROM metrics WHERE name = $1", "test").Scan(&val)
	assert.NoError(t, err)
	assert.Equal(t, 2.0, val)
}

func TestUpdateGaugeOK(t *testing.T) {
	s := storage.GetStorage()
	url := host + "/update/gauge/test/2.5"
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "gauge"},
		{Key: "metric", Value: "test"},
		{Key: "value", Value: "2.5"},
	}
	c.Request.Header.Set("Content-Type", "text/plain")

	UpdateMetricParams(c)
	UpdateMetricParams(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	v, exist := s.GetMetric("test")
	assert.True(t, exist)
	assert.Equal(t, 2.5, v)
}
*/

/*
func TestUpdateInvalidPath(t *testing.T) {
	url := host + "/update/counter/test"
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "counter"},
		{Key: "metric", Value: "test"},
	}
	c.Request.Header.Set("Content-Type", "text/plain")

	UpdateMetricParams(c)

	assert.Equal(t, http.StatusNotFound, c.Writer.Status())
}

func TestUpdateInvalidMethod(t *testing.T) {
	url := host + "/update/counter/test/1"

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "counter"},
		{Key: "metric", Value: "test"},
		{Key: "value", Value: "1"},
	}
	c.Request.Header.Set("Content-Type", "text/plain")

	UpdateMetricParams(c)

	assert.Equal(t, http.StatusMethodNotAllowed, c.Writer.Status())
}

func TestUpdateInvalidContentType(t *testing.T) {
	url := host + "/update/counter/test/1"

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "counter"},
		{Key: "metric", Value: "test"},
		{Key: "value", Value: "1"},
	}
	c.Request.Header.Set("Content-Type", "incorrect")

	UpdateMetricParams(c)

	assert.Equal(t, http.StatusMethodNotAllowed, c.Writer.Status())
}

func TestUpdateInvalidType(t *testing.T) {
	url := host + "/update/incorrect/test/1"

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "incorrect"},
		{Key: "metric", Value: "test"},
		{Key: "value", Value: "1"},
	}
	c.Request.Header.Set("Content-Type", "text/plain")

	UpdateMetricParams(c)

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestUpdateInvalidGaugeValue(t *testing.T) {
	url := host + "/update/gauge/test/incorrect"

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "counter"},
		{Key: "metric", Value: "test"},
		{Key: "value", Value: "incorrect"},
	}
	c.Request.Header.Set("Content-Type", "text/plain")

	UpdateMetricParams(c)

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestUpdateInvalidCounterValue(t *testing.T) {
	url := host + "/update/counter/test/2.5"

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", url, nil)
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Params = gin.Params{
		{Key: "type", Value: "counter"},
		{Key: "metric", Value: "test"},
		{Key: "value", Value: "2.5"},
	}
	c.Request.Header.Set("Content-Type", "text/plain")

	UpdateMetricParams(c)

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}
*/
