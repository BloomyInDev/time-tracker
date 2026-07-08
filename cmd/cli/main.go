package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/bloomyindev/time-tracker/internal/config"
	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/service/auth"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "cli",
		Usage: "manage time-tracker data directly against the db",
		Commands: []*cli.Command{
			registerCommand(),
			exportUsersCommand(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
}

func registerCommand() *cli.Command {
	return &cli.Command{
		Name:  "register",
		Usage: "create a new user",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "email", Required: true},
			&cli.StringFlag{Name: "password", Required: true},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			conn, err := openDB()
			if err != nil {
				return err
			}
			defer conn.Close()

			svc := auth.NewService(conn, config.Load().JWTSecret)
			email := cmd.String("email")
			if err := svc.Register(email, cmd.String("password")); err != nil {
				return fmt.Errorf("register: %w", err)
			}
			fmt.Printf("user %s created\n", email)
			return nil
		},
	}
}

func exportUsersCommand() *cli.Command {
	return &cli.Command{
		Name:  "export-users",
		Usage: "export users as JSON (excludes password hashes)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			conn, err := openDB()
			if err != nil {
				return err
			}
			defer conn.Close()

			users, err := db.ListUsers(conn)
			if err != nil {
				return fmt.Errorf("list users: %w", err)
			}

			type userExport struct {
				ID    int64  `json:"id"`
				Email string `json:"email"`
			}
			out := make([]userExport, len(users))
			for i, u := range users {
				out[i] = userExport{ID: u.ID, Email: u.Email}
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(out)
		},
	}
}

func openDB() (*sql.DB, error) {
	return db.Open(config.Load().DBPath)
}
