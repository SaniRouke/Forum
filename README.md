<h1>forum-moderation</h1>

## About
This forum is made for skufs.

## Audit
[Audit list is here ->](https://github.com/01-edu/public/tree/master/subjects/forum/audit)

## Usage

```bash
go run ./cmd/web
```

The project is written with **Go version 1.23.1**. If you have older version of Go use **Docker** to run our forum.


**How to run Docker**

Build Docker image:
```bash
docker build -t forum-app .
```
Run Docker container:
```bash
docker run -p 8443:8443 forum-app
```

Now you can go to the [localhost:8443](http://localhost:8443) and check the program!

## Team
**Rustam** [@srouke](https://01.alem.school/git/srouke)

**Konstantin** [@kbaraban](https://01.alem.school/git/kbaraban)