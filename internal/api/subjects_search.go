package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
)

// SearchSubjects finds and returns subjects as JSON.
//
//	@Summary	finds and returns subjects as JSON
//	@Id			SearchSubjects
//	@Tags		Subjects
//	@Produce	json
//	@Success	200				{object}	search.SubjectResults
//	@Failure	401,429,403,400	{object}	i18n.Response
//	@Param		count			query		int		true	"maximum number of results"	minimum(1)	maximum(100000)
//	@Param		offset			query		int		false	"search result offset"		minimum(0)	maximum(100000)
//	@Param		order			query		string	false	"sort order"				Enums(name, count, added, relevance)
//	@Param		hidden			query		string	false	"show hidden"				Enums(yes, no)
//	@Param		files			query		int		false	"minimum number of files"
//	@Param		q				query		string	false	"search query"
//	@Router		/api/v1/subjects [get]
func SearchSubjects(router *gin.RouterGroup) {
	router.GET("/subjects", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePeople, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		var f form.SearchSubjects

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		result, err := search.Subjects(f)

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		AddCountHeader(c, len(result))
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c, s)

		c.JSON(http.StatusOK, result)
	})
}
