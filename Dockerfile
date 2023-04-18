# Start from a Golang 1.20 image
FROM golang:1.20.3

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files to the container
COPY go.mod go.sum ./

# Install the dependencies
RUN go mod download

# Copy the rest of the application files to the container
COPY . .

# Build the application
RUN mkdir /mainbin
RUN go build -o /mainbin/main .

# Set the environment variables
ENV PORT=8080
ENV DATABASE_URL=postgres://user:password@postgres_db:5432/products_db?sslmode=disable

# Expose the port
EXPOSE $PORT

# Run the application
CMD ["/mainbin/main"]
