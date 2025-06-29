package main

import (
	"board/internal/api"
	"board/internal/pagerenderer"
	"board/internal/repo"
	"log"
	"net/http"

	"github.com/ilyakaznacheev/cleanenv"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	r_cfg := repo.RepoConfig{}
	a_cfg := api.ApiConfig{}
	pr_cfg := pagerenderer.PagerendererConfig{}
	err := cleanenv.ReadConfig("config.env", &r_cfg)
	if err != nil {
		cleanenv.ReadEnv(r_cfg)
	}
	err = cleanenv.ReadConfig("config.env", &a_cfg)
	if err != nil {
		cleanenv.ReadEnv(r_cfg)
	}
	err = cleanenv.ReadConfig("config.env", &pr_cfg)
	if err != nil {
		cleanenv.ReadEnv(r_cfg)
	}
	pr, err := repo.NewPostsRepo(r_cfg)
	if err != nil {
		panic(err)
	}
	a := api.NewApi(pr, a_cfg)
	pager := pagerenderer.NewPageRenderer(pr, pr_cfg)
	go http.ListenAndServe(a.PORT, a.Router)

	if pager.HTTPS {
		/*go func() {
			if err := http.ListenAndServe(pager.PORT, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+pager.DOMAIN+":443"+r.RequestURI, http.StatusMovedPermanently)
			})); err != nil {
				log.Fatalf("ListenAndServe error: %v", err)
			}
		}()*/
		log.Fatal(http.Serve(autocert.NewListener(pager.DOMAIN, pager.EXTRA_DOMAIN), pager.Router))
	} else {
		http.ListenAndServe(pager.PORT, pager.Router)
	}
}
