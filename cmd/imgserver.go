package main

import (

    "imgserver/object"
    "imgserver/config"
    "github.com/labstack/echo"
    "imgserver"

    "imgserver/response"
)

func main() {
    imgConfig := config.GetInstance()
    imgConfig.Init("config/config.yml")
    // Echo instance
    e := echo.New()

    // Route => handler
    e.GET("/*", func(ctx echo.Context) error {
        obj, err := object.NewFileObject(ctx.Request().URL.Path)
        if err != nil {
            return  ctx.NoContent(400)
        }

        res := imgserver.Process(obj)
        e.Logger.Info("res headers %s", res.Headers)
        //return ctx.JSON(200, obj)
        return ctx.Blob(res.StatusCode, res.Headers[response.ContentType], res.Body)
    })

    // Start server
    e.Logger.Fatal(e.Start(":8080"))
}
