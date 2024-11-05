package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mmm-bank/card-manager-service/generate"
	"github.com/mmm-bank/card-manager-service/models"
	"github.com/mmm-bank/card-manager-service/storage"
	"github.com/mmm-bank/infra/middleware"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"time"
)

const UniqueViolationCode = "23505"

type CardService struct {
	db     storage.Storage
	logger *zap.Logger
}

type AccountRequest struct {
	models.Card
	AccountNumber string `json:"account_number" validate:"required"`
}

func NewCardService(db storage.Storage) *CardService {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	return &CardService{db: db, logger: logger}
}

func parseAccountRequest(r *http.Request) (req AccountRequest, Err string) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	if err := json.NewDecoder(r.Body).Decode(&req.Card); err != nil {
		Err = "Failed to parse JSON request"
		return
	}
	req.Card.ID = uuid.New()
	req.UserID = userID
	req.AccountID = uuid.New()
	req.CardNumber = generate.CardNumber()
	req.AccountNumber = generate.AccountNumber()
	if err := validator.New().Struct(req); err != nil {
		Err = "Missing fields"
	}
	return
}

func (c *CardService) createAccount(data *AccountRequest) error {
	transactionServiceURL := os.Getenv("TRANSACTION_SERVICE_URL") + "/service/account/create"

	body, err := json.Marshal(data)
	if err != nil {
		c.logger.Error("Failed to marshal account creation request", zap.Error(err))
		return err
	}

	req, err := http.NewRequest(http.MethodPost, transactionServiceURL, bytes.NewReader(body))
	if err != nil {
		c.logger.Error("Failed to prepare account creation HTTP request", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Error("Failed to create account in transaction service", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		c.logger.Error("Failed to create account in transaction service", zap.Int("status", resp.StatusCode))
		return fmt.Errorf("transaction service returned unexpected status: %d", resp.StatusCode)
	}
	return nil
}

func (c *CardService) postAddCardHandler(w http.ResponseWriter, r *http.Request) {
	req, Err := parseAccountRequest(r)
	if Err != "" {
		http.Error(w, Err, http.StatusBadRequest)
		return
	}

	for err := c.db.AddCard(req.Card); err != nil; err = c.db.AddCard(req.Card) {
		if err.Error() != UniqueViolationCode {
			c.logger.Error("Failed to add card", zap.Error(err), zap.String("user_id", req.UserID.String()))
			http.Error(w, "Failed to add card", http.StatusInternalServerError)
			return
		}
	}

	if err := c.createAccount(&req); err != nil {
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func parseCardRequest(r *http.Request) (card models.Card, Err string) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		Err = "Failed to parse JSON request"
		return
	}
	card.ID = uuid.New()
	card.UserID = userID
	card.AccountID = uuid.New()
	card.CardNumber = generate.CardNumber()
	if err := validator.New().Struct(card); err != nil {
		Err = "Missing fields"
	}
	return
}

func (c *CardService) linkCard(data *models.Card) error {
	transactionServiceURL := os.Getenv("TRANSACTION_SERVICE_URL") + "/service/account/link/card"

	body, err := json.Marshal(data)
	if err != nil {
		c.logger.Error("Failed to marshal card linking request", zap.Error(err))
		return err
	}

	req, err := http.NewRequest(http.MethodPost, transactionServiceURL, bytes.NewReader(body))
	if err != nil {
		c.logger.Error("Failed to prepare card linking HTTP request", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Error("Failed to link card in transaction service", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		c.logger.Error("Failed to link card in transaction service", zap.Int("status", resp.StatusCode))
		return fmt.Errorf("transaction service returned unexpected status: %d", resp.StatusCode)
	}
	return nil
}

func (c *CardService) postAddLinkedCardHandler(w http.ResponseWriter, r *http.Request) {
	card, Err := parseCardRequest(r)
	if Err != "" {
		http.Error(w, Err, http.StatusBadRequest)
		return
	}

	for err := c.db.AddCard(card); err != nil; err = c.db.AddCard(card) {
		if err.Error() != UniqueViolationCode {
			c.logger.Error("Failed to add card", zap.Error(err), zap.String("user_id", card.UserID.String()))
			http.Error(w, "Failed to add card", http.StatusInternalServerError)
			return
		}
	}

	if err := c.linkCard(&card); err != nil {
		http.Error(w, "Failed to link card", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *CardService) getCardsInfoHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	cards, err := c.db.GetAllCards(userID)
	if err != nil {
		c.logger.Error("Failed to get cards info", zap.Error(err), zap.String("user_id", userID.String()))
		http.Error(w, "Failed to get cards info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

func CreateAndRunServer(s *CardService, addr string) error {
	r := chi.NewRouter()
	r.Use(mymiddleware.ExtractPayload)
	r.Route("/cards", func(r chi.Router) {
		r.Post("/create", s.postAddCardHandler)
		r.Post("/create/link", s.postAddLinkedCardHandler)

		r.Get("/all", s.getCardsInfoHandler) // TODO read balance from transaction-service
	})
	return http.ListenAndServe(addr, r)
}
