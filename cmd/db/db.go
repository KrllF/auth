package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	desc "github.com/KrllF/auth/pkg/auth_v1"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var DB *pgxpool.Pool

func CreateUser(ctx context.Context, name string, email string, password string, role desc.Role) (*desc.CreateResponse, error) {
	var role_to_db string
	if role == desc.Role_Admin {
		role_to_db = "Admin"
	} else {
		role_to_db = "User"
	}

	hashpass, err := hashPassword(password)
	if err != nil {
		log.Printf("bad password: %v", err)
		return nil, err
	}

	builderInsert := sq.Insert("users").
		PlaceholderFormat(sq.Dollar).
		Columns("name", "email", "password", "role").
		Values(name, email, hashpass, role_to_db).
		Suffix("RETURNING id")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v", err)
		return nil, err
	}

	var userID int
	err = DB.QueryRow(ctx, query, args...).Scan(&userID)
	if err != nil {
		log.Printf("failed to insert user: %v", err)
		return nil, err
	}

	log.Printf("inserted user with id: %d", int64(userID))
	return &desc.CreateResponse{Id: int64(userID)}, nil
}

func GetUser(ctx context.Context, userid int64) (*desc.GetResponse, error) {
	builderSelect := sq.Select("id", "name", "email", "role", "created_at", "updated_at").
		From("users").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": userid}).
		Limit(1)

	query, args, err := builderSelect.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v", err)
		return nil, err
	}

	row := DB.QueryRow(ctx, query, args...)

	var id int64
	var name, email string
	var role string
	var createdAt time.Time
	var updatedAt sql.NullTime

	err = row.Scan(&id, &name, &email, &role, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("No user found with ID %d", userid)
			return nil, fmt.Errorf("user not found")
		}
		log.Printf("Failed to scan user with ID %d: %v", userid, err)
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	log.Printf("id: %d, name: %s, email: %s, role: %s, created_at: %v, updated_at: %v\n", id, name, email, role, createdAt, updatedAt)

	createAt := timestamppb.New(createdAt)

	var updatedTime *timestamppb.Timestamp
	if updatedAt.Valid {
		updatedTime = timestamppb.New(updatedAt.Time)
	} else {
		updatedTime = nil
	}
	var role_to_db desc.Role
	switch role {
	case "Admin":
		role_to_db = desc.Role_Admin
	case "User":
		role_to_db = desc.Role_User
	default:
		log.Printf("Unknown role: %s", role)
		return nil, fmt.Errorf("unknown role: %s", role)
	}
	return &desc.GetResponse{Id: id, Name: name, Email: email, Role: desc.Role(role_to_db), CreateAt: createAt, UpdatedAt: updatedTime},
		nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
