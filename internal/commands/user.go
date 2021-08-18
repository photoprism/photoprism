package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/urfave/cli"
)

// UserCommand Create, List, Update and Delete Users.
var UserCommand = cli.Command{
	Name:  "users",
	Usage: "Manage Users from CLI",
	Subcommands: []cli.Command{
		{
			Name:   "add",
			Usage:  "creates a new user. Provide at least username and password",
			Action: userAdd,
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
			Name:   "modify",
			Usage:  "modify a users information.",
			Action: userModify,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "fullname, n",
					Usage: "full name of the new user",
				},
				//cli.StringFlag{
				//	Name:  "username, u",
				//	Usage: "unique username",
				//},
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
			Usage:     "deletes user by username",
			Action:    userDelete,
			ArgsUsage: "takes username as argument",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "execute deletion",
				},
			},
		},
		{
			Name:   "list",
			Usage:  "prints a list of all users",
			Action: userList,
		},
	},
}

func userAdd(ctx *cli.Context) error {
	return withDependencies(ctx, func(conf *config.Config) error {

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
			fmt.Printf("please enter full name: ")
			reader := bufio.NewReader(os.Stdin)
			text, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			uc.FullName = strings.TrimSpace(text)
		}

		if interactive && uc.UserName == "" {
			fmt.Printf("please enter a username: ")
			reader := bufio.NewReader(os.Stdin)
			text, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			uc.UserName = strings.TrimSpace(text)
		}

		if interactive && uc.Email == "" {
			fmt.Printf("please enter email: ")
			reader := bufio.NewReader(os.Stdin)
			text, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			uc.Email = strings.TrimSpace(text)
		}

		if interactive && len(ctx.String("password")) < 4 {
			for {
				fmt.Printf("please enter a new password for %s (at least 4 characters)\n", txt.Quote(uc.UserName))
				pw := getPassword("New password: ")
				if confirm := getPassword("Confirm password: "); confirm == pw {
					uc.Password = pw
					break
				} else {
					log.Infof("passwords did not match or too short. please try again\n")
				}
			}
		}

		if err := entity.CreateWithPassword(uc); err != nil {
			return err
		}
		return nil
	})
}

func userDelete(ctx *cli.Context) error {
	return withDependencies(ctx, func(conf *config.Config) error {
		username := ctx.Args()[0]
		if !ctx.Bool("force") {
			user := entity.FindUserByName(username)
			if user != nil {
				log.Infof("found user %s with uid: %s. Use -f to perform actual deletion\n", user.UserName, user.UserUID)
				return nil
			}
			return errors.New("user not found")
		}
		err := query.DeleteUserByName(username)
		if err != nil {
			log.Errorf("%s\n", err)
			return nil
		}
		log.Infof("sucessfully deleted %s\n", username)
		return nil
	})
}

func userList(ctx *cli.Context) error {
	return withDependencies(ctx, func(conf *config.Config) error {
		users := query.AllUsers()
		fmt.Printf("%-16s %-16s %-16s\n", "Username", "Full Name", "Email")
		fmt.Printf("%-16s %-16s %-16s\n", "--------", "---------", "-----")
		for _, user := range users {
			fmt.Printf("%-16s %-16s %-16s", user.UserName, user.FullName, user.PrimaryEmail)
			fmt.Printf("\n")
		}
		fmt.Printf("total users found: %v\n", len(users))
		return nil
	})
}

func userModify(ctx *cli.Context) error {
	return withDependencies(ctx, func(conf *config.Config) error {
		username := ctx.Args().First()
		if username == "" {
			return errors.New("pass username as argument")
		}

		u := entity.FindUserByName(username)
		if u == nil {
			return errors.New("user not found")
		}

		uc := form.UserCreate{
			UserName: strings.TrimSpace(ctx.String("username")),
			FullName: strings.TrimSpace(ctx.String("fullname")),
			//Email:    strings.TrimSpace(ctx.String("email")),
			Password: strings.TrimSpace(ctx.String("password")),
		}

		if ctx.IsSet("password") {
			err := u.SetPassword(uc.Password)
			if err != nil {
				return err
			}
			fmt.Printf("password successfully changed: %v\n", u.UserName)
		}

		//if ctx.IsSet("username") {
		//	u.UserName = uc.UserName
		//}

		if ctx.IsSet("fullname") {
			u.FullName = uc.FullName
		}

		if ctx.IsSet("email") {
			u.PrimaryEmail = uc.Email
		}

		if err := u.Validate(); err != nil {
			return err
		}
		if err := u.Save(); err != nil {
			return err
		}
		fmt.Printf("user successfully updated: %v\n", u.UserName)

		return nil
	})
}

func withDependencies(ctx *cli.Context, f func(conf *config.Config) error) error {
	conf := config.NewConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	// command is executed here
	if err := f(conf); err != nil {
		return err
	}
	return nil
}
