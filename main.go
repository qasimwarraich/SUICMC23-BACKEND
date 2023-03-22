package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/mail"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/mailer"
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

	app.OnRecordAfterCreateRequest("participants").Add(func(e *core.RecordCreateEvent) error {
		templateData := struct {
			Name  string
			URL   string
			TWINT string
			Logo  string
		}{
			Name:  e.Record.GetString("nick_name"),
			URL:   "https://suicmc23.vercel.app/cream",
			TWINT: "https://suicmc23.vercel.app/_app/immutable/assets/reg_twint_sm-72c3cf7b.png",
			Logo:  "https://suicmc23.vercel.app/_app/immutable/assets/warnwest-721abf10.png",
		}

		tmpl, err := template.ParseFiles("assets/registration_email_template.html")
		if err != nil {
			fmt.Println(err)
		}

		buf := new(bytes.Buffer)
		tmpl.Execute(buf, templateData)

		message := &mailer.Message{
			From: mail.Address{
				Address: app.Settings().Meta.SenderAddress,
				Name:    app.Settings().Meta.SenderName,
			},
			To:      mail.Address{Address: "example@example.com"},
			Subject: "Thank you for registering for SUICMC23 BERN",
			HTML:    buf.String(),
		}

		return app.NewMailClient().Send(message)
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
