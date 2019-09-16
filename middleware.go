package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/revel/revel"
)

type route struct {
	method string
	path   string
}

var nullLogger *log.Logger
var loadTestHandler = false

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}
func httpHandlerFunc(w http.ResponseWriter, r *http.Request) {}

func httpHandlerFuncTest(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, r.RequestURI)
}
func (m *mockResponseWriter) WriteHeader(int) {}
func init() {

	runtime.GOMAXPROCS(1)

	// makes logging 'webscale' (ignores them)
	log.SetOutput(new(mockResponseWriter))
	nullLogger = log.New(new(mockResponseWriter), "", 0)
	initGin()
	initRevel()

}
func ginHandle(_ *gin.Context) {}

func ginHandleWrite(c *gin.Context) {
	io.WriteString(c.Writer, c.Params.ByName("name"))
}

func ginHandleTest(c *gin.Context) {

	io.WriteString(c.Writer, c.Request.RequestURI)
}

func initGin() {
	gin.SetMode(gin.ReleaseMode)
}
func loadGinSingle(method, path string, handle gin.HandlerFunc) http.Handler {
	router := gin.New()
	router.Use(gin.Logger()) // use logger
	router.Handle(method, path, handle)
	return router
}

// echo
func echoHandler(c echo.Context) error {
	return c.String(200, "Hello")
}
func echoHandlerWrite(c echo.Context) error {

	return c.String(200, "Hello")
}
func echoHandlerTest(c echo.Context) error {
	io.WriteString(c.Response(), c.Request().RequestURI)
	return nil
}
func loadEchoSingle(method, path string, h echo.HandlerFunc) http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	switch method {
	case "GET":
		e.GET(path, h)
	case "POST":
		e.POST(path, h)
	case "PUT":
		e.PUT(path, h)
	case "PATCH":
		e.PATCH(path, h)
	case "DELETE":
		e.DELETE(path, h)
	default:
		panic("Unknow HTTP method: " + method)
	}
	return e
}
func goJSONRESTHandler(w rest.ResponseWriter, req *rest.Request) {
	w.WriteJson(map[string]string{"Body": "Hello World!"})
}

func gOJSONRESTHandlerWrite(w rest.ResponseWriter, req *rest.Request) {
	w.WriteJson(map[string]string{"Body": "Hello World!"})
}

func goJSONRESTHandlerTest(w rest.ResponseWriter, req *rest.Request) {
	io.WriteString(w.(io.Writer), req.RequestURI)
}
func loadGOJSONRESTSingle(method, path string, hfunc rest.HandlerFunc) http.Handler {
	api := rest.NewApi()
	// api.Use(rest.DefaultDevStack...)
	api.Use(&rest.AccessLogApacheMiddleware{})
	router, err := rest.MakeRouter(
		&rest.Route{method, path, hfunc},
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	return api.MakeHandler()
}

// 	buffalo
//  ENV ...
// var ENV = envy.Get("GO_ENV", "development")
var r *render.Engine

func buffaloHandler(c buffalo.Context) error {
	return c.Render(200, r.String("Hello"))
}
func buffaloHandlerWrite(c buffalo.Context) error {
	return c.Render(200, r.String("Hello"))
}
func buffaloHandlerTest(c buffalo.Context) error {
	io.WriteString(c.Response(), c.Request().RequestURI)
	return nil
}
func loadBuffaloSingle(method, path string, h buffalo.Handler) http.Handler {
	app := buffalo.New(buffalo.Options{})
	app.Use(buffalo.RequestLogger)
	// app.Use(contenttype.Set("pain/text"))
	// buffalo.NewOptions()
	switch method {
	case "GET":
		app.GET(path, h)
	case "POST":
		app.POST(path, h)
	case "PUT":
		app.PUT(path, h)
	case "PATCH":
		app.PATCH(path, h)
	case "DELETE":
		app.DELETE(path, h)
	default:
		panic("Unknow HTTP method: " + method)
	}
	// app.Serve()
	// app.PreHandlers = http.Handler.ServeHTTP(nil, nil)
	return app
}

// RevelController ...
// Revel (Router only)
// In the following code some Revel internals are modelled.
// The original revel code is copyrighted by Rob Figueiredo.
// See https://github.com/revel/revel/blob/master/LICENSE
type RevelController struct {
	*revel.Controller
	router *revel.Router
}

//Handle ...
func (rc *RevelController) Handle() revel.Result {
	return rc.RenderText("Hello")
}

//HandleWrite ...
func (rc *RevelController) HandleWrite() revel.Result {
	return rc.RenderText("Hello")
}

//HandleTest ...
func (rc *RevelController) HandleTest() revel.Result {
	return rc.RenderText(rc.Request.GetRequestURI())
}

type revelResult struct {
	Name string
}

func (rr revelResult) Apply(req *revel.Request, resp *revel.Response) {}

func (rc *RevelController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Dirty hacks, do NOT copy!
	revel.MainRouter = rc.router
	upgrade := r.Header.Get("Upgrade")
	if upgrade == "websocket" || upgrade == "Websocket" {
		panic("Not implemented")
	} else {
		context := revel.NewGoContext(nil)
		context.Response.SetResponse(w)
		context.Request.SetRequest(r)

		c := revel.NewController(context)

		c.Request.WebSocket = nil
		// revel.Filters[0](c, revel.Filters[1:])
		if c.Result != nil {
			c.Result.Apply(c.Request, c.Response)
		} else if c.Response.Status != 0 {
			panic("Not implemented")
		}
		// Close the Writer if we can
		if w, ok := c.Response.GetWriter().(io.Closer); ok {
			w.Close()
		}
	}
}
func initRevel() {
	// Only use the Revel filters required for this benchmark
	// revel.AppLog =
	revel.Filters = []revel.Filter{
		revel.RouterFilter,
		revel.ParamsFilter,
		revel.ActionInvoker,
	}

	revel.RegisterController((*RevelController)(nil),
		[]*revel.MethodType{
			{
				Name: "Handle",
			},
			{
				Name: "HandleWrite",
			},
			{
				Name: "HandleTest",
			},
		})
}

var (
	appModule = &revel.Module{Name: "App"}
)

func loadRevelSingle(method, path, action string) http.Handler {
	router := revel.NewRouter("")

	route := revel.NewRoute(appModule, method, path, action, "", "", 0)
	if err := router.Tree.Add(route.TreePath, route); err != nil {
		panic(err)
	}

	rc := new(RevelController)
	rc.router = router
	return rc
}

func main() {
	fmt.Println("Usage: go test -bench=. -timeout=20m")
	// e := echo.New()
	// e.Use(middleware.Logger())
	os.Exit(1)
}
