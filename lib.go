package octanox

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	_ "github.com/joho/godotenv/autoload"
)

// Current is the current instance of the Octanox framework. Can be nil if no instance has been created.
var Current *Instance

// Instance is a struct that represents an instance of the Octanox framework.
type Instance struct {
	*Router
	// Gin is the underlying Gin engine that powers the Octanox framework's web server.
	Gin *gin.Engine
	// DB is the underlying GORM database connection that powers the Octanox framework's database operations.
	DB *gorm.DB
	// hooks is a map of hooks to their respective functions.
	hooks map[Hook][]func()
	// errorHandlers is a list of error handlers that can be called when an error occurs.
	errorHandlers []func(error)
	// isDebug is a flag that indicates whether the Octanox framework is running in debug mode.
	isDebug bool
	// isDryRun is a flag that indicates whether the Octanox framework is running in dry-run mode.
	isDryRun bool
}

// New creates a new instance of the Octanox framework. If an instance already exists, it will return the existing instance.
// This won't start the Octanox runtime, you need to call Run() on the instance to start the runtime.
func New() *Instance {
	if Current != nil {
		return Current
	}

	ginEngine := gin.New()

	Current := &Instance{
		Router: &Router{
			gin:    &ginEngine.RouterGroup,
			routes: make([]route, 0),
		},
		Gin:           ginEngine,
		hooks:         make(map[Hook][]func()),
		errorHandlers: make([]func(error), 0),
		isDebug:       gin.Mode() == gin.DebugMode,
		isDryRun:      os.Getenv("NOX__DRY_RUN") == "true",
	}

	return Current
}

// Hook registers a hook function to be called at a specific point in the Octanox runtime.
func (i *Instance) Hook(hook Hook, f func()) {
	if _, ok := i.hooks[hook]; !ok {
		i.hooks[hook] = make([]func(), 0)
	}

	i.hooks[hook] = append(i.hooks[hook], f)
}

// ErrorHandler registers an error handler function to be called when an error occurs in the Octanox runtime.
func (i *Instance) ErrorHandler(f func(error)) {
	i.errorHandlers = append(i.errorHandlers, f)
}

// Run starts the Octanox runtime. This function will block the current goroutine. If any error occurs, it will panic.
func (i *Instance) Run() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Println("Starting Octanox...")
	go i.runInternally()

	<-ctx.Done()

	log.Println("Shutting down...")
	i.emitHook(Hook_Shutdown)
}

func (i *Instance) emitHook(hook Hook) {
	if hooks, ok := i.hooks[hook]; ok {
		for _, f := range hooks {
			f()
		}
	}
}

func (i *Instance) emitError(err error) {
	for _, f := range i.errorHandlers {
		f(err)
	}
}

func (i *Instance) runInternally() {
	i.emitHook(Hook_Start)

	i.Gin.Use(cors())
	i.Gin.Use(logger())
	i.Gin.Use(recovery())
	i.Gin.Use(errorCollectorToHandler())

	i.emitHook(Hook_InitRoutes)

	if i.isDryRun {
		//TODO: run generator and stop
		return
	}

	i.Gin.Run()
}
