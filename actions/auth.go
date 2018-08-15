package actions

import (
	"fmt"
	"os"

	"github.com/gobuffalo/buffalo"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/azuread"
	"net/http"
	"github.com/pkg/errors"
	"github.com/gorilla/sessions"
	"github.com/gorilla/securecookie"
	"github.com/gobuffalo/buffalo/worker"
)

func init() {
	gothic.Store = App().SessionStore
	switch store := gothic.Store.(type) {
	case *sessions.CookieStore:
		codec := store.Codecs[0]
		if cookie, ok := codec.(*securecookie.SecureCookie); ok {
			cookie.MaxLength(5120)
		}
	}
	callback := fmt.Sprintf("%s%s", App().Host, "/auth/azuread/callback")
	if App().Host == "http://127.0.0.1:3000" {
		callback = fmt.Sprintf("%s%s", "http://localhost:3000", "/auth/azuread/callback")
	}
	goth.UseProviders(azuread.New(os.Getenv("AZURE_KEY"), os.Getenv("AZURE_SECRET"), callback, nil))
}

func AuthCallback(c buffalo.Context) error {
	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return c.Error(401, err)
	}

	app.Worker.Perform(worker.Job{
		Queue: "email",
		Args: map[string]interface{}{
			"email": user.Email,
		},
		Handler: "loginEmail",
	})

	c.Session().Set("current_user_id", user.Email)
	return c.Redirect(301, "/")
}

func Logout(c buffalo.Context) error {
	err := gothic.Logout(c.Response(), c.Request())
	if err != nil {
		return errors.WithStack(err)
	}
	c.Session().Delete("current_user_id")
	c.Set("current_user", nil)
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
