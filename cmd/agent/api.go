package main

import (
	"log"

	"github.com/kataras/iris/v12"
)

func Routing(app *iris.Application, agent *Agent) {
	app.Post("/disk/{name:string}/{size:int64}", func(ctx iris.Context) {
		name := ctx.Params().GetString("name")
		size := ctx.Params().GetInt64Default("size", 1024*1024*10)
		err := agent.CreateVolume(name, size)
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{
				"message": err.Error(),
			})
			log.Println(err)
		}
	})

	app.Put("/disk/{name:string}/{size:int64}", func(ctx iris.Context) {
		name := ctx.Params().GetString("name")
		size, err := ctx.Params().GetInt64("size")
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{
				"message": err.Error(),
			})
			log.Println(err)
			return
		}
		err = agent.ExpandVolume(name, size)
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{
				"message": err.Error(),
			})
			log.Println(err)
		}
	})

	app.Delete("/disk/{name:string}", func(ctx iris.Context) {
		name := ctx.Params().GetString("name")
		err := agent.DeleteVolume(name)
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{
				"message": err.Error(),
			})
			log.Println(err)
		}
	})

	app.Post("/disk/{name:string}/publish/{node:string}", func(ctx iris.Context) {
		name := ctx.Params().GetString("name")
		node := ctx.Params().GetString("node")
		err := agent.PublishVolume(name, node)
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{
				"message": err.Error(),
			})
			log.Println(err)
		}
	})

	app.Delete("/disk/{name:string}/publish/{node:string}", func(ctx iris.Context) {
		name := ctx.Params().GetString("name")
		node := ctx.Params().GetString("node")
		err := agent.UnpublishVolume(name, node)
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{
				"message": err.Error(),
			})
			log.Println(err)
		}
	})

	app.Get("/dev-path/{name:string}", func(ctx iris.Context) {
		name := ctx.Params().GetString("name")
		path, err := agent.GetDevPath(name)
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{
				"message": err.Error(),
			})
			log.Println(err)
			return
		}
		ctx.Write([]byte(path))
	})

	app.Get("/capacity", func(ctx iris.Context) {
		ctx.JSON(iris.Map{
			"Available":         1024 * 1024 * 1024 * 1024 * 2,
			"MaximumVolumeSize": 1024 * 1024 * 1024 * 100,
			"MinimumVolumeSize": 1024 * 1024 * 10,
		})
	})

	app.Get("/probe", func(ctx iris.Context) {
		ctx.JSON(iris.Map{
			"Message": "ok",
		})
	})
}
