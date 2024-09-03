package octanox

import (
	"io"
	"net/http"
	"reflect"

	"github.com/goccy/go-json"

	"github.com/gin-gonic/gin"
)

type Request struct{}

type failedRequest struct {
	status  int
	message string
}

// Failed is a function that can be called to indicate that the request has failed and should abort with a specific status code and message.
// This function will panic with a failedRequest struct that will be caught by the Octanox framework.
func (r Request) Failed(status int, message string) {
	panic(failedRequest{status, message})
}

// GetRequest is a struct that represents a GET request.
type GetRequest struct {
	Request
}

// PostRequest is a struct that represents a POST request.
type PostRequest struct {
	Request
}

// PutRequest is a struct that represents a PUT request.
type PutRequest struct {
	Request
}

// DeleteRequest is a struct that represents a DELETE request.
type DeleteRequest struct {
	Request
}

// PatchRequest is a struct that represents a PATCH request.
type PatchRequest struct {
	Request
}

// OptionsRequest is a struct that represents an OPTIONS request.
type OptionsRequest struct {
	Request
}

// HeadRequest is a struct that represents a HEAD request.
type HeadRequest struct {
	Request
}

// TraceRequest is a struct that represents a TRACE request.
type TraceRequest struct {
	Request
}

// populateRequest is a function that extracts the request data from the Gin context, creates a new empty request struct from the given type, and populates it with the extracted data.
func populateRequest(c *gin.Context, reqType reflect.Type, user User) any {
	reqValue := reflect.New(reqType).Elem()

	for i := 0; i < reqType.NumField(); i++ {
		field := reqType.Field(i)
		fieldValue := reqValue.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		if field.Anonymous {
			embeddedReq := populateRequest(c, field.Type, user)
			fieldValue.Set(reflect.ValueOf(embeddedReq).Elem())
			continue
		}

		if userTag := field.Tag.Get("user"); userTag != "" {
			if user == nil {
				panic(failedRequest{
					status:  http.StatusUnauthorized,
					message: "Unauthorized: User is required but not provided",
				})
			}

			if fieldValue.Kind() == reflect.Ptr {
				fieldValue.Set(reflect.ValueOf(user))
			} else {
				fieldValue.Set(reflect.ValueOf(user).Elem())
			}

			continue
		}

		if ginTag := field.Tag.Get("gin"); ginTag != "" {
			if fieldValue.Kind() == reflect.Ptr {
				fieldValue.Set(reflect.ValueOf(c))
			} else {
				panic("field with 'gin' tag must be a pointer to a gin.Context")
			}

			continue
		}

		if pathParam := field.Tag.Get("path"); pathParam != "" {
			fieldValue.SetString(c.Param(pathParam))
		} else if queryParam := field.Tag.Get("query"); queryParam != "" {
			queryValue := c.Query(queryParam)
			if queryValue == "" && field.Tag.Get("optional") != "true" {
				panic(failedRequest{
					status:  http.StatusBadRequest,
					message: "Missing required query parameter: " + queryParam,
				})
			}
			fieldValue.SetString(queryValue)
		} else if headerParam := field.Tag.Get("header"); headerParam != "" {
			headerValue := c.GetHeader(headerParam)
			if headerValue == "" && field.Tag.Get("optional") != "true" {
				panic(failedRequest{
					status:  http.StatusBadRequest,
					message: "Missing required header: " + headerParam,
				})
			}
			fieldValue.SetString(headerValue)
		} else if bodyParam := field.Tag.Get("body"); bodyParam != "" {
			if field.Type.Kind() == reflect.Ptr {
				bodyInstance := reflect.New(field.Type.Elem()).Interface()

				if err := bindJsonFast(c, bodyInstance); err != nil {
					message := "Invalid JSON body"

					if Current.isDebug {
						message += ": " + err.Error()
					}

					panic(failedRequest{
						status:  http.StatusBadRequest,
						message: message,
					})
				}

				fieldValue.Set(reflect.ValueOf(bodyInstance))
			} else {
				bodyInstance := reflect.New(field.Type).Interface()

				if err := bindJsonFast(c, bodyInstance); err != nil {
					message := "Invalid JSON body"

					if Current.isDebug {
						message += ": " + err.Error()
					}

					panic(failedRequest{
						status:  http.StatusBadRequest,
						message: message,
					})
				}

				fieldValue.Set(reflect.ValueOf(bodyInstance).Elem())
			}
		}
	}

	// Return a pointer to the populated struct as an `any`
	return reqValue.Addr().Interface()
}

func bindJsonFast(c *gin.Context, v any) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}
