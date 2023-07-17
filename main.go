package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bytebase/bytebase/backend/api/auth"
	"github.com/bytebase/bytebase/backend/common"
	api "github.com/bytebase/bytebase/backend/legacyapi"
	_ "github.com/bytebase/bytebase/backend/plugin/db/pg"
	"github.com/bytebase/bytebase/backend/store"
)

func createHandler() func(w http.ResponseWriter, r *http.Request) {
	mdDB := store.NewMetadataDBWithExternalPg(os.Getenv("BYTEBASE_UNAUTH_PG_URL"), "", "", common.ReleaseModeProd)
	db, err := mdDB.Connect(0, false, "")
	if err != nil {
		log.Fatalln("cannot connect to database", err)
	}
	if err := db.Open(context.Background()); err != nil {
		log.Fatalln("cannot connect to database", err)
	}
	storeInstance := store.New(db)

	s := api.SettingAuthSecret
	as, err := storeInstance.GetSettingV2(context.Background(), &store.FindSettingMessage{Name: &s})
	if err != nil {
		log.Fatalln("cannot get auth secret setting", err)
	}
	authSecret := as.Value

	cID := os.Getenv("BYTEBASE_UNAUTH_CREATOR_ID")
	if cID == "" {
		cID = "101"
	}
	creatorID, err := strconv.Atoi(cID)
	if err != nil {
		log.Fatalln("invalid creator id: " + strconv.Itoa(creatorID))
	}

	groupPrefix := os.Getenv("BYTEBASE_UNAUTH_GROUP_PREFIX")

	return func(w http.ResponseWriter, r *http.Request) {
		email := r.Header.Get("X-User-Email")
		username := r.Header.Get("X-User-Name")
		groups := r.Header.Get("X-User-Role")

		var role api.Role
		for _, g := range strings.Split(groups, ",") {
			if role != "" {
				break
			}

			g = strings.TrimSpace(g)
			if !strings.HasPrefix(g, groupPrefix) {
				continue
			}

			g = strings.TrimLeft(g, groupPrefix)
			role = api.Role(strings.ToUpper(g))
		}
		if role == "" {
			sendError(w, 403, fmt.Errorf("no groups are provided"))
			return
		}

		user, err := storeInstance.GetUser(r.Context(), &store.FindUserMessage{Email: &email, ShowDeleted: true})
		if err != nil {
			sendError(w, 500, fmt.Errorf("find user failed: %w", err))
			return
		}

		if user == nil {
			user, err = storeInstance.CreateUser(r.Context(), &store.UserMessage{
				Email: email,
				Name:  username,
				Type:  api.EndUser,
			}, creatorID)
			if err != nil {
				sendError(w, 500, fmt.Errorf("create user failed: %w", err))
				return
			}
		}

		user, err = storeInstance.UpdateUser(r.Context(),
			user.ID,
			&store.UpdateUserMessage{Role: &role},
			creatorID,
		)
		if err != nil {
			sendError(w, 500, fmt.Errorf("update user role failed: %w", err))
			return
		}

		sendLoginSuccess(w, r, user, authSecret)
	}
}

func sendLoginSuccess(w http.ResponseWriter, r *http.Request, user *store.UserMessage, authSecret string) {
	token, err := auth.GenerateAccessToken(user.Name, user.ID, common.ReleaseModeProd, authSecret)
	if err != nil {
		sendError(w, 500, fmt.Errorf("generate access token failed: %w", err))
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(user.Name, user.ID, common.ReleaseModeProd, authSecret)
	if err != nil {
		sendError(w, 500, fmt.Errorf("generate refresh token failed: %w", err))
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "user", Value: strconv.Itoa(user.ID)})
	http.SetCookie(w, &http.Cookie{Name: "access-token", Value: token})
	http.SetCookie(w, &http.Cookie{Name: "refresh-token", Value: refreshToken})
	http.Redirect(w, r, r.URL.Query().Get("redirect"), 302)
}

func sendError(w http.ResponseWriter, st int, err error) {
	w.WriteHeader(st)
	w.Write([]byte(err.Error()))
	log.Println(err)
}

func main() {
	http.HandleFunc("/", createHandler())

	addr := os.Getenv("BYTEBASE_UNAUTH_LISTEN_ADDRESS")
	if addr == "" {
		addr = ":8080"
	}
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalln("failed to start http server", err)
	}
}
