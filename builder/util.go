package builder

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/engine-api/types"
	mruby "github.com/mitchellh/go-mruby"
)

func (b *Builder) commit(cacheKey string, hook func(b *Builder, id string) (string, error)) error {
	if os.Getenv("NO_CACHE") != "" {
		cacheKey = ""
	}

	b.config.Image = b.imageID

	resp, err := b.client.ContainerCreate(
		context.Background(),
		b.config,
		nil,
		nil,
		"",
	)
	if err != nil {
		return err
	}

	if hook != nil {
		tmp, err := hook(b, resp.ID)
		if err != nil {
			return err
		}

		if tmp != "" && os.Getenv("NO_CACHE") == "" {
			cacheKey = tmp
		}
	}

	commitResp, err := b.client.ContainerCommit(context.Background(), resp.ID, types.ContainerCommitOptions{Config: b.config, Comment: cacheKey})
	if err != nil {
		return fmt.Errorf("Error during commit: %v", err)
	}

	err = b.client.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return fmt.Errorf("Could not remove intermediate container %q: %v", resp.ID, err)
	}

	b.imageID = commitResp.ID

	return nil
}

func createException(m *mruby.Mrb, msg string) mruby.Value {
	val, err := m.Class("Exception", nil).New(mruby.String(msg))
	if err != nil {
		panic(fmt.Sprintf("could not construct exception for return: %v", err))
	}

	return val
}
