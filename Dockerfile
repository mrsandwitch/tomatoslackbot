FROM golang:1.16-alpine

WORKDIR /tomatoClock
COPY . /tomatoClock
RUN apk update && apk upgrade
RUN apk add --update nodejs-current npm
RUN apk add g++

WORKDIR /tomatoClock/web_app
RUN sed -i 's/"http:\/\/localhost:8000"/""/' src/components/Clock.vue
RUN npm install -g @vue/cli
RUN npm install
RUN npm run build
RUN cp -r dist/* /tomatoClock/dist/

WORKDIR /tomatoClock
RUN go build -o /tomatoClockServer

WORKDIR /
EXPOSE 8000
CMD ["/tomatoClockServer"]


