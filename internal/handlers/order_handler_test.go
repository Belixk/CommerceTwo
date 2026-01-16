package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestOrderHandler_CreateOrder_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &OrderHandler{service: nil}

	r := gin.Default()
	r.POST("/orders", h.CreateOrder)

	invalidJSON := []byte(`{"user_id": 1, "items": [ {"name": "item1" } ] }`)

	req, _ := http.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}
