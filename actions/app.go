package actions

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/ssl"
	"github.com/gobuffalo/envy"
	"github.com/unrolled/secure"

	"github.com/devigned/demo/models"
	"github.com/gobuffalo/buffalo/middleware/csrf"
	"github.com/gobuffalo/buffalo/middleware/i18n"
	"github.com/gobuffalo/packr"

	"github.com/markbates/goth/gothic"

	azWorker "github.com/Azure/buffalo-azure/sdk/worker"
	"github.com/gobuffalo/buffalo/worker"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:         ENV,
			SessionName: "_demo_session",
		})
		// Automatically redirect to SSL
		app.Use(ssl.ForceSSL(secure.Options{
			SSLRedirect:     ENV == "production",
			SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
		}))

		if ENV == "development" {
			app.Use(middleware.ParameterLogger)
		}

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		app.Use(csrf.New)

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.PopTransaction)
		// Remove to disable this.
		app.Use(middleware.PopTransaction(models.DB))

		// Setup and use translations:
		var err error
		if T, err = i18n.New(packr.NewBox("../locales"), "en-US"); err != nil {
			app.Stop(err)
		}
		app.Use(T.Middleware())

		app.GET("/", HomeHandler)
		auth := app.Group("/auth")
		auth.GET("/logout", Logout)
		auth.GET("/{provider}", buffalo.WrapHandlerFunc(gothic.BeginAuthHandler))
		auth.GET("/{provider}/callback", AuthCallback)
		app.ServeFiles("/", assetsBox) // serve files from the public directory

		// Setup Buffalo Service Bus Worker
		if err = setupServiceBusWorker(); err != nil {
			app.Stop(err)
		}
		registerWorkers()
	}

	return app
}

func setupServiceBusWorker() error {
	serviceBusWorker, err := azWorker.NewServiceBus(os.Getenv("AZURE_SERVICEBUS_CONN_STR"), 1)
	if err != nil {
		return err
	}
	serviceBusWorker.ServiceBusReceiver.UpsertQueue(context.Background(), "email")
	app.Worker = serviceBusWorker
	return nil
}

func registerWorkers() {
	app.Worker.Register("loginEmail", func(args worker.Args) error {
		return successfulSendEmail(args)
	})
}

func successfulSendEmail(args worker.Args) error {
	if email, ok := args["email"].(string); ok {
		app.Logger.Info("sending email to: ", email)
	}
	return nil
}

func failSendingEmail() error {
	app.Logger.Error("failed sending email...")
	time.Sleep(1 * time.Second)
	return errors.New("broken")
}
