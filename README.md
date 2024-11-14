<h1>forum</h1>

## About
This forum is made for skufs. If you are one of us, be welcome to [skuf.life](http://skuf.life).

## Audit
[Audit list is here ->](https://github.com/01-edu/public/tree/master/subjects/forum/audit)

## Usage

The project is written with **Go version 1.23.1**. If you have older version of Go use **Docker** to run our forum.


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