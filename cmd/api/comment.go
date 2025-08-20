package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/yanpavel/social_project/internal/store"
)

type CreateCommentPayload struct {
	Content string `json:"content" validate:"required,max=100"`
}

func (app *application) getCommentByPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	ctx := context.Background()

	comments, err := app.store.Comments.GetByPostID(ctx, post.Id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, comments); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) postCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateCommentPayload
	post := getPostFromCtx(r)

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	comment := &store.Comment{
		PostID:  post.Id,
		UserID:  1,
		Content: payload.Content,
	}

	ctx := context.Background()

	id, err := app.store.Comments.CreateComment(ctx, comment)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, id); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
