<h1>forum</h1>

## About
This forum is made for skufs. If you are one of us, welcome to [skuf.life](http://skuf.life)
The project was written on [GitHub](https://github.com/SaniRouke/Forum)

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
docker run -p 8080:8080 forum-app
```

Now you can go to the [localhost:8080](http://localhost:8080) and check the program!

## Team
**Rustam** [@srouke](https://01.alem.school/git/srouke)

**Konstantin** [@kbaraban](https://01.alem.school/git/kbaraban)

**Azamat** [@aurazimb](https://01.alem.school/git/aurazimb)