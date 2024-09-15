package main

import "fmt"

type Traxer struct {
	DB string
}

type Inserter interface {
	Insert()
}

func (t Traxer) Insert() {
	t.Penitration()
}

func (t Traxer) Penitration() {
	fmt.Println("in and out... repeat")
}

func main() {
	var savva Inserter
	savva = Traxer{
		DB: "chix list",
	}

	savva.Insert()
	fmt.Println(savva.DB)
	savva.DB = "muchachos list"
	fmt.Println(savva.DB)
	savva.Penitrator()
}
