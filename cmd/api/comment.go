package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yanpavel/social_project/internal/store"
)

type commentKey string

const commentCtx commentKey = "comment"

type CommentPayload struct {
	Content string `json:"content" validate:"required,max=100"`
}

// GetComments godoc
//
//	@Summary		Get comments
//	@Description	Get comments
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int	true	"Post ID"
//	@Success		200		{object}	store.Comment
//	@Failure		400		{object}	app.badRequestError
//	@Failure		401		{object}	app.unauthorizedErrorResponse
//	@Failure		500		{object}	app.internalServerError
//	@Security		ApiKeyAuth
//	@Router			/posts/{id}/comment [get]
func (app *application) getCommentByPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	ctx, cancel := context.WithTimeout(r.Context(), QueryTimeOutDuration)
	defer cancel()

	comments, err := app.store.Comments.GetByPostID(ctx, post.Id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, comments); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// CreateComment godoc
//
//	@Summary		Creates a comment
//	@Description	Creates a comment
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateCommentPayload	true	"Comment payload"
//	@Param			postID	path		int						true	"Post ID"
//	@Success		201		{object}	store.Comment.ID
//	@Failure		400		{object}	app.badRequestError
//	@Failure		401		{object}	app.unauthorizedErrorResponse
//	@Failure		500		{object}	app.internalServerError
//	@Security		ApiKeyAuth
//	@Router			/posts/{postId}/comment [post]
func (app *application) postCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload CommentPayload
	post := getPostFromCtx(r)
	user := getUserFromContext(r)

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
		UserID:  user.Id,
		Content: payload.Content,
	}

	ctx, cancel := context.WithTimeout(r.Context(), QueryTimeOutDuration)
	defer cancel()

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

// UpdateComment godoc
//
//		@Summary		Updates the comment
//		@Description	Updates the comment
//		@Tags			comments
//		@Accept			json
//		@Produce		json
//		@Param			postID	path		int						true	"Post ID"
//	 	@Param			commentID	path		int						true	"Comment ID"
//		@Success		201		{object}	store.Comment.ID
//		@Failure		400		{object}	app.badRequestError
//		@Failure		401		{object}	app.unauthorizedErrorResponse
//		@Failure		500		{object}	app.internalServerError
//		@Security		ApiKeyAuth
//		@Router			/posts/{postId}/comment/{commentId}/ [put]
func (app *application) updateCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload CommentPayload
	comment := getCommentFromContext(r)

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	comment = &store.Comment{Content: payload.Content}

	ctx, cancel := context.WithTimeout(r.Context(), QueryTimeOutDuration)
	defer cancel()

	err := app.store.Comments.UpdateCommentById(ctx, comment)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusAccepted, comment.ID); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// DeleteComment godoc
//
//		@Summary		Deletes the comment
//		@Description	Deletes the comment
//		@Tags			comments
//		@Accept			json
//		@Produce		json
//		@Param			postID	path		int						true	"Post ID"
//	 	@Param			commentID	path		int						true	"Comment ID"
//		@Success		201		{object}	store.Comment.ID
//		@Failure		400		{object}	app.badRequestError
//		@Failure		401		{object}	app.unauthorizedErrorResponse
//		@Failure		500		{object}	app.internalServerError
//		@Security		ApiKeyAuth
//		@Router			/posts/{postId}/comment/{commentId}/ [delete]
func (app *application) deleteCommentByIdHandler(w http.ResponseWriter, r *http.Request) {
	comment := getCommentFromContext(r)

	ctx, cancel := context.WithTimeout(r.Context(), QueryTimeOutDuration)
	defer cancel()

	err := app.store.Comments.DeleteCommentById(ctx, comment.ID)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, comment.ID); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) commentContextMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "commentID")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), QueryTimeOutDuration)
		defer cancel()

		comment, err := app.store.Comments.GetCommentById(ctx, id)
		if err != nil {
			switch err {
			case store.ErrNotFound:
				app.notFoundError(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, commentCtx, comment)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getCommentFromContext(r *http.Request) *store.Comment {
	comment, _ := r.Context().Value(commentCtx).(*store.Comment)
	return comment
}
