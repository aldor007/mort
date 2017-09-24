package main

import (
    "net/http"
    "time"

    "mort/object"
    "mort/config"
    "github.com/labstack/echo"
    "mort"

    "mort/response"
)

func main() {
    imgConfig := config.GetInstance()
    imgConfig.Init("configuration/config.yml")
    // Echo instance
    e := echo.New()

    // Route => handler
    e.GET("/*", func(ctx echo.Context) error {
        obj, err := object.NewFileObject(ctx.Request().URL.Path)
        if err != nil {
            return  ctx.NoContent(400)
        }
// dodac placeholder
        res := mort.Process(obj)
        res.WriteHeaders(ctx.Response())

        e.Logger.Info("res headers %s", res.Headers)
        //return ctx.JSON(200, obj)
        return ctx.Blob(res.StatusCode, res.Headers[response.ContentType], res.Body)
    })

    s := &http.Server{
        Addr:         ":8080",
        ReadTimeout:  1 * time.Minute,
        WriteTimeout: 1 * time.Minute,
    }

    e.Logger.Fatal(e.StartServer(s))

}
