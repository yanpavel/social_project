package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginationFeedQuery struct {
	Limit          int        `json:"limit" validate:"gte=1,lte=20"`
	Offset         int        `json:"offset" validate:"gte=0"`
	Sort           string     `json:"sort" validate:"oneofci=asc desc"`
	Search         string     `json:"search"`
	Tags           []string   `json:"tags"`
	CreatedAtStart *time.Time `json:"createdAtStart"`
	CreatedAtEnd   *time.Time `json:"createdAtEnd"`
}

func (fq PaginationFeedQuery) Parse(r *http.Request) (PaginationFeedQuery, error) {
	qs := r.URL.Query()

	limit := qs.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return fq, nil
		}

		fq.Limit = l
	}

	offset := qs.Get("offset")
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return fq, nil
		}

		fq.Offset = o
	}

	sort := qs.Get("sort")
	if sort != "" {
		fq.Sort = sort
	}

	search := qs.Get("search")
	if sort != "" {
		fq.Search = search
	}

	tags := qs.Get("tags")
	if tags != "" {
		fq.Tags = strings.Split(tags, ",")
	} else {
		fq.Tags = []string{}
	}

	createdAtStart := qs.Get("createdAtStart")
	if createdAtStart != "" {
		fq.CreatedAtStart = timeParse(createdAtStart)
	}

	createdAtEnd := qs.Get("createdAtEnd")
	if createdAtEnd != "" {
		fq.CreatedAtEnd = timeParse(createdAtEnd)
	}

	return fq, nil
}

func timeParse(s string) *time.Time {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return nil
	}

	return &t
}
