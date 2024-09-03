package octanox

// Hook is the type of a hook function that can be registered within the Octanox framework.
type Hook string

const (
	// Start is a hook that is called before the Octanox runtime starts.
	Hook_Start Hook = "start"
	// InitRoutes is a hook that is called when the Octanox runtime is initializing its routes. Here all routes should be registered.
	Hook_InitRoutes Hook = "init_routes"
	// Shutdown is a hook that is called when the Octanox runtime is shutting down.
	Hook_Shutdown Hook = "shutdown"
)
