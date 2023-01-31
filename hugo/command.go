package hugo

import (
	"github.com/gohugoio/hugo/commands"
	"github.com/pkg/errors"
)

func Run(source string) error {
	response := commands.Execute(
		[]string{"serve", "-p", "34040", "-s", source, "-w"},
	)

	return errors.Wrap(response.Err, "Hugo failed")
}
