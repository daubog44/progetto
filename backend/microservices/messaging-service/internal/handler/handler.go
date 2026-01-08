package handler

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/username/progetto/messaging-service/internal/repository"
)

type Handler struct {
	Repo      repository.UserRepository
	Publisher message.Publisher
	Logger    *slog.Logger
}
