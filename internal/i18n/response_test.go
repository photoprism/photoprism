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
		assert.Equal(t, "Unexpected error, please try again", resp.Err)
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
			assert.Equal(t, `{"code":200,"success":"Changes successfully saved"}`, string(s))
		}
	})
}
