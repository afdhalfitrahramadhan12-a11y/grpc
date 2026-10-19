# gRPC Polyglot Lab

Komunikasi inter-service dengan gRPC: satu **server Go** dan dua **client** (**Python** dan **Node.js**) yang memakai satu kontrak `user.proto`.

| Komponen | Bahasa | Cara memakai `user.proto` |
|---|---|---|
| `go-server` | Go | Static code generation (`protoc`) |
| `python-client` | Python | Static code generation (`grpc_tools.protoc`) |
| `node-client` | Node.js | Dynamic runtime loading (`@grpc/proto-loader`) |

Server berjalan di `localhost:50051`.

## Struktur Proyek

```text
grpc-polyglot-lab/
├── .gitignore
├── user.proto
├── go-server/
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── pb/
│       ├── user.pb.go
│       └── user_grpc.pb.go
├── python-client/
│   ├── client.py
│   ├── user_pb2.py
│   └── user_pb2_grpc.py
└── node-client/
    ├── client.js
    ├── package.json
    └── package-lock.json
```

## 0. Prasyarat (Linux Ubuntu/Debian)

Lewati yang sudah terpasang.

```bash
sudo apt update
sudo apt install -y git python3 python3-pip python3-venv nodejs npm protobuf-compiler
```

Go dipasang dari paket resmi di <https://go.dev/dl/>, lalu buka terminal baru agar `PATH` terbarui.

Verifikasi:

```bash
git --version
python3 --version
node -v
npm -v
go version
protoc --version
```

Plugin `protoc` untuk Go dipasang di langkah 3. Pastikan `$(go env GOPATH)/bin` ada di `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

## 1. Setup Folder dan Git

```bash
mkdir grpc-polyglot-lab
cd grpc-polyglot-lab
git init

echo "node_modules/" >> .gitignore
echo "__pycache__/" >> .gitignore
echo ".venv/" >> .gitignore
echo "*.exe" >> .gitignore

git add .gitignore
git commit -m "Initial commit: Add gitignore rules"
```

## 2. Menulis Kontrak API (`user.proto`)

Buat `user.proto` di folder utama proyek. Berkas ini adalah *Single Source of Truth* bagi server dan semua client.

```protobuf
syntax = "proto3";

package user;

option go_package = "example.com/grpc-polyglot-lab/go-server/pb;pb";

message UserRequest {
    int32 user_id = 1;
}

message UserResponse {
    string name = 1;
    string email = 2;
    string phone = 3;
}

service UserService {
    rpc GetUserProfile(UserRequest) returns (UserResponse);
}
```

> Field `phone = 3` adalah tambahan dari Tugas Evaluasi Mandiri No. 1. Jika ingin mengikuti tahap awal modul, tulis `UserResponse` tanpa `phone` dahulu, lalu tambahkan di langkah 7.

## 3. Server Go

### 3.1 Inisialisasi modul dan dependensi

```bash
mkdir go-server
cd go-server
go mod init example.com/grpc-polyglot-lab/go-server

go get google.golang.org/grpc
go get google.golang.org/protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

> Nama modul pada `go mod init` harus sama dengan nilai `module=` pada perintah `protoc` dan jalur `import` di `main.go`. Modul praktikum memakai `go mod init go-server`; jika memakai nama itu, sesuaikan juga `go_package` dan opsi `module=`.

### 3.2 Generate kode Go

Jalankan dari dalam folder `go-server`:

```bash
protoc --proto_path=../ \
  --go_out=. \
  --go_opt=module=example.com/grpc-polyglot-lab/go-server \
  --go-grpc_out=. \
  --go-grpc_opt=module=example.com/grpc-polyglot-lab/go-server \
  ../user.proto
```

Hasil: folder `pb/` berisi `user.pb.go` dan `user_grpc.pb.go`.

### 3.3 Tulis `main.go`

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"example.com/grpc-polyglot-lab/go-server/pb"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedUserServiceServer
}

