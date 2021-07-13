package users

import (
	"encoding/json"
	"fmt"

	"github.com/boreq/guinea"
	authAdapters "github.com/boreq/velo/adapters/auth"
	"github.com/boreq/velo/application/auth"
	authDomain "github.com/boreq/velo/domain/auth"
	"github.com/boreq/velo/internal/config"
	"github.com/boreq/velo/internal/wire"
	"github.com/pkg/errors"
)

var UsersCmd = guinea.Command{
	Run: runUsers,
	Subcommands: map[string]*guinea.Command{
		"list":           &listCmd,
		"reset_password": &resetPasswordCmd,
	},
	ShortDescription: "manage users",
}

func runUsers(c guinea.Context) error {
	return guinea.ErrInvalidParms
}

var listCmd = guinea.Command{
	Run: runList,
	Arguments: []guinea.Argument{
		{
			Name:        "data_directory",
			Optional:    false,
			Multiple:    false,
			Description: "Path to the directory used for data storage",
		},
	},
	ShortDescription: "list all users",
}

func runList(c guinea.Context) error {
	conf := config.Default()
	conf.DataDirectory = c.Arguments[0]

	auth, err := wire.BuildAuth(conf)
	if err != nil {
		return errors.Wrap(err, "failed to build the application")
	}

	users, err := auth.List.Execute()
	if err != nil {
		return errors.Wrap(err, "failed to list users")
	}

	j, err := json.MarshalIndent(users, "", "    ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal to json")
	}

	fmt.Println(string(j))

	return nil
}

var resetPasswordCmd = guinea.Command{
	Run: runResetPassword,
	Arguments: []guinea.Argument{
		{
			Name:        "data_directory",
			Optional:    false,
			Multiple:    false,
			Description: "Path to the directory used for data storage",
		},
		{
			Name:        "username",
			Optional:    false,
			Multiple:    false,
			Description: "Username",
		},
	}, ShortDescription: "resets a user's password",
}

func runResetPassword(c guinea.Context) error {
	conf := config.Default()
	conf.DataDirectory = c.Arguments[0]

	a, err := wire.BuildAuth(conf)
	if err != nil {
		return errors.Wrap(err, "failed to build the application")
	}

	generator := authAdapters.NewCryptoStringGenerator()
	s, err := generator.Generate(256 / 8)
	if err != nil {
		return errors.Wrap(err, "failed to generate a secure string")
	}

	password, err := authDomain.NewPassword(s)
	if err != nil {
		return errors.Wrap(err, "invalid password")
	}

	cmd := auth.SetPassword{
		Username: c.Arguments[1],
		Password: password,
	}

	err = a.SetPassword.Execute(cmd)
	if err != nil {
		return errors.Wrap(err, "failed to set a password")
	}

	fmt.Println(cmd.Password.String())

	return nil
}
