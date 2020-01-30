// This package implements helper functions to parse service requests and render service responses.
//
// Although service is the only entry point into the functions of this module and the implementation of HTTP handlers
// are left completely to the developer, this package still provides utilities that assumes usage of Go's native HTTP
// stack, which proves to be the common scenario. This will make HTTP handler implementation even easier.
package handlerutil
