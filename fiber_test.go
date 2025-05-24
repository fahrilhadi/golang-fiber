package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

var app = fiber.New(fiber.Config{
	ErrorHandler: func (ctx *fiber.Ctx, err error) error {
		ctx.Status(fiber.StatusInternalServerError)
		return ctx.SendString("Error : " + err.Error())
	},
})

func TestRoutingHelloWorld(t *testing.T)  {
	app.Get("/", func (ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World")
	})

	request := httptest.NewRequest("GET", "/", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello World", string(bytes))
}

func TestCtx(t *testing.T)  {
	app.Get("/hello", func (ctx *fiber.Ctx) error {
		name := ctx.Query("name", "Guest")
		return ctx.SendString("Hello " + name)
	})

	request := httptest.NewRequest("GET", "/hello?name=Fahril", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Fahril", string(bytes))

	request1 := httptest.NewRequest("GET", "/hello", nil)
	response1, err := app.Test(request1)
	assert.Nil(t, err)
	assert.Equal(t, 200, response1.StatusCode)

	bytes1, err1 := io.ReadAll(response1.Body)
	assert.Nil(t, err1)
	assert.Equal(t, "Hello Guest", string(bytes1))
}

func TestHttpRequest(t *testing.T)  {
	app.Get("/request", func (ctx *fiber.Ctx) error {
		first := ctx.Get("firstname")
		last := ctx.Cookies("lastname")
		return ctx.SendString("Hello " + first + " " + last)
	})

	request := httptest.NewRequest("GET", "/request", nil)
	request.Header.Set("firstname", "Fahril")
	request.AddCookie(&http.Cookie{Name: "lastname", Value: "Hadi"})
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Fahril Hadi", string(bytes))
}

func TestRouteParameter(t *testing.T)  {
	app.Get("/users/:userId/orders/:orderId", func (ctx *fiber.Ctx) error {
		userId := ctx.Params("userId")
		orderId := ctx.Params("orderId")
		return ctx.SendString("Get Order " + orderId + " From User " + userId)
	})

	request := httptest.NewRequest("GET", "/users/fahril/orders/10", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Get Order 10 From User fahril", string(bytes))
}

func TestFormRequest(t *testing.T)  {
	app.Post("/hello", func (ctx *fiber.Ctx) error {
		name := ctx.FormValue("name")
		return ctx.SendString("Hello " + name)
	})

	body := strings.NewReader("name=Fahril")
	request := httptest.NewRequest("POST", "/hello", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Fahril", string(bytes))
}

//go:embed source/contoh.txt
var contohFile []byte

func TestFormUpload(t *testing.T)  {
	app.Post("/upload", func (ctx *fiber.Ctx) error {
		file, err := ctx.FormFile("file")
		if err != nil {
			return err
		}

		err = ctx.SaveFile(file, "./target/"+file.Filename)
		if err != nil {
			return err
		}

		return ctx.SendString("Upload Success")
	})

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	file, err := writer.CreateFormFile("file", "contoh.txt")
	assert.Nil(t, err)
	file.Write(contohFile)
	writer.Close()
	
	request := httptest.NewRequest("POST", "/upload", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Upload Success", string(bytes))
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func TestRequestBody(t *testing.T)  {
	app.Post("/login", func (ctx *fiber.Ctx) error {
		body := ctx.Body()

		request := new(LoginRequest)
		err := json.Unmarshal(body, request)
		if err != nil {
			return err
		}

		return ctx.SendString("Hello " + request.Username)
	})

	body := strings.NewReader(`{"username":"Fahril", "password":"rahasia"}`)
	request := httptest.NewRequest("POST", "/login", body)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Fahril", string(bytes))
}

type RegisterRequest struct {
	Username string `json:"username" xml:"username" form:"username"`
	Password string `json:"password" xml:"password" form:"password"`
	Name string `json:"name" xml:"name" form:"name"`
}

func TestBodyParser(t *testing.T)  {
	app.Post("/register", func (ctx *fiber.Ctx) error {
		request := new(RegisterRequest)
		err := ctx.BodyParser(request)
		if err != nil {
			return nil
		}

		return ctx.SendString("Register Success " + request.Username)
	})
}

func TestBodyParserJSON(t *testing.T)  {
	TestBodyParser(t)

	body := strings.NewReader(`{"username":"Fahril", "password":"rahasia", "name":"Fahril Hadi"}`)
	request := httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register Success Fahril", string(bytes))
}

func TestBodyParserForm(t *testing.T)  {
	TestBodyParser(t)

	body := strings.NewReader(`username=Fahril&password=rahasia&name=Fahril+Hadi`)
	request := httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register Success Fahril", string(bytes))
}

func TestBodyParserXML(t *testing.T)  {
	TestBodyParser(t)

	body := strings.NewReader(
		`<RegisterRequest>
			<username>Fahril</username>
			<password>rahasia</password>
			<name>Fahril Hadi</name>
		</RegisterRequest>
		`)
	request := httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type", "application/xml")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register Success Fahril", string(bytes))
}

func TestResponseJSON(t *testing.T)  {
	app.Get("/user", func (ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"username" : "hadi",
			"name" : "Fahril Hadi",
		})
	})

	request := httptest.NewRequest("GET", "/user", nil)
	request.Header.Set("Accept", "application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, `{"name":"Fahril Hadi","username":"hadi"}`, string(bytes))
}

func TestDownloadFile(t *testing.T)  {
	app.Get("/download", func (ctx *fiber.Ctx) error {
		return ctx.Download("./source/contoh.txt", "contoh.txt")
	})

	request := httptest.NewRequest("GET", "/download", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, `attachment; filename="contoh.txt"`, response.Header.Get("Content-Disposition"))

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "this is sample file for upload", string(bytes))
}

func TestRoutingGroup(t *testing.T)  {
	helloworld := func (ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World")
	}

	api := app.Group("/api")
	api.Get("/hello", helloworld) // /api/hello
	api.Get("/world", helloworld) // /api/world

	web := app.Group("/web")
	web.Get("/hello", helloworld) // /web/hello
	web.Get("/world", helloworld) // /web/world

	request := httptest.NewRequest("GET", "/api/hello", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello World", string(bytes))
}

func TestStatic(t *testing.T)  {
	app.Static("/public", "./source")

	request := httptest.NewRequest("GET", "/public/contoh.txt", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "this is sample file for upload", string(bytes))
}

func TestErrorHandler(t *testing.T)  {
	app.Get("/error", func (ctx *fiber.Ctx) error {
		return errors.New("ups")
	})

	request := httptest.NewRequest("GET", "/error", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 500, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Error : ups", string(bytes))
}