package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/api/race-numbers",
			Handler: func(c echo.Context) error {
				q := app.DB().NewQuery("SELECT race_number FROM participants")
				s := make([]int, 0)
				q.Column(&s)

				return c.JSON(http.StatusOK, s)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/api/racer-names",
			Handler: func(c echo.Context) error {
				q := app.DB().NewQuery("SELECT first_name FROM participants")
				s := make([]string, 0)
				q.Column(&s)

				return c.JSON(http.StatusOK, s)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		return nil
	})
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
