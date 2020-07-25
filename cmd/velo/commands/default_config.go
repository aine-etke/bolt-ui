package commands

import (
	"encoding/json"
	"fmt"

	"github.com/boreq/errors"
	"github.com/boreq/guinea"
	"github.com/boreq/velo/internal/config"
)

var defaultConfigCmd = guinea.Command{
	Run:              runDefaultConfig,
	ShortDescription: "prints default config to stdout",
}

func runDefaultConfig(c guinea.Context) error {
	conf := config.Default()

	j, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		return errors.Wrap(err, "could not marshal the configuration")
	}

	fmt.Println(string(j))

	return nil
}
