// Package transport re-exports admin's gRPC handlers under a clean name
// for the bootstrap layer. Keeping the actual handler files in `grpc/`
// (where they sit alongside the generated stubs) without forcing
// bootstrap to import a package literally named `grpc`.
package transport

import grpcimpl "github.com/artem13815/hr/admin/internal/transport/grpc"

type AdminServiceAPI = grpcimpl.AdminServiceAPI

var NewAdminServiceAPI = grpcimpl.NewAdminServiceAPI
