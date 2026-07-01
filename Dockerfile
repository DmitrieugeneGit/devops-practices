FROM golang:1.26

WORKDIR /app

COPY backend/ ./backend/
COPY frontend/ ./frontend/

RUN cd backend && go build -o /app/tasks-app .

ENV FRONTEND_DIR="/app/frontend"

EXPOSE 8080

CMD ["./tasks-app"]
