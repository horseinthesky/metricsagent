package gapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestPingDB(t *testing.T) {
	ctx := context.Background()

	client, closer := runTestServer(ctx, "")
	defer closer()

	_, err := client.PingDB(ctx, &emptypb.Empty{})
	require.NoError(t, err)
}
