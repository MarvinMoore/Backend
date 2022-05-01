package main

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) wrap(next http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		//pass httprouter.Params to request context
		ctx := context.WithValue(r.Context(), "params", ps)
		//call next middleware with new context
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (app *application) routes() http.Handler {
	router := httprouter.New()

	secure := alice.New(app.checkToken)

	router.HandlerFunc(http.MethodGet, "/status", app.statusHandler)
	router.HandlerFunc(http.MethodGet, "/v1/Welcome", app.Welcome)
	router.HandlerFunc(http.MethodGet, "/v1/admin/checkadmin", app.checkAdmin)
	router.HandlerFunc(http.MethodPost, "/v1/signup", app.Signup)
	router.HandlerFunc(http.MethodPost, "/v1/signin", app.Signin)
	router.HandlerFunc(http.MethodGet, "/v1/search", app.Search)
	router.HandlerFunc(http.MethodGet, "/v1/get-items/", app.getNewItem)
	router.HandlerFunc(http.MethodGet, "/v1/get-discount/", app.getDiscount)
	router.HandlerFunc(http.MethodGet, "/v1/get-discounts/", app.getAllDiscounts)
	router.HandlerFunc(http.MethodGet, "/v1/order/updateShipping/", app.updateShipping)
	router.HandlerFunc(http.MethodPost, "/v1/admin/updateadmin", app.UpdateAdmin)
	router.HandlerFunc(http.MethodPost, "/v1/getItem", app.getItem)
	router.HandlerFunc(http.MethodGet, "/v1/movie/:id", app.getOneMovie)
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.getAllMovies)
	//router.HandlerFunc(http.MethodGet, "/v1/movies/:genre_id", app.getAllMoviesByGenre)

	router.HandlerFunc(http.MethodGet, "/v1/genres", app.getAllGenres)

	router.POST("/v1/admin/addItem", app.wrap(secure.ThenFunc(app.addItem)))
	//router.HandlerFunc(http.MethodPost, "/v1/admin/addItem", app.addItem)
	router.POST("/v1/admin/deleteItem", app.wrap(secure.ThenFunc(app.deleteItem)))
	router.GET("/v1/getItems", app.wrap(secure.ThenFunc(app.getItems)))
	//router.HandlerFunc(http.MethodPost, "/v1/admin/deleteItem", app.deleteMovie)
	// router.HandlerFunc(http.MethodPost, "/v1/admin/editmovie", app.editMovie)

	//router.GET("/v1/admin/deletemovie/:id", app.wrap(secure.ThenFunc(app.deleteMovie)))
	// router.HandlerFunc(http.MethodGet, "/v1/admin/deletemovie/:id", app.deleteMovie)

	return app.enableCORS(router)
}
