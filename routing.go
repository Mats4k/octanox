package octanox

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

// Router is a struct that represents a router in the Octanox framework. It wraps around a Gin router group with the only two differences
// to populate the request handlers, handling responses and emit the DTOs to the client code generation process.
type SubRouter struct {
	gin    *gin.RouterGroup
	routes []route
}

// route is a struct containing metadata about a route in the Octanox framework.
type route struct {
	method       string
	path         string
	requestType  reflect.Type
	responseType reflect.Type
}

// Router creates a new router with the given URL prefix.
func (r *SubRouter) Router(url string) *SubRouter {
	return &SubRouter{
		gin:    r.gin.Group(url),
		routes: make([]route, 0),
	}
}

// Register registers a new route handler. The function automatically detects the method, request and response type. If any of these detection fails, it will panic.
func (r *SubRouter) Register(path string, handler interface{}) {
	handlerType := reflect.TypeOf(handler)

	if handlerType.Kind() != reflect.Func || handlerType.NumIn() != 1 || handlerType.NumOut() != 1 {
		panic("Handler function must have one input parameter and one return value")
	}

	reqType := handlerType.In(0)
	resType := handlerType.Out(0)

	method := detectHTTPMethod(reqType)

	routeMeta := route{
		method:       method,
		path:         path,
		requestType:  reqType,
		responseType: resType,
	}

	if Current.isDryRun {
		r.routes = append(r.routes, routeMeta)
	}

	r.gin.Handle(method, routeMeta.path, func(c *gin.Context) {
		wrapHandler(c, reqType, reflect.ValueOf(handler))
	})
}

// detectHTTPMethod determines the HTTP method from the embedded struct in the request type.
func detectHTTPMethod(reqType reflect.Type) string {
	for i := 0; i < reqType.NumField(); i++ {
		field := reqType.Field(i)

		if field.Anonymous {
			switch field.Type {
			case reflect.TypeOf(GetRequest{}):
				return http.MethodGet
			case reflect.TypeOf(PostRequest{}):
				return http.MethodPost
			case reflect.TypeOf(PutRequest{}):
				return http.MethodPut
			case reflect.TypeOf(DeleteRequest{}):
				return http.MethodDelete
			case reflect.TypeOf(PatchRequest{}):
				return http.MethodPatch
			case reflect.TypeOf(OptionsRequest{}):
				return http.MethodOptions
			case reflect.TypeOf(HeadRequest{}):
				return http.MethodHead
			case reflect.TypeOf(TraceRequest{}):
				return http.MethodTrace
			}
		}
	}

	panic("Failed to detect HTTP method: No recognized embedded request struct found")
}

// wrapHandler wraps the gin context and the handler function to call the handler function with the correct parameters and handle the response.
func wrapHandler(c *gin.Context, reqType reflect.Type, handler reflect.Value) {
	//TODO: handle last param (nil) as user.
	req := populateRequest(c, reqType, nil)
	res := handler.Call([]reflect.Value{reflect.ValueOf(req)})[0].Interface()

	if res == nil {
		c.Status(204)
		return
	}

	if _, ok := res.(error); ok {
		panic(res)
	}

	c.JSON(200, res)
}
