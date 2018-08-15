package grifts

import (
	"github.com/devigned/buffalo-service-bus-worker/actions"
	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
