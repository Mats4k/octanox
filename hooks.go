package octanox

// Hook is the type of a hook function that can be registered within the Octanox framework.
type Hook string

const (
	// Init is a hook that is called when the Octanox runtime is initializing.
	Hook_Init Hook = "init"
	// BeforeStart is a hook that is called when the Octanox runtime is registering its routes just before starting the web server. Here all routes should be registered.
	Hook_BeforeStart Hook = "before_start"
	// Shutdown is a hook that is called when the Octanox runtime is shutting down.
	Hook_Shutdown Hook = "shutdown"
)
