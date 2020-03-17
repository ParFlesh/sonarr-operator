package image_inspect

import (
	"context"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
)

type ImageInspectorInterface interface {
	GetImageLabels(ctx context.Context, builderImageArg string) (*types.ImageInspectInfo, error)
}

type ImageInspector struct{}

// getImageLabels attempts to inspect an image existing in a remote registry.
func (r *ImageInspector) GetImageLabels(ctx context.Context, builderImageArg string) (*types.ImageInspectInfo, error) {
	ref, err := alltransports.ParseImageName(builderImageArg)
	if err != nil {
		return &types.ImageInspectInfo{}, err
	}

	img, err := ref.NewImage(ctx, &types.SystemContext{})
	if err != nil {
		return nil, err
	}

	imageMetadata, err := img.Inspect(ctx)
	if err != nil {
		return nil, err
	}

	return imageMetadata, nil
}

type MockImageInspector struct {
	GetImageLabelsOutput *types.ImageInspectInfo
	GetImageLabelsError  error
}

func (r *MockImageInspector) GetImageLabels(_ context.Context, _ string) (*types.ImageInspectInfo, error) {
	return r.GetImageLabelsOutput, r.GetImageLabelsError
}
