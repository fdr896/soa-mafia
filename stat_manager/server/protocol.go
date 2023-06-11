package server

import "mime/multipart"

/////////// Request ///////////

type PlayerUsername struct {
	Username string `uri:"username" binding:"required"`
}

type Player struct {
	Username string `form:"username" binding:"required"`
	Email string    `form:"email" binding:"required,email"`
	Gender string   `form:"gender" binding:"required,oneof=male female"`

	Avatar *multipart.FileHeader `form:"avatar"`
}

type PlayerForUpdate struct {
	Email  *string    `form:"email"`
	Gender *string    `form:"gender"`

	Avatar *multipart.FileHeader `form:"avatar"`
}

/////////// Response /////////// 

type PlayerInfo struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Gender   string `json:"gender"`
}

type PlayerInfos struct {
	Players []*PlayerInfo `json:"players"`
}
