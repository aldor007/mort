package main

import (

    "imgserver/object"
    "imgserver/config"
    "github.com/labstack/echo"
    "imgserver"
)

func main() {
    imgConfig := config.GetInstance()
    imgConfig.Init("config/config.yml")
    // Echo instance
    e := echo.New()

    // Route => handler
    e.GET("/*", func(ctx echo.Context) error {
        obj := object.NewFileObject(ctx.Request().URL.Path)
        response := imgserver.Process(obj)
        if response.Error != nil {
            return  ctx.NoContent(500)
        }

        //return ctx.JSON(200, obj)
        return ctx.Blob(response.StatusCode, response.Headers["content-type"], response.Body)
    })

    // Start server
    e.Logger.Fatal(e.Start(":8080"))
}
