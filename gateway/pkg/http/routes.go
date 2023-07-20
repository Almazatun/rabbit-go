package routes

import (
	"log"

	inputs "github.com/Almazatun/rabbit-go/gateway/pkg/common"
	"github.com/Almazatun/rabbit-go/gateway/pkg/rmq"
	"github.com/gofiber/fiber/v2"
)

func PublicRoutes(a *fiber.App, rm *rmq.RMQ) {
	route := a.Group("/api/")

	route.Post("rpc", func(c *fiber.Ctx) error {
		input := inputs.RpcInput{}

		//  Parse body into product struct
		if err := c.BodyParser(&input); err != nil {
			log.Println(err)
			c.Status(400).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
			return nil
		}

		res, err := rm.SendMsgReply(input.RK, input.Msg, input.Method)

		if err != nil {
			c.Status(400).JSON(&fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
			return nil
		}

		c.Status(200).JSON(&fiber.Map{
			"success": true,
			"message": "✅",
			"res":     res,
		})

		return nil
	})
}
