// Copyright 2022 rkonfj@fnla.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"

	"github.com/kataras/iris/v12"
)

func Routing(app *iris.Application, agent *Agent) {
	// create volume
	app.Post("/kind/{kind:string}/disk/{name:string}/{size:int64}", func(ctx iris.Context) {
		kind := ctx.Params().GetStringDefault("kind", "default")
		name := ctx.Params().GetString("name")
		size := ctx.Params().GetInt64Default("size", 1024*1024*10)
		err := agent.CreateVolume(kind, name, size)
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{
				"Message": err.Error(),
			})
			log.Println(err)
			return
		}
		ctx.JSON(iris.Map{
			"Message": "ok",
		})
	})

	app.Put("/disk/{name:string}/{size:int64}", func(ctx iris.Context) {
		name := ctx.Params().GetString("name")
		size, err := ctx.Params().GetInt64("size")
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{
				"Message": err.Error(),
			})
			log.Println(err)
			return
		}
		err = agent.ExpandVolume(name, size)
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{
				"Message": err.Error(),
			})
			log.Println(err)
			return
		}
		ctx.JSON(iris.Map{
			"Message": "ok",
		})
	})

	// delete volume
	app.Delete("/disk/{name:string}", func(ctx iris.Context) {
		name := ctx.Params().GetString("name")
		err := agent.DeleteVolume(name)
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{
				"Message": err.Error(),
			})
			log.Println(err)
			return
		}
		ctx.JSON(iris.Map{
			"Message": "ok",
		})
	})

	// publish volume
	app.Post("/disk/{name:string}/publish/{node:string}", func(ctx iris.Context) {
		name := ctx.Params().GetString("name")
		node := ctx.Params().GetString("node")
		err := agent.PublishVolume(name, node)
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{
				"Message": err.Error(),
			})
			log.Println(err)
			return
		}
		ctx.JSON(iris.Map{
			"Message": "ok",
		})
	})

	// unpublish volume
	app.Delete("/disk/{name:string}/publish/{node:string}", func(ctx iris.Context) {
		name := ctx.Params().GetString("name")
		node := ctx.Params().GetString("node")
		err := agent.UnpublishVolume(name, node)
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{
				"Message": err.Error(),
			})
			log.Println(err)
			return
		}
		ctx.JSON(iris.Map{
			"Message": "ok",
		})
	})

	// get devpath
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
