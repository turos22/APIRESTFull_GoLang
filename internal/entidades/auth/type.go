package auth

type CreateUserParams struct {
	id       int64  `json: id`
	email    string `json: email`
	password string `json: password`
	name     string `json: name`
	role     string `json: role`
}

type LoginUserParams struct{
	id       int64  `json: id`
	email    string `json: email`
	password string `json: password`
}
