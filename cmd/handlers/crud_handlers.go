package hdl

import (
	"forum/cmd/utils"
	db "forum/database"
	"log"
	"net/http"
)

func HandlerHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		utils.ErrorPage(w, http.StatusNotFound, "Page not found")
		return
	}

	if r.Method != "GET" {
		utils.ErrorPage(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	allPosts, err := db.GetAllPosts()
	if err != nil {
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal server error")
		log.Println(err)
		return
	}

	data := struct {
		Posts []db.Post
	}{
		Posts: allPosts,
	}

	err = utils.RenderTemplate(w, "home.html", data, http.StatusOK)
	if err != nil {
		log.Println(err)
	}
}

func HandlerPost(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.ErrorPage(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	post, err := db.GetPost(id)
	if err != nil {
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
		log.Println(err)
		return
	}

	if post.ID == 0 {
		utils.ErrorPage(w, http.StatusNotFound, "Post not found")
		return
	}

	err = utils.RenderTemplate(w, "post.html", post, http.StatusOK)
	if err != nil {
		log.Println(err)
	}
}

func HandlerCreatePost(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet:
		err := utils.RenderTemplate(w, "create.html", nil, http.StatusOK)
		if err != nil {
			log.Println(err)
		}
	case r.Method == http.MethodPost:
		topic := r.FormValue("topic")
		body := r.FormValue("body")
		err := db.CreatePost(topic, body)
		if err != nil {
			http.Error(w, "Unable to create post", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func HandlerDeletePost(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.ErrorPage(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	err := db.DeletePost(id)
	if err != nil {
		utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
		log.Println(err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandlerEditPost(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.ErrorPage(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		post, err := db.GetPost(id)
		if err != nil {
			utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
			log.Println(err)
			return
		}

		if post.ID == 0 {
			utils.ErrorPage(w, http.StatusNotFound, "Post not found")
			return
		}

		err = utils.RenderTemplate(w, "edit.html", post, http.StatusOK)
		if err != nil {
			log.Println(err)
		}

	case http.MethodPost:
		topic := r.FormValue("topic")
		body := r.FormValue("body")

		err := db.EditPost(id, topic, body)
		if err != nil {
			utils.ErrorPage(w, http.StatusInternalServerError, "Internal Server Error")
			log.Println(err)
			return
		}

		http.Redirect(w, r, "/post?id="+id, http.StatusSeeOther)
	}
}
