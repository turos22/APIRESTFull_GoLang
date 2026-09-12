docker compose down
docker compose up -d
start "worker" cmd /k "cd cmd\worker && go run main.go"
cd cmd
go run main.go api.go
cd ..