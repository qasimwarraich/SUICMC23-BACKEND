package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/mail"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/mailer"
)

func main() {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		_, err := e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/api/race-numbers",
			Handler: func(c echo.Context) error {
				q := app.DB().NewQuery("SELECT race_number FROM participants")
				s := make([]int, 0)
				err := q.Column(&s)
				if err != nil {
					return err
				}

				return c.JSON(http.StatusOK, s)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		_, err = e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/api/racer-names",
			Handler: func(c echo.Context) error {
				q := app.DB().NewQuery("SELECT first_name FROM participants")
				s := make([]string, 0)
				err = q.Column(&s)
				if err != nil {
					return err
				}

				return c.JSON(http.StatusOK, s)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})
		if err != nil {
			return err
		}
		return nil
	})

	app.OnRecordAfterCreateRequest("participants").Add(func(e *core.RecordCreateEvent) error {
		err := godotenv.Load()
		if err != nil {
			fmt.Println("Error loading .env file")
		}

		nickName := e.Record.GetString("nick_name")
		if nickName == "" {
			nickName = e.Record.GetString("first_name")
		}

		templateData := struct {
			Name  string
			URL   string
			TWINT string
			Logo  string
		}{
			Name:  nickName,
			URL:   os.Getenv("PAYMENT_URL"),
			TWINT: os.Getenv("TWINT_IMAGE"),
			Logo:  os.Getenv("LOGO_IMAGE"),
		}

		tmpl, err := template.ParseFiles("assets/registration_email_template.html")
		if err != nil {
			fmt.Println(err)
		}

		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, templateData)
		if err != nil {
			fmt.Println(err)
		}

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
