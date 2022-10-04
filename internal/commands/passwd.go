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
)

// PasswdCommand updates a password.
var PasswdCommand = cli.Command{
	Name:   "passwd",
	Usage:  "Changes the admin account password",
	Action: passwdAction,
}

// passwdAction updates a password.
func passwdAction(ctx *cli.Context) error {
	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	user := entity.Admin

	log.Infof("please enter a new password for %s (mininum 8 characters)\n", clean.Log(user.Name()))

	newPassword := getPassword("New Password: ")

	if len(newPassword) < 6 {
		return errors.New("new password is too short, please try again")
	}

	retypePassword := getPassword("Retype Password: ")

	if newPassword != retypePassword {
		return errors.New("passwords did not match, please try again")
	}

	if err := user.SetPassword(newPassword); err != nil {
		return err
	}

	log.Infof("changed password for %s\n", clean.Log(user.Name()))

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
