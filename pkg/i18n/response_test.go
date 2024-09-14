package i18n

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResponse(t *testing.T) {
	t.Run("already exists", func(t *testing.T) {
		resp := NewResponse(http.StatusConflict, ErrAlreadyExists, "A cat")
		assert.Equal(t, http.StatusConflict, resp.Code)
		assert.Equal(t, "A cat already exists", resp.Err)
		assert.Equal(t, "", resp.Msg)
	})

	t.Run("unexpected error", func(t *testing.T) {
		resp := NewResponse(http.StatusInternalServerError, ErrUnexpected, "A cat")
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, "Something went wrong, try again", resp.Err)
		assert.Equal(t, "", resp.Msg)
	})

	t.Run("changes saved", func(t *testing.T) {
		resp := NewResponse(http.StatusOK, MsgChangesSaved)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "", resp.Err)
		assert.Equal(t, "Changes successfully saved", resp.Msg)

		if s, err := json.Marshal(resp); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, `{"code":200,"message":"Changes successfully saved"}`, string(s))
		}
	})
}

func TestResponse_String(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		resp := Response{404, "Not found", "page not found", "xyz"}
		assert.Equal(t, "Not found", resp.String())
	})

	t.Run("no error", func(t *testing.T) {
		t.Run("error", func(t *testing.T) {
			resp := Response{200, "", "Ok", "xyz"}
			assert.Equal(t, "Ok", resp.String())
		})
	})
}

func TestResponse_LowerString(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		resp := Response{404, "Not found", "page not found", "xyz"}
		assert.Equal(t, "not found", resp.LowerString())
	})

	t.Run("no error", func(t *testing.T) {
		t.Run("error", func(t *testing.T) {
			resp := Response{200, "", "Ok", "xyz"}
			assert.Equal(t, "ok", resp.LowerString())
		})
	})
}

func TestResponse_Error(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		resp := Response{404, "Not found", "page not found", "xyz"}
		assert.Equal(t, "Not found", resp.Error())
	})

	t.Run("no error", func(t *testing.T) {
		t.Run("error", func(t *testing.T) {
			resp := Response{200, "", "Ok", "xyz"}
			assert.Equal(t, "", resp.Error())
		})
	})
}

func TestResponse_Success(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		resp := Response{404, "Not found", "page not found", "xyz"}
		assert.Equal(t, false, resp.Success())
	})

	t.Run("no error", func(t *testing.T) {
		t.Run("error", func(t *testing.T) {
			resp := Response{200, "", "Ok", "xyz"}
			assert.Equal(t, true, resp.Success())
		})
	})
}
