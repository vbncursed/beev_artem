// Command `admin` is the operational CLI for the auth service. It bypasses
// the gRPC layer and writes directly through the persistence adapter so
// operators can bootstrap the first admin (chicken-and-egg: only an admin
// can promote another admin via the RPC, but the platform ships without
// one).
//
// Usage:
//
//	admin promote --email=<email>     promote user to admin
//	admin demote  --email=<email>     demote user back to "user"
//
// Reuses auth's config + bootstrap.InitPGStorage so the connection string
// and SSL mode follow the same APP_ENV / configPath resolution as the
// running service. Run inside the auth container:
//
//	docker exec hr-auth admin promote --email=user@example.com
//
// (or `make admin-promote EMAIL=...` from the repo root.)
package main

import (
	"cmp"
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/artem13815/hr/auth/config"
	"github.com/artem13815/hr/auth/internal/bootstrap"
	"github.com/artem13815/hr/auth/internal/domain"
	"github.com/artem13815/hr/auth/internal/infrastructure/auth_storage"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	sub := os.Args[1]
	args := os.Args[2:]

	var err error
	switch sub {
	case "promote":
		err = setRole(args, domain.RoleAdmin)
	case "demote":
		err = setRole(args, domain.RoleUser)
	case "-h", "--help", "help":
		usage()
		return
	default:
		usage()
		os.Exit(2)
	}

	if err != nil {
		slog.Error(sub+" failed", "err", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `auth admin CLI — manage user roles directly via DB.

Subcommands:
  promote --email=<email>   set role = "admin"
  demote  --email=<email>   set role = "user"

Reads config the same way auth-service does:
  configPath env > APP_ENV=prod → config.docker.prod.yaml > dev fallback.

Run inside the running auth container:
  docker exec hr-auth admin promote --email=you@example.com

To inspect users, use psql directly:
  docker exec hr-postgres psql -U admin -d hr -c \
    "SELECT id, email, role FROM auth_users ORDER BY id;"`)
}

func setRole(args []string, role string) error {
	fs := flag.NewFlagSet("setRole", flag.ContinueOnError)
	email := fs.String("email", "", "user email to update")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*email) == "" {
		return errors.New("--email is required")
	}

	storage, cleanup, err := openStorage()
	if err != nil {
		return err
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := storage.GetUserByEmail(ctx, *email)
	if err != nil {
		return fmt.Errorf("lookup user %q: %w", *email, err)
	}

	if user.Role == role {
		fmt.Printf("user %d (%s) already has role=%s — no-op\n",
			user.ID, user.Email, role)
		return nil
	}

	if err := storage.UpdateUserRole(ctx, user.ID, role); err != nil {
		return fmt.Errorf("update role: %w", err)
	}

	fmt.Printf("✓ user %d (%s): %s → %s\n",
		user.ID, user.Email, user.Role, role)
	fmt.Println("note: existing JWTs still carry the old role — sign out and back in to refresh.")
	return nil
}

func openStorage() (*auth_storage.AuthStorage, func(), error) {
	configPath := cmp.Or(
		os.Getenv("configPath"),
		defaultConfigPathByEnv(os.Getenv("APP_ENV")),
	)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("load config %q: %w", configPath, err)
	}
	storage, err := bootstrap.InitPGStorage(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("init storage: %w", err)
	}
	return storage, func() { storage.Close() }, nil
}

func defaultConfigPathByEnv(env string) string {
	switch env {
	case "prod", "production":
		return "config.docker.prod.yaml"
	default:
		return "config.docker.dev.yaml"
	}
}
