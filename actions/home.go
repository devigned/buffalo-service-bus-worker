package actions

import (
	"github.com/gobuffalo/buffalo"
	"fmt"
)

// HomeHandler is a default handler to serve up
// a home page.
func HomeHandler(c buffalo.Context) error {
	for i := 0; i < 30; i++ {
		fmt.Println()
	}
	userID := c.Session().Get("current_user_id")
	c.Set("current_user", userID)
	return c.Render(200, r.HTML("index.html"))
}
