package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/sudomateo/terraform-provider-todo/todo"
)

func main() {
	providerserver.Serve(context.Background(), todo.New, providerserver.ServeOpts{
		Address: "sudomateo.dev/sudomateo/todo",
	})
}
