package valkyrie

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	ConfigOpts struct {
		Config    interface{}
		Paths     []string
		Filenames []string
	}
)

func Config(opts ConfigOpts) error {
	for _, p := range opts.Paths {
		filePath := fmt.Sprintf("%s/%s", p, ".env")
		// Set ENV for development
		_ = cleanenv.ReadConfig(fmt.Sprintf("%s/%s", p, filePath), opts.Config)
	}
	var err error
	for _, f := range opts.Filenames {
		for _, p := range opts.Paths {
			filePath := fmt.Sprintf("%s/%s", p, f)
			if err = cleanenv.ReadConfig(filePath, opts.Config); err != nil {
				break
			}
		}
	}

	return err
}
