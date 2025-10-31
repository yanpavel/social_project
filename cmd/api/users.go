package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yanpavel/social_project/internal/store"
)

type userKey string

const userCtx userKey = "user"

// ActivateUser gdoc
//
//	@Summary		Activate a user
//	@Description	Activate a user by token
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			token	path		string	true	"Token"
//	@Success		204		{string}	string	"User activated"
//	@Failure		400		{object}	error	"User with the token found"
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*8)
	defer cancel()

	if err := app.store.Users.Activate(ctx, token); err != nil {
		switch err {
		case store.ErrNotFound:
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userId	path		int	true	"User ID"
//	@Success		200		{object}	store.User
//	@Failure		400		{object}	error	"Invalid parameter"
//	@Failure		404		{object}	error	"User is not found"
//	@Failure		500		{object}	error	"Server error"
//	@Security		ApiKeyAuth
//	@Router			/users/{userId} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*8)
	defer cancel()

	user, err := app.getUser(ctx, userID)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// FollowUser godoc
//
//	@Summary		Follows a user
//	@Description	Follows a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User followed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Failure		409		{object}	error	"Conflict error"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/follow [put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)
	followedId, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	if err := app.store.Followers.FollowUser(ctx, followedId, followerUser.Id); err != nil {
		switch err {
		case store.ErrConflict:
			app.conflictError(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// UnfollowUser gdoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User unfollowed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Failure		409		{object}	error	"Conflict error"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/unfollow [put]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	unfollowerUser := getUserFromContext(r)
	unfollowedId, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	if err := app.store.Followers.UnfollowUser(ctx, unfollowedId, unfollowerUser.Id); err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}