func (s *server) GetUserProfile(
	ctx context.Context,
	req *pb.UserRequest,
) (*pb.UserResponse, error) {
	log.Printf("Menerima request untuk User ID: %d", req.GetUserId())

	if req.GetUserId() == 1 {
		return &pb.UserResponse{
			Name:  "Afdhal Fitrah Ramadhan",
			Email: "afdhalfitrahramadhan2@gmail.com",
			Phone: "083163859104",
		}, nil
	}

	return &pb.UserResponse{
		Name:  "User Tidak Ditemukan",
		Email: "-",
		Phone: "-",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Gagal membuka port: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &server{})

	fmt.Println("gRPC Server (Go) berjalan di port :50051...")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
```

Bagian atas berkas (`package`, `import`, `struct server`) ditulis mengikuti modul dan disesuaikan dengan nama modul di atas.

## 4. Client Python

```bash
cd ..
mkdir python-client
cd python-client

python3 -m venv .venv
source .venv/bin/activate
pip install grpcio grpcio-tools

python -m grpc_tools.protoc --proto_path=../ --python_out=. --grpc_python_out=. ../user.proto
```

Hasil: `user_pb2.py` dan `user_pb2_grpc.py`.

Buat `client.py`:

```python
import grpc
import user_pb2
import user_pb2_grpc


def run():
    with grpc.insecure_channel("localhost:50051") as channel:
        stub = user_pb2_grpc.UserServiceStub(channel)

        print("Mengirim request ke Go Server menggunakan Python...")

        response = stub.GetUserProfile(
            user_pb2.UserRequest(user_id=1)
        )

        print(
            f"Respon dari Server -> Nama: {response.name}, "
            f"Email: {response.email}, "
            f"Telepon: {response.phone}"
        )


if __name__ == "__main__":
    run()
```

## 5. Client Node.js

```bash
cd ..
mkdir node-client
cd node-client
npm init -y
npm install @grpc/grpc-js @grpc/proto-loader
```

Buat `client.js`. Berkas `.proto` dibaca saat program berjalan, jadi tidak ada langkah `protoc`.

```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const path = require('path');

const PROTO_PATH = path.join(__dirname, '../user.proto');
const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true
});
const userProto = grpc.loadPackageDefinition(packageDefinition).user;

function main() {
    const client = new userProto.UserService('localhost:50051', grpc.credentials.createInsecure());
    console.log("Mengirim request ke Go Server menggunakan Node.js...");
    client.GetUserProfile({ user_id: 1 }, (error, response) => {
        if (!error) {
            console.log(`Respon dari Server -> Nama: ${response.name}, Email: ${response.email}, Telepon: ${response.phone}`);
        } else {
            console.error("Terjadi error:", error);
        }
    });
}
main();
```

> Baris `Telepon: ${response.phone}` adalah penyesuaian untuk field `phone`; kode modul aslinya hanya mencetak nama dan email.

## 6. Pengujian

Buka **3 terminal** dari folder utama proyek.

**Terminal 1 — Server Go**

```bash
cd go-server
go run main.go
```

**Terminal 2 — Client Python**

```bash
cd python-client
source .venv/bin/activate
python client.py
```

**Terminal 3 — Client Node.js**

```bash
cd node-client
node client.js
```

Program berhasil jika:

1. Tidak ada error crash (misalnya `Connection refused` atau `Cannot find module`) di ketiga terminal.
2. Data profil tercetak di terminal Python maupun Node.js, padahal data hanya ditulis di `main.go`.
3. Terminal server menampilkan log `Menerima request untuk User ID: 1` setiap ada client yang menjalankan kode.

Contoh keluaran client (dengan data pada `main.go` di atas):

```text
Mengirim request ke Go Server menggunakan Python...
Respon dari Server -> Nama: Afdhal Fitrah Ramadhan, Email: afdhalfitrahramadhan2@gmail.com, Telepon: 083163859104
```

## 7. Menambah Field `phone` (Tugas Evaluasi Mandiri No. 1)

1. Tambahkan `string phone = 3;` pada `UserResponse` di `user.proto`.
2. Generate ulang kode **Go** (perintah di langkah 3.2) dan **Python** (perintah di langkah 4).
3. Isi `Phone` pada kedua respons di `go-server/main.go`.
4. Cetak `response.phone` di `client.py` dan `client.js`.
5. Jalankan ulang server (`Ctrl+C`, lalu `go run main.go`) dan kedua client. Field berhasil jika baris `Telepon: ...` muncul di kedua client.



```bash
git add README.md
git commit -m "Docs: Add README with static vs dynamic code generation analysis"
git push
```
