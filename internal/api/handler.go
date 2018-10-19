package api

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/photoprism"
)

type Handler struct {
	router *gin.RouterGroup
	conf   *photoprism.Config
}

type Params struct {
	Router *gin.RouterGroup
	Conf   *photoprism.Config
}

func (p *params) validate() error {
	if p.Router == nil {
		return errors.New("empty router group not allowed")
	}
	if p.Conf == nil {
		return errors.New("empty photoprism config not allowed")
	}

	return nil
}

func New(params HandlerParams) (*Handler, error) {
	handler := &Handler{
		router: params.Router,
		conf:   Conf,
	}

	return handler, nil
}
