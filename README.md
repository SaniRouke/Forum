<h1>forum</h1>

## About
## Features

## Usage

The project is written with **Go version 1.23.1**. If you have older version of Go use **Docker** to run our forum.

Skuf aller Länder, vereinigt Euch!
Rate your Skuf level from 8 to 10.
Are you skuf? 
Yes, I am | Definitely not. I just want to know who is it


**How to run Docker**

Build Docker image:
```bash
docker build -t forum-app .
```
Run Docker container:
```bash
docker run -p 8080:8080 forum-app
```

Now you can go to the [localhost:8080](http://localhost:8080) and check the program!

To run unit tests run this command:

```bash
go test ./... -v
```