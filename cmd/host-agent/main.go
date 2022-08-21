package main

import (
	"log"

	"github.com/kataras/iris/v12"
)

func main() {
	app := iris.New()

	app.Post("/disk/{name:string}/{size:int64}", func(ctx iris.Context) {
		name := ctx.Params().GetString("name")
		size := ctx.Params().GetInt64Default("size", 1024*1024*10)
		err := DoCreateVolume(name, size)
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
		err = DoExpandVolume(name, size)
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
		err := DoDeleteVolume(name)
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
		err := DoPublishVolume(name, node)
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
		err := DoUnpublishVolume(name, node)
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
		path, err := DoGetDevPath(name)
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

	app.Listen(":8080")
}
