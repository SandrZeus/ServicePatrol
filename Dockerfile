FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y libsqlite3-0 && rm -rf /var/lib/apt/lists/*
COPY servicepatrol .
EXPOSE 8080
CMD ["./servicepatrol"]
