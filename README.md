# Anubis POW Solver
![joke](./docs/not_so_happy.png)

Solves proof-of-work challenges served by [TecharoHQ/anubis](https://github.com/TecharoHQ/anubis) (https://anubis.techaro.lol/).

## Why?
Bots will always find a way to bypass WAFs while real users are forced to waste CPU cycles on a loading screen. This is a proof of concept to show how easy it is to bypass this type of antibot protection.

## Usage
This project comes with two different modes of operation. The former approach is recommended if you require quick solves, the latter is (supposed to be) more future-proof.

### Native solver
To solve the challenge almost instantly, without evaluating any of the JavaScript challenge code, pass in the `--native` flag:

```bash
$ go run main.go --native
```

This solves the challenge even faster than a real browser could!

### Dynamic solver
To solve the challenge in a more 'legit' manner, by evaluating the JavaScript challenge code, omit the flag:

```bash
$ go run main.go
```

## Goals
To make this a little more interesting for myself, I set some goals for this project:

- Use the [V8](https://v8.dev/) engine directly to execute the JavaScript code dynamically. No browser automation. No hardcoding.
- Make (almost) no changes to the original challenge JS bundle.
- Speed.
