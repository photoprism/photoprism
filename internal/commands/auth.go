package commands

import (
	"errors"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// AuthCommands registers the API authentication subcommands.
var AuthCommands = &cli.Command{
	Name:    "auth",
	Aliases: []string{"sess"},
	Usage:   "API authentication subcommands",
	Subcommands: []*cli.Command{
		AuthListCommand,
		AuthAddCommand,
		AuthShowCommand,
		AuthRemoveCommand,
		AuthResetCommand,
		AuthJWTCommands,
	},
}

// tokensFlag represents a CLI flag to include tokens in a report.
var tokensFlag = &cli.BoolFlag{
	Name:  "tokens",
	Usage: "show preview and download tokens",
}

// authFindSession finds a session by access token, session ID, or reference ID and returns errors
// with their CLI exit code. The errors never repeat the identifier, since it may be a token.
func authFindSession(id string) (*entity.Session, error) {
	m, err := query.Session(id)

	switch {
	case err == nil:
		return &m, nil
	case errors.Is(err, query.ErrInvalidSessionID):
		return nil, cli.Exit(err, 2)
	case gorm.IsRecordNotFoundError(err):
		return nil, cli.Exit(errors.New("session not found"), 3)
	default:
		return nil, cli.Exit(err, 1)
	}
}

// authSessionLabel describes a session by its reference ID and the fields that let an operator
// recognize it, none of which is a credential.
func authSessionLabel(m *entity.Session) string {
	var details []string

	if m.ClientName != "" {
		if m.IsApplication() {
			details = append(details, "app "+clean.LogQuote(m.ClientName))
		} else {
			details = append(details, "client "+clean.LogQuote(m.ClientName))
		}
	}

	if m.UserName != "" {
		details = append(details, "user "+clean.LogQuote(m.UserName))
	}

	if m.GrantType != "" {
		details = append(details, "grant "+clean.Log(m.GrantType))
	}

	if !m.CreatedAt.IsZero() {
		details = append(details, "created "+m.CreatedAt.UTC().Format("2006-01-02"))
	}

	if m.LastActive > 0 {
		details = append(details, "last active "+txt.UnixTime(m.LastActive))
	}

	if len(details) == 0 {
		return "session " + clean.Log(m.RefID)
	}

	return "session " + clean.Log(m.RefID) + " (" + strings.Join(details, ", ") + ")"
}
