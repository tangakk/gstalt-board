package main

import (
	"board/internal/api"
	"board/internal/pagerenderer"
	"board/internal/repo"
	"net/http"

	"github.com/ilyakaznacheev/cleanenv"
)

func main() {
	pr_cfg := repo.RepoConfig{}
	err := cleanenv.ReadConfig("config.env", &pr_cfg)
	if err != nil {
		cleanenv.ReadEnv(pr_cfg)
	}
	pr, err := repo.NewPostsRepo(pr_cfg)
	if err != nil {
		panic(err)
	}
	a := api.NewApi(pr)
	pager := pagerenderer.NewPageRenderer(pr)
	go http.ListenAndServe(":8080", a.Router)
	http.ListenAndServe(":8081", pager.Router)
}
