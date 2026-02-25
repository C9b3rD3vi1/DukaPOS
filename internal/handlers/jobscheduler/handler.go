package handlers

import (
	"github.com/C9b3rD3vi1/DukaPOS/internal/services/job"
	"github.com/gofiber/fiber/v2"
)

type JobSchedulerHandler struct {
	scheduler *job.Scheduler
}

func NewJobSchedulerHandler(scheduler *job.Scheduler) *JobSchedulerHandler {
	return &JobSchedulerHandler{scheduler: scheduler}
}

func (h *JobSchedulerHandler) RegisterRoutes(app fiber.Router) {
	jobs := app.Group("/jobs")
	jobs.Get("/status", h.GetStatus)
	jobs.Post("/run/:name", h.RunJob)
	jobs.Get("/list", h.ListJobs)
}

func (h *JobSchedulerHandler) GetStatus(c *fiber.Ctx) error {
	if h.scheduler == nil {
		return c.JSON(fiber.Map{
			"error": "Job scheduler not available",
		})
	}

	status := h.scheduler.GetStatus()
	return c.JSON(status)
}

func (h *JobSchedulerHandler) RunJob(c *fiber.Ctx) error {
	if h.scheduler == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Job scheduler not available",
		})
	}

	jobName := c.Params("name")
	if jobName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Job name is required",
		})
	}

	err := h.scheduler.RunJob(jobName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Job started",
		"job":     jobName,
	})
}

func (h *JobSchedulerHandler) ListJobs(c *fiber.Ctx) error {
	if h.scheduler == nil {
		return c.JSON(fiber.Map{
			"error": "Job scheduler not available",
		})
	}

	jobs := h.scheduler.ListJobs()
	return c.JSON(fiber.Map{
		"jobs": jobs,
	})
}
