wget -O go1.24.0.linux-amd64.tar.gz https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
docker run --rm -v /share/Download/photobk:/app -w /app quay.io/pypa/manylinux2014_x86_64 bash -c '
  set -e
  tar -C /usr/local -xzf /app/go1.24.0.linux-amd64.tar.gz
  export PATH=/usr/local/go/bin:$PATH
  export GOPROXY=https://goproxy.cn,direct
  export GOSUMDB=sum.golang.google.cn

  # 编译 server
  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -o photo-backup-server cmd/server/main.go

  # 编译 CLI
  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -o photo-backup-cli cmd/cli/*
'