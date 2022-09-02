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
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/report"
)

const UsernameUsage = "unique login identifier"
const EmailUsage = "unique email address"
const PasswordUsage = "secure login password"

// UsersCommand registers user management subcommands.
var UsersCommand = cli.Command{
	Name:  "users",
	Usage: "User management subcommands",
	Subcommands: []cli.Command{
		{
			Name:    "ls",
			Aliases: []string{"list"},
			Usage:   "Shows registered users",
			Flags:   report.CliFlags,
			Action:  usersListAction,
		},
		{
			Name:    "add",
			Aliases: []string{"create"},
			Usage:   "Adds a new user account",
			Action:  usersAddAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "username, u",
					Usage: UsernameUsage,
				},
				cli.StringFlag{
					Name:  "email, m",
					Usage: EmailUsage,
				},
				cli.StringFlag{
					Name:  "password, p",
					Usage: PasswordUsage,
				},
			},
		},
		{
			Name:    "mod",
			Aliases: []string{"update"},
			Usage:   "Updates a user account",
			Action:  usersUpdateAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "username, u",
					Usage: UsernameUsage,
				},
				cli.StringFlag{
					Name:  "email, m",
					Usage: EmailUsage,
				},
				cli.StringFlag{
					Name:  "password, p",
					Usage: PasswordUsage,
				},
			},
		},
		{
			Name:      "rm",
			Aliases:   []string{"delete"},
			Usage:     "Removes a user account",
			Action:    usersDeleteAction,
			ArgsUsage: "[username]",
		},
	},
}

func usersAddAction(ctx *cli.Context) error {
	return callWithDependencies(ctx, func(conf *config.Config) error {

		uc := form.UserCreate{
			Username: strings.TrimSpace(ctx.String("username")),
			Email:    strings.TrimSpace(ctx.String("email")),
			Password: strings.TrimSpace(ctx.String("password")),
		}

		interactive := true

		if uc.Username != "" && uc.Password != "" {
			log.Debugf("creating user in non-interactive mode")
			interactive = false
		}

		if interactive && uc.Username == "" {
			prompt := promptui.Prompt{
				Label: "Username",
			}

			res, err := prompt.Run()

			if err != nil {
				return err
			}

			uc.Username = strings.TrimSpace(res)
		}

		if interactive && uc.Email == "" {
			prompt := promptui.Prompt{
				Label: "Email",
			}

			res, err := prompt.Run()

			if err != nil {
				return err
			}

			uc.Email = strings.TrimSpace(res)
		}

		if interactive && len(ctx.String("password")) < entity.PasswordLength {
			validate := func(input string) error {
				if len(input) < entity.PasswordLength {
					return fmt.Errorf("password must have at least %d characters", entity.PasswordLength)
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
		login := strings.TrimSpace(ctx.Args().First())

		if login == "" {
			return errors.New("please provide a username")
		}

		actionPrompt := promptui.Prompt{
			Label:     fmt.Sprintf("Delete %s?", clean.Log(login)),
			IsConfirm: true,
		}

		if _, err := actionPrompt.Run(); err == nil {
			if m := entity.FindUserByLogin(login); m == nil {
				return errors.New("login name not found")
			} else if err := m.Delete(); err != nil {
				return err
			} else {
				log.Infof("%s deleted", clean.LogQuote(login))
			}
		} else {
			log.Infof("keeping user")
		}

		return nil
	})
}

func usersListAction(ctx *cli.Context) error {
	return callWithDependencies(ctx, func(conf *config.Config) error {
		cols := []string{"UID", "Role", "Username", "Email", "Display Name"}

		users := query.RegisteredUsers()
		rows := make([][]string, len(users))

		log.Infof("found %s", english.Plural(len(users), "user", "users"))

		for i, user := range users {
			rows[i] = []string{user.UserUID, user.AclRole().String(), user.UserName(), user.UserEmail(), user.RealName()}
		}

		result, err := report.Render(rows, cols, report.CliFormat(ctx))

		fmt.Println(result)

		return err
	})
}

func usersUpdateAction(ctx *cli.Context) error {
	return callWithDependencies(ctx, func(conf *config.Config) error {
		login := ctx.Args().First()

		if login == "" {
			return errors.New("please provide the username as argument")
		}

		u := entity.FindUserByLogin(login)
		if u == nil {
			return errors.New("username not found")
		}

		uc := form.UserCreate{
			Username: strings.TrimSpace(ctx.String("username")),
			Email:    strings.TrimSpace(ctx.String("email")),
			Password: strings.TrimSpace(ctx.String("password")),
		}

		if ctx.IsSet("email") && len(uc.Email) > 0 {
			u.Email = uc.Email
		}

		if ctx.IsSet("password") {
			err := u.SetPassword(uc.Password)
			if err != nil {
				return err
			}
			fmt.Printf("password successfully changed: %s\n", clean.Log(u.UserName()))
		}

		if err := u.Validate(); err != nil {
			return err
		}

		if err := u.Save(); err != nil {
			return err
		}

		fmt.Printf("user account successfully updated: %s\n", clean.Log(u.UserName()))

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
