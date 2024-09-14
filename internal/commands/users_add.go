package commands

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// UsersAddCommand configures the command name, flags, and action.
var UsersAddCommand = cli.Command{
	Name:      "add",
	Usage:     "Creates a new user account",
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
			return authn.ErrUsernameRequired
		} else if m := entity.FindUserByName(frm.UserName); m != nil {
			if !m.IsDeleted() {
				return authn.ErrAccountAlreadyExists
			}

			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Restore user %s?", m.String()),
				IsConfirm: true,
			}

			if _, err := prompt.Run(); err != nil {
				return authn.ErrAccountAlreadyExists
			}

			if err := m.RestoreFromCli(ctx, frm.Password); err != nil {
				return err
			}

			log.Infof("user %s has been restored", m.String())

			return nil
		}

		// Enter account email.
		if interactive && frm.UserEmail == "" && frm.Provider().SupportsPasswordAuthentication() {
			prompt := promptui.Prompt{
				Label: "Email",
			}

			res, err := prompt.Run()

			if err != nil {
				return err
			}

			frm.UserEmail = clean.Email(res)
		}

		// Enter account password.
		if interactive && (frm.Provider().RequiresLocalPassword() || frm.Password != "") &&
			len([]rune(ctx.String("password"))) < entity.PasswordLength {
			validate := func(input string) error {
				if len([]rune(input)) < entity.PasswordLength {
					return fmt.Errorf("password must have at least %d characters", entity.PasswordLength)
				} else if len(input) > txt.ClipPassword {
					return authn.ErrPasswordTooLong
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
					return authn.ErrPasswordsDoNotMatch
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
				return authn.ErrInvalidPassword
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
