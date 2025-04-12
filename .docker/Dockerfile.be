FROM golang:1.23.2-alpine

ENV GO111MODULE=on

# 必要パッケージのインストール
RUN apk update && apk --no-cache add git bash
RUN go install github.com/air-verse/air@latest
# 作業ディレクトリ作成
WORKDIR /app

# go.mod と go.sum を先にコピー（キャッシュ効率）
COPY ./be/go.mod ./

# 依存取得
RUN go mod download
RUN go mod tidy

# アプリの全コードをコピー
COPY ./be .

ENV PATH="/go/bin:${PATH}"

# アプリ起動
# CMD ["go", "run", "server.go"]

CMD ["air","-c",".air.toml"]