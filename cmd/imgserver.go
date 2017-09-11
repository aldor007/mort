package cmd

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
        obj, err := object.NewFileObject(ctx.Request().URL.Path)
        if err != nil {
            return  ctx.NoContent(400)
        }

        response := imgserver.Process(obj)
        if response.Error != nil {
            e.Logger.Error(response.Error)
            return  ctx.NoContent(response.StatusCode)
        }

        //return ctx.JSON(200, obj)
        return ctx.Blob(response.StatusCode, response.Headers["content-type"], response.Body)
    })

    // Start server
    e.Logger.Fatal(e.Start(":8080"))
}
