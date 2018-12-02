# pidstat
This is *my* re-imagination of what a modern `pidstat` would look like.

At its core, it will graph CPU and memory usage over time and display it via the
built-in web server.

## Install
If you have Go:
```
$ go get github.com/dselans/pidstat
$ pidstat help 
```

## Usage
```
# To start pidstat in web mode
$ pidstat serve [-l LISTEN_ADDRESS]

# To start pidstat in sexy console mode
$ pidstat cli
```

## Features
* Boojee web mode
* Sexy console mode
* Pretty graphs
* Rich reporting (JSON, CSV, HTML)

## Motivation
I am a backend developer and the last time I did "frontend" dev, I used bootstrap,
jQuery and flask for serving up templates (which was 4+ years ago?).

This project has two purposes:

1. Allows me to try out Golang libs I haven't had a chance to use in my day job
2. Practice frontend dev

This project uses:

* For the router: [go-chi/chi](https://github.com/go-chi/chi)
    * My usual go to: [gorilla/mux](https://github.com/gorilla/mux)
* For logging: [uber-go/zap](https://github.com/uber-go/zap)
    * My usual go to: [sirupsen/logrus](https://github.com/sirupsen/logrus)
* For CLI flag parsing: [urfave/cli](https://github.com/urfave/cli)
    * My usual go to: [alecthomas/kingpin](https://github.com/alecthomas/kingpin)
* For vendoring: [vgo](https://github.com/golang/go/wiki/vgo)
    * My usual go to: [kardianos/govendor](https://github.com/kardianos/govendor)
* For UI: [vuejs](https://github.com/vuejs)
    * My usual go to: [w3schools](https://www.w3schools.com/)

## Personal Notes
* Should we vendor dependencies?
    * Looks like go modules, by default, do not utilize vendoring. Feels weird.
    * [Vendoring doc](https://github.com/golang/go/wiki/Modules#how-do-i-use-vendoring-with-modules-is-vendoring-going-away)
    * _A: Won't vendor until/unless it becomes a hassle._

## Contribute
See something dumb? Let's fix it - open a PR and let's discuss it! I am usually
pretty quick with PR's but if you're not seeing any traction, message '' on
[Gopher Slack](https://invite.slack.golangbridge.org/).