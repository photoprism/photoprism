/*

Package api provides REST API request handlers.

Copyright (c) 2018 - 2022 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/
package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
)

var log = event.Log

func logError(prefix string, err error) {
	if err != nil {
		log.Errorf("%s: %s", prefix, err.Error())
	}
}

func logWarn(prefix string, err error) {
	if err != nil {
		log.Warnf("%s: %s", prefix, err.Error())
	}
}

func UpdateClientConfig() {
	conf := service.Config()

	event.Publish("config.updated", event.Data{"config": conf.UserConfig()})
}

func Abort(c *gin.Context, code int, id i18n.Message, params ...interface{}) {
	resp := i18n.NewResponse(code, id, params...)

	log.Debugf("api-v1: abort %s with code %d (%s)", clean.Log(c.FullPath()), code, strings.ToLower(resp.String()))

	c.AbortWithStatusJSON(code, resp)
}

func Error(c *gin.Context, code int, err error, id i18n.Message, params ...interface{}) {
	resp := i18n.NewResponse(code, id, params...)

	if err != nil {
		resp.Details = err.Error()
		log.Errorf("api-v1: error %s with code %d in %s (%s)", clean.Log(err.Error()), code, clean.Log(c.FullPath()), strings.ToLower(resp.String()))
	}

	c.AbortWithStatusJSON(code, resp)
}

func AbortUnauthorized(c *gin.Context) {
	Abort(c, http.StatusUnauthorized, i18n.ErrUnauthorized)
}

func AbortEntityNotFound(c *gin.Context) {
	Abort(c, http.StatusNotFound, i18n.ErrEntityNotFound)
}

func AbortSaveFailed(c *gin.Context) {
	Abort(c, http.StatusInternalServerError, i18n.ErrSaveFailed)
}

func AbortDeleteFailed(c *gin.Context) {
	Abort(c, http.StatusInternalServerError, i18n.ErrDeleteFailed)
}

func AbortUnexpected(c *gin.Context) {
	Abort(c, http.StatusInternalServerError, i18n.ErrUnexpected)
}

func AbortBadRequest(c *gin.Context) {
	Abort(c, http.StatusBadRequest, i18n.ErrBadRequest)
}

func AbortAlreadyExists(c *gin.Context, s string) {
	Abort(c, http.StatusConflict, i18n.ErrAlreadyExists, s)
}

func AbortFeatureDisabled(c *gin.Context) {
	Abort(c, http.StatusForbidden, i18n.ErrFeatureDisabled)
}

func AbortBusy(c *gin.Context) {
	Abort(c, http.StatusTooManyRequests, i18n.ErrBusy)
}
