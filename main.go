package main

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware"

	"go-fiber-todos/postgres"

	_ "github.com/lib/pq"
)

func mapTodo(todo postgres.Todo) interface{} {
	return struct {
		ID int64 `json:"id"`
		Name string `json:"name"`
		Completed bool `json:"completed"`
	}{
		ID: todo.ID,
		Name: todo.Name,
		Completed: todo.Completed.Bool,
	}
}

type Handlers struct {
	Repo *postgres.Repo
}

func NewHandlers(repo *postgres.Repo) *Handlers {
	return &Handlers{Repo: repo}
}

func main() {
	db, err := sql.Open("postgres", fmt.Sprintf("dbname=%s password=postgres user=postgres sslmode=disable", "fiber_todos_dev"))
	if err != nil {
		panic(err)
	}

	repo := postgres.NewRepo(db)

	app := fiber.New()

	app.Use(middleware.Logger())
	app.Use(middleware.Recover())

	app.Get("/", func(ctx *fiber.Ctx) {
		ctx.Send("hello world")
	})

	handlers := NewHandlers(repo)

	SetupApiV1(app, handlers)

	err = app.Listen(3000)
	if err != nil {
		panic(err)
	}
}

func SetupApiV1(app *fiber.App, handlers *Handlers) {
	v1 := app.Group("/v1")

	SetupTodosRoutes(v1, handlers)
}

func SetupTodosRoutes(grp fiber.Router, handlers *Handlers) {
	todosRoutes := grp.Group("/todos")
	todosRoutes.Get("/", handlers.GetTodos)
	todosRoutes.Post("/", handlers.CreateTodo)
	todosRoutes.Get("/:id", handlers.GetTodo)
	todosRoutes.Delete("/:id", handlers.DeleteTodo)
	todosRoutes.Patch("/:id", handlers.UpdateTodo)
}

func (h *Handlers) UpdateTodo(ctx *fiber.Ctx) {
	type request struct {
		Name      *string `json:"name"`
		Completed *bool   `json:"completed"`
	}

	paramsId := ctx.Params("id")
	id, err := strconv.Atoi(paramsId)
	if err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse id",
		})
		return
	}

	var body request
	err = ctx.BodyParser(&body)
	if err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse body",
		})
		return
	}

	todo, err := h.Repo.GetTodoById(ctx.Context(), int64(id))
	if err != nil {
		ctx.Status(fiber.StatusNotFound)
		return
	}

	if body.Name != nil {
		todo.Name = *body.Name
	}

	if body.Completed != nil {
		todo.Completed = sql.NullBool{
			Bool: *body.Completed,
			Valid: true,
		}
	}

	todo, err = h.Repo.UpdateTodo(ctx.Context(), postgres.UpdateTodoParams{
		ID:        int64(id),
		Name:      todo.Name,
		Completed: todo.Completed,
	})
	if err != nil {
		ctx.SendStatus(fiber.StatusNotFound)
		return
	}

	if err := ctx.Status(fiber.StatusOK).JSON(mapTodo(todo)); err != nil {
		ctx.Status(fiber.StatusInternalServerError).Send(err.Error())
		return
	}
}

func (h *Handlers) DeleteTodo(ctx *fiber.Ctx) {
	paramsId := ctx.Params("id")
	id, err := strconv.Atoi(paramsId)
	if err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse id",
		})
		return
	}

	_, err = h.Repo.GetTodoById(ctx.Context(), int64(id))
	if err != nil {
		ctx.Status(fiber.StatusNotFound)
		return
	}

	err = h.Repo.DeleteTodoById(ctx.Context(), int64(id))
	if err != nil {
		ctx.SendStatus(fiber.StatusNotFound)
		return
	}

	ctx.SendStatus(fiber.StatusNoContent)
}

func (h *Handlers) GetTodo(ctx *fiber.Ctx) {
	paramsId := ctx.Params("id")
	id, err := strconv.Atoi(paramsId)
	if err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse id",
		})
		return
	}

	todo, err := h.Repo.GetTodoById(ctx.Context(), int64(id))
	if err != nil {
		ctx.Status(fiber.StatusNotFound)
		return
	}

	if err := ctx.Status(fiber.StatusOK).JSON(mapTodo(todo)); err != nil {
		ctx.Status(fiber.StatusInternalServerError).Send(err.Error())
		return
	}
}

func (h *Handlers) CreateTodo(ctx *fiber.Ctx) {
	type request struct {
		Name string `json:"name"`
	}

	var body request

	err := ctx.BodyParser(&body)
	if err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse json",
		})
		return
	}

	if len(body.Name) <= 2 {
		ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name not long enough",
		})
		return
	}

	todo, err := h.Repo.CreateTodo(ctx.Context(), body.Name)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).Send(err.Error())
		return
	}

	if err := ctx.Status(fiber.StatusCreated).JSON(mapTodo(todo)); err != nil {
		ctx.Status(fiber.StatusInternalServerError).Send(err.Error())
		return
	}
}

func (h *Handlers) GetTodos(ctx *fiber.Ctx) {
	todos, err := h.Repo.GetAllTodos(ctx.Context())
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).Send(err.Error())
		return
	}

	result := make([]interface{}, len(todos))
	for i, todo := range todos {
		result[i] = mapTodo(todo)
	}

	if err := ctx.Status(fiber.StatusOK).JSON(result); err != nil {
		ctx.Status(fiber.StatusInternalServerError).Send(err.Error())
		return
	}
}
