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

		emailAddress := e.Record.GetString("email")
		raceNumber := e.Record.GetInt("race_number")
		intendedPayment := e.Record.GetInt("intended_payment")
		paymentMethod := e.Record.GetString("payment_method")

		templateData := struct {
			Name            string
			Email           string
			RaceNumber      int
			URL             string
			TWINT           string
			Logo            string
			PaymentMethod   string
			IntendedPayment int
			RegEmail        string
		}{
			Name:            nickName,
			Email:           emailAddress,
			RaceNumber:      raceNumber,
			URL:             os.Getenv("PAYMENT_URL"),
			TWINT:           os.Getenv("TWINT_IMAGE"),
			Logo:            os.Getenv("LOGO_IMAGE"),
			PaymentMethod:   paymentMethod,
			IntendedPayment: intendedPayment,
			RegEmail:        os.Getenv("REG_EMAIL"),
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
			From:    mail.Address{Name: app.Settings().Meta.SenderName, Address: app.Settings().Meta.SenderAddress},
			To:      []mail.Address{{Address: emailAddress}},
			Bcc:     []mail.Address{},
			Cc:      []mail.Address{},
			Subject: "Thank you for registering for SUICMC23 BERN",
			HTML:    buf.String(),
			Text:    buf.String(),
			Headers: map[string]string{},
		}

		return app.NewMailClient().Send(message)
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
