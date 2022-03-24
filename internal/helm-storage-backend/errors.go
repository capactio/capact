package helmstoragebackend

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func gRPCInternalError(err error) error {
	if err == nil {
		return nil
	}
	return status.Error(codes.Internal, err.Error())
}
