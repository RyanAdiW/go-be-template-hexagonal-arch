package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/config"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/platform/database"
	"log/slog"
	"os"
	"strconv"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fatal(err)
	}
	m, err := migrate.New("file://migrations", database.DSN(cfg.Database))
	if err != nil {
		fatal(err)
	}
	defer m.Close()
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	switch command {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-1)
	case "version":
		version, dirty, versionErr := m.Version()
		if versionErr != nil && !errors.Is(versionErr, migrate.ErrNilVersion) {
			fatal(versionErr)
		}
		fmt.Printf("version=%d dirty=%t\n", version, dirty)
		return
	case "force":
		if len(os.Args) != 3 {
			fatal(fmt.Errorf("usage: migrate force <version>"))
		}
		version, parseErr := strconv.Atoi(os.Args[2])
		if parseErr != nil {
			fatal(parseErr)
		}
		err = m.Force(version)
	default:
		fatal(fmt.Errorf("unknown command %q", command))
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		fatal(err)
	}
}
func fatal(err error) { slog.Error("migration failed", "error", err); os.Exit(1) }

var _ *sql.DB
