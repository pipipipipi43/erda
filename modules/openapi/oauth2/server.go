// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package oauth2

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"

	"github.com/erda-project/erda/modules/openapi/oauth2/clientstore/mysqlclientstore"
	"github.com/erda-project/erda/modules/openapi/oauth2/tokenstore/mysqltokenstore"
)

type OAuth2Server struct {
	srv *server.Server
}

func (o *OAuth2Server) Server() *server.Server {
	return o.srv
}

func NewOAuth2Server() *OAuth2Server {
	manager := manage.NewDefaultManager()
	manager.SetClientTokenCfg(&manage.Config{
		AccessTokenExp:    time.Hour * 1,
		RefreshTokenExp:   0,
		IsGenerateRefresh: false,
	})

	manager.MustClientStorage(mysqlclientstore.NewClientStore())
	manager.MustTokenStorage(mysqltokenstore.NewTokenStore(mysqltokenstore.WithTokenStoreGCInterval(time.Second * 30)))

	// jwt token generate
	manager.MapAccessGenerate(NewJWTAccessGenerate([]byte(JWTKey), jwt.SigningMethodHS512))

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(false) // POST
	srv.SetClientInfoHandler(server.ClientFormHandler)

	// logger
	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		logrus.Errorf("oauth2 server internal err: %v", err)
		re = &errors.Response{Error: err}
		return
	})
	srv.SetResponseErrorHandler(func(re *errors.Response) {
		logrus.Errorf("oauth2 server response err: %v", re.Error.Error())
	})

	return &OAuth2Server{
		srv: srv,
	}
}

// {{openapi}}/oauth2/token?grant_type=client_credentials&client_id=pipeline&client_secret=devops/pipeline&scope=read
func (o *OAuth2Server) Token(w http.ResponseWriter, r *http.Request) {
	err := o.srv.HandleTokenRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// {{openapi}}/oauth2/invalidate_token?access_token=xxx
func (o *OAuth2Server) InvalidateToken(w http.ResponseWriter, r *http.Request) {
	ti, err := o.srv.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// remote from store
	err = o.srv.Manager.RemoveAccessToken(ti.GetAccess())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(o, w, ti)
}

// {{openapi}}/oauth2/validate_token?access_token=xxx
func (o *OAuth2Server) ValidateToken(w http.ResponseWriter, r *http.Request) {
	ti, err := o.srv.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(o, w, ti)
}

func writeJSON(o *OAuth2Server, w http.ResponseWriter, tokenInfo oauth2.TokenInfo) {
	tokenData := o.srv.GetTokenData(tokenInfo)
	expiresAt := tokenInfo.GetAccessCreateAt().Add(tokenInfo.GetAccessExpiresIn())
	tokenData["expiresAt"] = expiresAt.Format(time.RFC3339)
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(&tokenData)
	w.Write(b)
}
