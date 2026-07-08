package i18n

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResponse(t *testing.T) {
	t.Run("AlreadyExists", func(t *testing.T) {
		resp := NewResponse(http.StatusConflict, ErrAlreadyExists, "A cat")
		assert.Equal(t, http.StatusConflict, resp.Code)
		assert.Equal(t, "A cat already exists", resp.Error)
		assert.Equal(t, "", resp.Message)
		assert.Equal(t, "%s already exists", resp.MessageID)
		assert.Equal(t, []any{"A cat"}, resp.MessageParams)
	})
	t.Run("UnexpectedError", func(t *testing.T) {
		resp := NewResponse(http.StatusInternalServerError, ErrUnexpected, "A cat")
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, "Something went wrong, try again", resp.Error)
		assert.Equal(t, "", resp.Message)
		assert.Equal(t, "Something went wrong, try again", resp.MessageID)
	})
	t.Run("ChangesSaved", func(t *testing.T) {
		resp := NewResponse(http.StatusOK, MsgChangesSaved)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "", resp.Error)
		assert.Equal(t, "Changes successfully saved", resp.Message)
		assert.Equal(t, "Changes successfully saved", resp.MessageID)

		if s, err := json.Marshal(resp); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, `{"code":200,"message":"Changes successfully saved","messageId":"Changes successfully saved"}`, string(s))
		}
	})
}

func TestResponse_String(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		resp := Response{Code: 404, Error: "Not found", Message: "page not found", Details: "xyz"}
		assert.Equal(t, "Not found", resp.String())
	})
	t.Run("NoError", func(t *testing.T) {
		t.Run("Error", func(t *testing.T) {
			resp := Response{Code: 200, Message: "Ok", Details: "xyz"}
			assert.Equal(t, "Ok", resp.String())
		})
	})
}

func TestResponse_LowerString(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		resp := Response{Code: 404, Error: "Not found", Message: "page not found", Details: "xyz"}
		assert.Equal(t, "not found", resp.LowerString())
	})
	t.Run("NoError", func(t *testing.T) {
		t.Run("Error", func(t *testing.T) {
			resp := Response{Code: 200, Message: "Ok", Details: "xyz"}
			assert.Equal(t, "ok", resp.LowerString())
		})
	})
}

func TestResponse_ErrorString(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		resp := Response{Code: 404, Error: "Not found", Message: "page not found", Details: "xyz"}
		assert.Equal(t, "Not found", resp.ErrorString())
	})
	t.Run("NoError", func(t *testing.T) {
		t.Run("Error", func(t *testing.T) {
			resp := Response{Code: 200, Message: "Ok", Details: "xyz"}
			assert.Equal(t, "", resp.ErrorString())
		})
	})
}

func TestResponse_Success(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		resp := Response{Code: 404, Error: "Not found", Message: "page not found", Details: "xyz"}
		assert.Equal(t, false, resp.Success())
	})
	t.Run("NoError", func(t *testing.T) {
		t.Run("Error", func(t *testing.T) {
			resp := Response{Code: 200, Message: "Ok", Details: "xyz"}
			assert.Equal(t, true, resp.Success())
		})
	})
}
