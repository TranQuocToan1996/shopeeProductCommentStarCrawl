package v1

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/TranQuocToan1996/shopeerating/internal/usecase"
	"github.com/TranQuocToan1996/shopeerating/pkg/logger"
	"github.com/gin-gonic/gin"
)

type ratingRoutes struct {
	u usecase.Rating
	l logger.Interface
}

func newRatingRoutes(handler *gin.RouterGroup, u usecase.Rating, l logger.Interface) {
	r := &ratingRoutes{u, l}

	h := handler.Group("/rating")
	{
		h.POST("/csv", r.getOne)
		h.POST("/words", r.words)
	}
}

func (r *ratingRoutes) getOne(c *gin.Context) {
	req := &struct {
		RawURL string   `json:"url"`
		Words  []string `json:"words"`
	}{}
	err := c.ShouldBindJSON(req)
	if len(req.RawURL) == 0 || err != nil {
		errorResponse(c, http.StatusInternalServerError, "URL empty")

		return
	}

	ratingObj, err := r.u.GetRatings(c.Request.Context(), req.RawURL)
	if err != nil {
		r.l.Error(err, "http - v1 - history")
		errorResponse(c, http.StatusInternalServerError, "some problems")

		return
	}

	records := [][]string{{"UserName", "Star", "Comments"}}
	for _, data := range ratingObj.Data.Ratings {
		match := true
		if len(req.Words) > 0 {
			for _, word := range req.Words {
				if !strings.Contains(data.Comment, word) {
					match = false
					break
				}
			}
		}

		if match {
			records = append(records, []string{
				data.AuthorUsername,
				fmt.Sprint(data.RatingStar),
				data.Comment,
			})
		}

	}

	filename := fmt.Sprintf("%v.csv", time.Now().Unix())

	f, err := os.Create(filename)
	if err != nil {
		r.l.Error(err, "http - v1 - history")
		errorResponse(c, http.StatusInternalServerError, "some problems")
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	for _, record := range records {
		if err := w.Write(record); err != nil {
			r.l.Error(err, "http - v1 - history")
			errorResponse(c, http.StatusInternalServerError, "some problems")
		}
	}
	c.FileAttachment(fmt.Sprintf("./%v", filename), filename)
	c.Writer.Header().Set("attachment", fmt.Sprintf("filename=%v", filename))
}

func (r *ratingRoutes) words(c *gin.Context) {
	req := &struct {
		RawURL string `json:"url"`
	}{}
	err := c.ShouldBindJSON(req)
	if len(req.RawURL) == 0 || err != nil {
		errorResponse(c, http.StatusInternalServerError, "URL empty")

		return
	}

	ratingObj, err := r.u.GetRatings(c.Request.Context(), req.RawURL)
	if err != nil {
		r.l.Error(err, "http - v1 - history")
		errorResponse(c, http.StatusInternalServerError, "some problems")

		return
	}

	records := [][]string{{"UserName", "Star", "Comments"}}
	for _, data := range ratingObj.Data.Ratings {
		records = append(records, []string{
			data.AuthorUsername,
			fmt.Sprint(data.RatingStar),
			data.Comment,
		})
	}

	filename := fmt.Sprintf("%v.csv", time.Now().Unix())

	f, err := os.Create(filename)
	if err != nil {
		r.l.Error(err, "http - v1 - history")
		errorResponse(c, http.StatusInternalServerError, "some problems")
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	for _, record := range records {
		if err := w.Write(record); err != nil {
			r.l.Error(err, "http - v1 - history")
			errorResponse(c, http.StatusInternalServerError, "some problems")
		}
	}
	c.FileAttachment(fmt.Sprintf("./%v", filename), filename)
	c.Writer.Header().Set("attachment", fmt.Sprintf("filename=%v", filename))
}
