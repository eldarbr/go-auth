FROM golang:1.23.2-alpine3.19 AS build
COPY . /app
RUN apk add --no-cache make && make -C /app build

FROM scratch AS run
COPY --from=build /app/bin/go-auth /app/go-auth
WORKDIR /app
USER 1000
CMD [ "./go-auth" ]
