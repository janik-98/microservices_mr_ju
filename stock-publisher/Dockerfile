# GoLang Image
FROM golang:1.20-alpine

# Verzeichnis setzen
WORKDIR /app

COPY . .

# Build Application/Dependencies
RUN go mod tidy && go build -o stock-publisher

# Port-Mapping
EXPOSE 8080

# Start
CMD ["./stock-publisher"]