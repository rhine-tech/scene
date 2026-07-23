package sessiontracker

import (
	"context"
	"testing"

	"github.com/rhine-tech/scene/lens/storage"
	"github.com/stretchr/testify/require"
)

func TestMemoryUploadSessionTrackerHonorsCanceledContext(t *testing.T) {
	tracker := NewMemoryUploadSessionTracker()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.ErrorIs(t, tracker.Save(ctx, "upload", storage.UploadSession{}), context.Canceled)
	_, err := tracker.Get(ctx, "upload")
	require.ErrorIs(t, err, context.Canceled)
	require.ErrorIs(t, tracker.Delete(ctx, "upload"), context.Canceled)

	_, err = tracker.Get(context.Background(), "upload")
	require.ErrorIs(t, err, storage.ErrUploadSessionNotFound)
}
