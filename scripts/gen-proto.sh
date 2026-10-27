#!/bin/bash
set -e

# Ensure tools are in PATH
export PATH="/opt/homebrew/bin:$(go env GOPATH)/bin:$PATH"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "Generating gRPC code from Protocol Buffers..."

PROTOS=(
  "proto/auth/v1/auth.proto"
  "proto/product/v1/product.proto"
  "proto/cart/v1/cart.proto"
  "proto/order/v1/order.proto"
  "proto/payment/v1/payment.proto"
)

for proto in "${PROTOS[@]}"; do
  echo "  -> Compiling $proto"
  protoc \
    --proto_path=. \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    "$proto"
done

echo "Protobuf and gRPC generation complete."
