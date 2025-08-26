// handlers/validate.go
package handlers

import (
	domain "github.com/gulll/deepmarket/backtesting/domain"
	engine "github.com/gulll/deepmarket/backtesting/engine"

	"github.com/gofiber/fiber/v2"
	"github.com/gulll/deepmarket/models"
)

type ValidateReq struct {
	Condition domain.Condition `json:"condition"`
}
type ValidateResp struct {
	Valid  bool   `json:"valid"`
	Reason string `json:"reason,omitempty"`
}

func ValidateConditionHandler(reg *engine.Registry) fiber.Handler {
	parser := &engine.Parser{Reg: reg}
	return func(c *fiber.Ctx) error {
		var req ValidateReq
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(models.APIResponse{
				Success: false,
				Message: "Invalid Request format",
			})
		}
		if err := parser.ValidateCondition(req.Condition); err != nil {
			return c.Status(200).JSON(models.APIResponse{
				Success: false,
				Message: err.Error(),
			})
		}
		r := models.APIResponse{
			Success: true,
			Message: "Condition is valid",
		}
		return c.JSON(r)
	}
}
