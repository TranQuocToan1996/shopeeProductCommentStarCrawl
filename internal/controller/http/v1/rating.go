package v1

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/TranQuocToan1996/shopeerating/internal/entity"
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
		h.POST("/csv", r.getAll)
		h.POST("/bylimitoffset", r.byLimitAndSkip)
	}
}

func (r *ratingRoutes) getAll(c *gin.Context) {
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

	records := r.makeCSVRow(req.Words, ratingObj.Data.Ratings)

	filename := fmt.Sprintf("%v.csv", time.Now().Unix())

	r.writeAllCSV(c, filename, records)
	c.FileAttachment(fmt.Sprintf("./%v", filename), filename)
	c.Writer.Header().Set("attachment", fmt.Sprintf("filename=%v", filename))
}

func (r *ratingRoutes) writeCSV(c *gin.Context, filename string, records [][]string) {
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
}

func (r *ratingRoutes) writeAllCSV(c *gin.Context, filename string, records [][]string) {
	f, err := os.Create(filename)
	if err != nil {
		r.l.Error(err, "http - v1 - history")
		errorResponse(c, http.StatusInternalServerError, "some problems")
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.WriteAll(records)
}

func (r *ratingRoutes) byLimitAndSkip(c *gin.Context) {
	req := &struct {
		RawURL string   `json:"url"`
		Limit  int      `json:"limit"`
		Offset int      `json:"offset"`
		Words  []string `json:"words"`
	}{}
	err := c.ShouldBindJSON(req)
	if len(req.RawURL) == 0 || err != nil {
		errorResponse(c, http.StatusInternalServerError, "URL empty")

		return
	}

	ratingObj, err := r.u.GetRatingsLimitSkip(c.Request.Context(), req.RawURL, req.Limit, req.Offset)
	if err != nil {
		r.l.Error(err, "http - v1 - history")
		errorResponse(c, http.StatusInternalServerError, "some problems")

		return
	}

	records := r.makeCSVRow(req.Words, ratingObj.Data.Ratings)

	filename := fmt.Sprintf("%v.csv", time.Now().Unix())
	r.writeCSV(c, filename, records)
	c.FileAttachment(fmt.Sprintf("./%v", filename), filename)
	c.Writer.Header().Set("attachment", fmt.Sprintf("filename=%v", filename))
}

func (r *ratingRoutes) makeCSVRow(words []string, ratings []entity.Ratings) [][]string {
	records := [][]string{{"UserName", "Star", "Comments"}}
	for _, data := range ratings {
		match := false
		if len(words) > 0 {
			for _, word := range words {
				if strings.Contains(data.Comment, word) {
					match = true
					break
				}
			}
		} else {
			match = true
		}

		if match {
			records = append(records, []string{
				data.AuthorUsername,
				fmt.Sprint(data.RatingStar),
				data.Comment,
			})
		}
	}

	return records
}
