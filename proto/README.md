# Proto Contracts

Step 5 adds protobuf/gRPC contracts under `proto/ft12/v1`.

Source files:

- `common.proto`
- `emulator.proto`
- `gateway.proto`

Generated Go code is committed under:

```text
internal/api/grpc/ft12/v1
```

Generate:

```powershell
make proto
```

Required tools:

```powershell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

`protoc` must also be available on `PATH`.
