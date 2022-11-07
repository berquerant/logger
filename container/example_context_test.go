package container_test

import (
	"context"
	"log"
	"os"

	"github.com/berquerant/logger"
	"github.com/berquerant/logger/container"
)

func ExampleContext() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	ctx := container.New(map[string]any{
		"RequestID": "stone1",
	}, logger.MustNewMapperFunc(logger.LogLevelToPrefixMapper).
		Next(logger.StandardLogConsumer)).WithContext(context.TODO())

	c1 := container.FromContext(ctx)
	c1.L().Info("first")
	c1.Data().Set("Path", "/update")
	c1.L().Info("second")

	func(ctx context.Context) {
		c2 := container.FromContext(ctx)
		c2.Data().Set("Verb", "POST")
		c2.L().Info("third")
	}(c1.Clone().WithContext(ctx))

	container.FromContext(ctx).L().Info("forth")
	// Output:
	// I | first | {"RequestID":"stone1"}
	// I | second | {"Path":"/update","RequestID":"stone1"}
	// I | third | {"Path":"/update","RequestID":"stone1","Verb":"POST"}
	// I | forth | {"Path":"/update","RequestID":"stone1"}
}
