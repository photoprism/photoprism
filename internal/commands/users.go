package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize/english"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// UsersCommand registers user management subcommands.
var UsersCommand = cli.Command{
	Name:  "users",
	Usage: "User management subcommands",
	Subcommands: []cli.Command{
		{
			Name:   "list",
			Usage:  "Lists registered users",
			Action: usersListAction,
		},
		{
			Name:   "add",
			Usage:  "Adds a new user",
			Action: usersAddAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "fullname, n",
					Usage: "full name of the new user",
				},
				cli.StringFlag{
					Name:  "username, u",
					Usage: "unique username",
				},
				cli.StringFlag{
					Name:  "password, p",
					Usage: "sets the users password",
				},
				cli.StringFlag{
					Name:  "email, m",
					Usage: "sets the users email",
				},
			},
		},
		{
			Name:   "update",
			Usage:  "Updates user information",
			Action: usersUpdateAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "fullname, n",
					Usage: "full name of the new user",
				},
				cli.StringFlag{
					Name:  "password, p",
					Usage: "sets the users password",
				},
				cli.StringFlag{
					Name:  "email, m",
					Usage: "sets the users email",
				},
			},
		},
		{
			Name:      "delete",
			Usage:     "Removes an existing user",
			Action:    usersDeleteAction,
			ArgsUsage: "[USERNAME]",
		},
	},
}

func usersAddAction(ctx *cli.Context) error {
	return callWithDependencies(ctx, func(conf *config.Config) error {

		uc := form.UserCreate{
			UserName: strings.TrimSpace(ctx.String("username")),
			FullName: strings.TrimSpace(ctx.String("fullname")),
			Email:    strings.TrimSpace(ctx.String("email")),
			Password: strings.TrimSpace(ctx.String("password")),
		}

		interactive := true

		if uc.UserName != "" && uc.Password != "" {
			log.Debugf("creating user in non-interactive mode")
			interactive = false
		}

		if interactive && uc.FullName == "" {
			prompt := promptui.Prompt{
				Label: "Full Name",
			}
			res, err := prompt.Run()
			if err != nil {
				return err
			}
			uc.FullName = strings.TrimSpace(res)
		}

		if interactive && uc.UserName == "" {
			prompt := promptui.Prompt{
				Label: "Username",
			}
			res, err := prompt.Run()
			if err != nil {
				return err
			}
			uc.UserName = strings.TrimSpace(res)
		}

		if interactive && uc.Email == "" {
			prompt := promptui.Prompt{
				Label: "E-Mail",
			}
			res, err := prompt.Run()
			if err != nil {
				return err
			}
			uc.Email = strings.TrimSpace(res)
		}

		if interactive && len(ctx.String("password")) < 4 {
			validate := func(input string) error {
				if len(input) < 4 {
					return errors.New("password must have min. 4 characters")
				}
				return nil
			}
			prompt := promptui.Prompt{
				Label:    "Password",
				Validate: validate,
				Mask:     '*',
			}
			resPasswd, err := prompt.Run()
			if err != nil {
				return err
			}
			validateRetype := func(input string) error {
				if input != resPasswd {
					return errors.New("password does not match")
				}
				return nil
			}
			confirm := promptui.Prompt{
				Label:    "Retype Password",
				Validate: validateRetype,
				Mask:     '*',
			}
			resConfirm, err := confirm.Run()
			if err != nil {
				return err
			}
			if resConfirm != resPasswd {
				return errors.New("passwords did not match or too short. please try again")
			} else {
				uc.Password = resPasswd
			}
		}

		if err := entity.CreateWithPassword(uc); err != nil {
			return err
		}

		return nil
	})
}

func usersDeleteAction(ctx *cli.Context) error {
	return callWithDependencies(ctx, func(conf *config.Config) error {
		userName := strings.TrimSpace(ctx.Args().First())

		if userName == "" {
			return errors.New("please provide a username")
		}

		actionPrompt := promptui.Prompt{
			Label:     fmt.Sprintf("Delete %s?", sanitize.Log(userName)),
			IsConfirm: true,
		}

		if _, err := actionPrompt.Run(); err == nil {
			if m := entity.FindUserByName(userName); m == nil {
				return errors.New("user not found")
			} else if err := m.Delete(); err != nil {
				return err
			} else {
				log.Infof("%s deleted", sanitize.Log(userName))
			}
		} else {
			log.Infof("keeping user")
		}

		return nil
	})
}

func usersListAction(ctx *cli.Context) error {
	return callWithDependencies(ctx, func(conf *config.Config) error {
		users := query.RegisteredUsers()
		log.Infof("found %s", english.Plural(len(users), "user", "users"))

		fmt.Printf("%-4s %-16s %-16s %-16s\n", "ID", "LOGIN", "NAME", "EMAIL")

		for _, user := range users {
			fmt.Printf("%-4d %-16s %-16s %-16s", user.ID, user.Username(), user.FullName, user.PrimaryEmail)
			fmt.Printf("\n")
		}

		return nil
	})
}

func usersUpdateAction(ctx *cli.Context) error {
	return callWithDependencies(ctx, func(conf *config.Config) error {
		username := ctx.Args().First()
		if username == "" {
			return errors.New("pass username as argument")
		}

		u := entity.FindUserByName(username)
		if u == nil {
			return errors.New("user not found")
		}

		uc := form.UserCreate{
			FullName: strings.TrimSpace(ctx.String("fullname")),
			Email:    strings.TrimSpace(ctx.String("email")),
			Password: strings.TrimSpace(ctx.String("password")),
		}

		if ctx.IsSet("password") {
			err := u.SetPassword(uc.Password)
			if err != nil {
				return err
			}
			fmt.Printf("password successfully changed: %s\n", sanitize.Log(u.Username()))
		}

		if ctx.IsSet("fullname") {
			u.FullName = uc.FullName
		}

		if ctx.IsSet("email") && len(uc.Email) > 0 {
			u.PrimaryEmail = uc.Email
		}

		if err := u.Validate(); err != nil {
			return err
		}

		if err := u.Save(); err != nil {
			return err
		}

		fmt.Printf("user successfully updated: %s\n", sanitize.Log(u.Username()))

		return nil
	})
}

func callWithDependencies(ctx *cli.Context, f func(conf *config.Config) error) error {
	conf := config.NewConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	// Run command.
	if err := f(conf); err != nil {
		return err
	}

	return nil
}
