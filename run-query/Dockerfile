FROM golang:1.24.2 as build
WORKDIR /run-query
# Copy dependencies list
COPY go.mod go.sum ./
# Build with optional lambda.norpc tag
COPY main.go .
COPY text-to-sql.go .
COPY run-query.go .
COPY schema.json .
RUN go build -tags lambda.norpc -o main main.go text-to-sql.go run-query.go
# Copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2023
COPY --from=build /run-query/main ./main 
COPY --from=build /run-query/schema.json ./schema.json
ENTRYPOINT [ "./main" ]