package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmm-bank/card-manager-service/models"
	"github.com/mmm-bank/infra/security"
	"log"
	"os"
)

const UniqueViolationCode = "23505"

var _ Storage = PostgresCards{}

type Storage interface {
	GetAllCards(userID uuid.UUID) ([]models.Card, error)
	AddCard(card models.Card) error
}

var key = os.Getenv("AES_KEY")

type PostgresCards struct {
	db *pgxpool.Pool
}

func NewPostgresCards(connString string) PostgresCards {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	return PostgresCards{pool}
}

func (p PostgresCards) AddCard(card models.Card) error {
	query := `
		INSERT INTO cards (card_id, account_id, user_id, card_number, phone_number)
		VALUES ($1, $2, $3, $4, $5)
	`
	cardNumber := security.Encrypt(card.CardNumber, key)
	phoneNumber := security.Encrypt(card.PhoneNumber, key)
	_, err := p.db.Exec(context.Background(), query,
		card.ID,
		card.AccountID,
		card.UserID,
		cardNumber,
		phoneNumber,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == UniqueViolationCode {
			return errors.New(UniqueViolationCode)
		}
		return fmt.Errorf("failed to insert card: %v", err)
	}
	return nil
}

func (p PostgresCards) GetAllCards(userID uuid.UUID) ([]models.Card, error) {
	query := `
		SELECT card_id, account_id, user_id, card_number, phone_number
		FROM cards WHERE user_id = $1 ORDER BY created_at DESC
	`
	rows, err := p.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch cards: %w", err)
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		var cardNumber []byte
		var phoneNumber []byte
		err = rows.Scan(
			&card.ID,
			&card.AccountID,
			&card.UserID,
			&cardNumber,
			&phoneNumber,
		)
		if err != nil {
			return nil, err
		}
		card.CardNumber = security.Decrypt(cardNumber, key)
		card.PhoneNumber = security.Decrypt(phoneNumber, key)
		cards = append(cards, card)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}
