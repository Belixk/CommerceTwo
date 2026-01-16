package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUserHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Invalid id format", func(t *testing.T) {
		h := &UserHandler{}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Params = []gin.Param{{Key: "id", Value: "abc"}}

		h.GetUser(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid is format")
	})
}
