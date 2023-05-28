package commands

import (
	"errors"
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
)

// UsersAddCommand configures the command name, flags, and action.
var UsersAddCommand = cli.Command{
	Name:      "add",
	Usage:     "Adds a new user account",
	ArgsUsage: "[username]",
	Flags:     UserFlags,
	Action:    usersAddAction,
}

// usersAddAction adds a new user account.
func usersAddAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		conf.MigrateDb(false, nil)

		frm := form.NewUserFromCli(ctx)

		interactive := true

		if frm.UserName != "" && frm.Password != "" {
			log.Debugf("user will be added in non-interactive mode")
			interactive = false
		}

		if interactive && frm.UserName == "" {
			prompt := promptui.Prompt{
				Label: "Username",
			}

			res, err := prompt.Run()

			if err != nil {
				return err
			}

			frm.UserName = clean.Username(res)
		}

		// Check if account exists but is deleted.
		if frm.UserName == "" {
			return fmt.Errorf("username is required")
		} else if m := entity.FindUserByName(frm.UserName); m != nil {
			if !m.Deleted() {
				return fmt.Errorf("user already exists")
			}

			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Restore user %s?", m.String()),
				IsConfirm: true,
			}

			if _, err := prompt.Run(); err != nil {
				return fmt.Errorf("user already exists")
			}

			if err := m.RestoreFromCli(ctx, frm.Password); err != nil {
				return err
			}

			log.Infof("user %s has been restored", m.String())

			return nil
		}

		if interactive && frm.UserEmail == "" {
			prompt := promptui.Prompt{
				Label: "Email",
			}

			res, err := prompt.Run()

			if err != nil {
				return err
			}

			frm.UserEmail = clean.Email(res)
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
					return errors.New("passwords do not match")
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
				return errors.New("password is invalid, please try again")
			} else {
				frm.Password = resPasswd
			}
		}

		if err := entity.AddUser(frm); err != nil {
			return err
		}

		return nil
	})
}
