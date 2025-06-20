package main

import (
	"board/internal/api"
	"board/internal/pagerenderer"
	"board/internal/postsrepo"
	"net/http"

	"github.com/ilyakaznacheev/cleanenv"
)

func main() {
	pr_cfg := postsrepo.PostsRepoConfig{}
	cleanenv.ReadConfig("config.env", &pr_cfg)
	pr, err := postsrepo.NewPostsRepo(pr_cfg)
	if err != nil {
		panic(err)
	}
	a := api.NewApi(pr)
	pager := pagerenderer.NewPageRenderer(pr)
	go http.ListenAndServe(":8080", a.Router)
	http.ListenAndServe(":8081", pager.Router)
}
