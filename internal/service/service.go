package service

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"simple-service/internal/dto"
	"simple-service/internal/repo"
	"simple-service/pkg/validator"
	"strconv"
)

// Слой бизнес-логики. Тут должна быть основная логика сервиса

// Service - интерфейс для бизнес-логики
type Service interface {
	CreateTask(ctx *fiber.Ctx) error
	GetTaskByID(ctx *fiber.Ctx) error // Новый метод

}

type service struct {
	repo repo.Repository
	log  *zap.SugaredLogger
}

// NewService - конструктор сервиса
func NewService(repo repo.Repository, logger *zap.SugaredLogger) Service {
	return &service{
		repo: repo,
		log:  logger,
	}
}

// CreateTask - обработчик запроса на создание задачи
func (s *service) CreateTask(ctx *fiber.Ctx) error {
	var req TaskRequest

	// Десериализация JSON-запроса
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		s.log.Error("Invalid request body", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request body")
	}

	// Валидация входных данных
	if vErr := validator.Validate(ctx.Context(), req); vErr != nil {
		return dto.BadResponseError(ctx, dto.FieldIncorrect, vErr.Error())
	}

	// Вставка задачи в БД через репозиторий
	task := repo.Task{
		Title:       req.Title,
		Description: req.Description,
	}
	taskID, err := s.repo.CreateTask(ctx.Context(), task)
	if err != nil {
		s.log.Error("Failed to insert task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

	// Формирование ответа
	response := dto.Response{
		Status: "success",
		Data:   map[string]int{"task_id": taskID},
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

// GetTaskByID - обработчик запроса на получение задачи по ID
func (s *service) GetTaskByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id") // Получаем ID из параметров маршрута

	// Преобразуем ID в int
	taskID, err := strconv.Atoi(id)
	if err != nil {
		s.log.Error("Invalid task ID", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldIncorrect, "Invalid task ID")
	}

	// Получаем задачу из репозитория
	task, err := s.repo.GetTaskByID(ctx.Context(), taskID)
	if err != nil {
		s.log.Error("Failed to get task by ID", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

	if task == nil {
		return dto.BadResponseError(ctx, dto.FieldIncorrect, "Task not found")
	}

	// Формируем ответ
	response := dto.Response{
		Status: "success",
		Data:   task, // Отправляем найденную задачу в ответе
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}
