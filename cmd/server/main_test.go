package main

/*
func TestMainHandler(t *testing.T) {
	mux := createServer()

	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/update/counter/test/1", "text/plain", nil)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
*/
