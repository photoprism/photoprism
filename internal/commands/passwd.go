package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// PasswdCommand configures the command name, flags, and action.
var PasswdCommand = cli.Command{
	Name:      "passwd",
	Usage:     "Changes the password of the user specified as argument",
	ArgsUsage: "[username]",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "show, s",
			Usage: "show bcrypt password hash",
		},
	},
	Action: passwdAction,
}

// passwdAction changes the password of the user specified as command argument.
func passwdAction(ctx *cli.Context) error {
	id := clean.Username(ctx.Args().First())

	// Name or UID provided?
	if id == "" {
		return cli.ShowSubcommandHelp(ctx)
	}

	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	// Find user record.
	var m *entity.User

	if rnd.IsUID(id, entity.UserUID) {
		m = entity.FindUserByUID(id)
	} else {
		m = entity.FindUserByName(id)
	}

	if m == nil {
		return fmt.Errorf("user %s not found", clean.LogQuote(id))
	} else if m.Deleted() {
		return fmt.Errorf("user %s has been deleted", clean.LogQuote(id))
	}

	log.Infof("please enter a new password for %s (%d-%d characters)\n", clean.Log(m.Username()), entity.PasswordLength, txt.ClipPassword)

	newPassword := getPassword("New Password: ")

	if len([]rune(newPassword)) < entity.PasswordLength {
		return fmt.Errorf("password must have at least %d characters", entity.PasswordLength)
	} else if len(newPassword) > txt.ClipPassword {
		return fmt.Errorf("password must have less than %d characters", txt.ClipPassword)
	}

	retypePassword := getPassword("Retype Password: ")

	if newPassword != retypePassword {
		return errors.New("passwords did not match, please try again")
	}

	if err = m.SetPassword(newPassword); err != nil {
		return err
	}

	// Show bcrypt password hash?
	if pw := entity.FindPassword(m.UserUID); ctx.Bool("show") && pw != nil {
		log.Infof("password for %s successfully changed to %s\n", clean.Log(m.Username()), pw.Hash)
	} else {
		log.Infof("password for %s successfully changed\n", clean.Log(m.Username()))
	}

	return nil
}

// License: MIT Open Source
// Copyright (c) Joe Linoff 2016
// Go code to prompt for password using only standard packages by utilizing syscall.ForkExec() and syscall.Wait4().
// Correctly resets terminal echo after ^C interrupts.
//
// techEcho() - turns terminal echo on or off.
func termEcho(on bool) {
	// Common settings and variables for both stty calls.
	attrs := syscall.ProcAttr{
		Dir:   "",
		Env:   []string{},
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
		Sys:   nil}
	var ws syscall.WaitStatus
	cmd := "echo"
	if on == false {
		cmd = "-echo"
	}

	// Enable/disable echoing.
	pid, err := syscall.ForkExec(
		"/bin/stty",
		[]string{"stty", cmd},
		&attrs)
	if err != nil {
		panic(err)
	}

	// Wait for the stty process to complete.
	_, err = syscall.Wait4(pid, &ws, 0, nil)
	if err != nil {
		panic(err)
	}
}

// getPassword - Prompt for password.
func getPassword(prompt string) string {
	fmt.Print(prompt)

	// Catch a ^C interrupt.
	// Make sure that we reset term echo before exiting.
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		for range signalChannel {
			fmt.Println("")
			termEcho(true)
			os.Exit(1)
		}
	}()

	// Echo is disabled, now grab the data.
	termEcho(false) // disable terminal echo
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	termEcho(true) // always re-enable terminal echo
	fmt.Println("")
	if err != nil {
		// The terminal has been reset, go ahead and exit.
		fmt.Println("ERROR:", err.Error())
		os.Exit(1)
	}
	return strings.TrimSpace(text)
}
